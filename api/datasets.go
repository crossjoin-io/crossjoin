package api

import (
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/crossjoin-io/crossjoin/config"
	"gopkg.in/yaml.v2"
)

func (api *API) StoreDataset(dataset config.Dataset) error {
	marshaledDatset, err := yaml.Marshal(dataset)
	if err != nil {
		return err
	}
	_, err = api.db.Exec("REPLACE INTO datasets (name, text) VALUES ($1, $2)",
		dataset.Name, marshaledDatset,
	)
	if err != nil {
		return err
	}
	if dataset.Refresh != nil {
		dur, err := time.ParseDuration(dataset.Refresh.Interval)
		if err != nil {
			return fmt.Errorf("parse refresh interval: %w", err)
		}
		ticker := time.NewTicker(dur)
		go func() {
			for range ticker.C {
				log.Println("refreshing", dataset.Name)
				api.refreshDataset(dataset.Name)
			}
		}()
	}
	err = api.refreshDataset(dataset.Name)
	return err
}

func (api *API) PreviewDataset(name string) ([]interface{}, error) {
	filename := filepath.Join(api.dataDir, name+".db")
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM " + name + " LIMIT 50")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	result := []interface{}{}
	for rows.Next() {
		row := map[string]interface{}{}
		values := make([]interface{}, len(columns))
		valPointers := make([]interface{}, len(values))
		for i := range values {
			valPointers[i] = &values[i]
		}
		err = rows.Scan(valPointers...)
		if err != nil {
			return nil, err
		}
		for i, col := range columns {
			row[col] = values[i]
		}
		result = append(result, row)
	}
	return result, nil
}

func (api *API) refreshDataset(name string) error {
	text := ""
	err := api.db.QueryRow("SELECT text FROM datasets WHERE name = $1", name).Scan(&text)
	if err != nil {
		return err
	}
	dataset := config.Dataset{}
	err = yaml.Unmarshal([]byte(text), &dataset)
	if err != nil {
		return err
	}
	err = api.createDataset(dataset)
	if err != nil {
		return fmt.Errorf("create dataset: %w", err)
	}
	workflows, err := api.GetWorkflows()
	if err != nil {
		return fmt.Errorf("get workflows: %w", err)
	}

	for _, workflow := range workflows {
		if workflow.On == nil {
			continue
		}
		for _, datasetID := range workflow.On.DatasetRefresh {
			if datasetID == name {
				err = api.StartWorkflow(workflow.ID, nil)
				if err != nil {
					return fmt.Errorf("start workflow: %w", err)
				}
			}
		}
	}
	return nil
}

func (api *API) createDataset(dataset config.Dataset) error {
	filename := filepath.Join(api.dataDir, dataset.Name+".db")

	// Does the file exist? If so, remove it.
	_, err := os.Stat(filename)
	if err == nil {
		log.Printf("`%s` already exists; removing", filename)
		os.Remove(filename)
	}

	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec("PRAGMA synchronous = OFF")
	if err != nil {
		return err
	}
	_, err = db.Exec("PRAGMA journal_mode = MEMORY")
	if err != nil {
		return err
	}
	_, err = db.Exec("PRAGMA cache_size = -2000000")
	if err != nil {
		return err
	}

	log.Printf("querying `%s`", dataset.DataSource.Name)
	err = api.fetchSingle(db, dataset.DataSource)
	if err != nil {
		return err
	}

	for _, join := range dataset.Joins {
		log.Printf("querying `%s`", join.DataSource.Name)
		err = api.fetchSingle(db, join.DataSource)
		if err != nil {
			return err
		}
	}

	joinClauses := ""
	for _, join := range dataset.Joins {
		joinColumns := []string{}
		for _, cols := range join.Columns {
			joinColumns = append(joinColumns, fmt.Sprintf(`%s."%s" = %s."%s"`, dataset.DataSource.Name, cols.LeftColumn, join.DataSource.Name, cols.RightColumn))
		}
		joinClauses += fmt.Sprintf(" %s %s ON %s", join.Type, join.DataSource.Name, strings.Join(joinColumns, " AND "))
	}

	log.Println("joining data")
	joinQuery := fmt.Sprintf("CREATE TABLE %s AS SELECT * FROM %s %s", dataset.Name, dataset.DataSource.Name, joinClauses)
	_, err = db.Exec(joinQuery)
	return err
}

func (api *API) fetchSingle(dest *sql.DB, dataSource *config.DataSource) error {
	dataConnection, err := api.ReadDataConnection(dataSource.DataConnection)
	if err != nil {
		return err
	}
	switch dataConnection.Type {
	case "csv":
		f, err := os.Open(dataConnection.Path)
		if err != nil {
			return err
		}
		defer f.Close()
		r := csv.NewReader(f)
		firstLine, err := r.Read()
		if err != nil {
			return err
		}
		columns := firstLine
		for i := range columns {
			columns[i] = strconv.Quote(columns[i])
		}

		_, err = dest.Exec(fmt.Sprintf("CREATE TABLE %s (%s)", dataSource.Name, strings.Join(columns, ",")))
		if err != nil {
			return err
		}

		params := []string{}
		for i := range columns {
			params = append(params, fmt.Sprintf("$%d", i+1))
		}
		stmt, err := dest.Prepare(fmt.Sprintf("INSERT INTO %s VALUES (%s)", dataSource.Name, strings.Join(params, ",")))
		if err != nil {
			return err
		}
		defer stmt.Close()

		for {
			record, err := r.Read()
			if err != nil {
				if err == io.EOF {
					return nil
				}
				return err
			}
			if len(record) != len(columns) {
				return errors.New("inconsistent number of fields")
			}
			values := make([]interface{}, len(record))
			for i := range record {
				values[i] = record[i]
			}
			_, err = stmt.Exec(values...)
			if err != nil {
				return err
			}
		}
	case "postgres":
		db, err := sql.Open(dataConnection.Type, dataConnection.ConnectionString)
		if err != nil {
			return err
		}
		defer db.Close()

		rows, err := db.Query(dataSource.Query)
		if err != nil {
			return err
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			return err
		}

		for i := range columns {
			columns[i] = `"` + columns[i] + `"`
		}

		_, err = dest.Exec(fmt.Sprintf("CREATE TABLE %s (%s)", dataSource.Name, strings.Join(columns, ",")))
		if err != nil {
			return err
		}

		params := []string{}
		for i := range columns {
			params = append(params, fmt.Sprintf("$%d", i+1))
		}
		stmt, err := dest.Prepare(fmt.Sprintf("INSERT INTO %s VALUES (%s)", dataSource.Name, strings.Join(params, ",")))
		if err != nil {
			return err
		}
		defer stmt.Close()

		for rows.Next() {
			cols := make([]interface{}, len(columns))
			colPointers := make([]interface{}, len(cols))
			for i := range cols {
				colPointers[i] = &cols[i]
			}

			if err := rows.Scan(colPointers...); err != nil {
				return err
			}

			values := []interface{}{}
			for i := range columns {
				val := colPointers[i].(*interface{})
				values = append(values, *val)
			}

			_, err = stmt.Exec(values...)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
