package api

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/crossjoin-io/crossjoin/config"
	_ "github.com/lib/pq"
	"gopkg.in/yaml.v2"
)

func (api *API) StoreDataset(hash string, dataset config.Dataset) error {
	marshaledDatset, err := yaml.Marshal(dataset)
	if err != nil {
		return err
	}
	_, err = api.db.Exec("REPLACE INTO datasets (config_hash, id, text) VALUES ($1, $2, $3)",
		hash, dataset.ID, marshaledDatset,
	)
	if err != nil {
		return err
	}
	return err
}

func (api *API) ReadDatasets() ([]config.Dataset, error) {
	hash, err := api.LatestConfigHash()
	if err != nil {
		return nil, fmt.Errorf("read latest config hash: %w", err)
	}

	rows, err := api.db.Query("SELECT id, text FROM datasets WHERE config_hash = $1", hash)
	if err != nil {
		return nil, fmt.Errorf("query datasets: %w", err)
	}
	defer rows.Close()

	datasets := []config.Dataset{}
	for rows.Next() {
		id := ""
		text := ""
		dataset := config.Dataset{}
		err = rows.Scan(&id, &text)
		if err != nil {
			return nil, fmt.Errorf("scan dataset: %w", err)
		}
		err = yaml.Unmarshal([]byte(text), &dataset)
		if err != nil {
			return nil, fmt.Errorf("unmarshal datasets: %w", err)
		}
		dataset.ID = id
		datasets = append(datasets, dataset)
	}
	return datasets, nil
}

func (api *API) PreviewDataset(id string) ([]interface{}, error) {
	filename := filepath.Join(api.dataDir, id+".db")
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM " + id + " LIMIT 25")
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

func (api *API) refreshDataset(hash string, id string) error {
	text := ""
	err := api.db.QueryRow("SELECT text FROM datasets WHERE config_hash = $1 AND id = $2", hash, id).Scan(&text)
	if err != nil {
		return err
	}
	dataset := config.Dataset{}
	err = yaml.Unmarshal([]byte(text), &dataset)
	if err != nil {
		return err
	}
	err = api.createDataset(hash, dataset)
	if err != nil {
		return fmt.Errorf("create dataset: %w", err)
	}
	workflows, err := api.GetWorkflows(hash)
	if err != nil {
		return fmt.Errorf("get workflows: %w", err)
	}

	for _, workflow := range workflows {
		if workflow.On == nil {
			continue
		}
		for _, datasetID := range workflow.On.DatasetRefresh {
			if datasetID == id {
				err = api.StartWorkflow(hash, workflow.ID, nil)
				if err != nil {
					return fmt.Errorf("start workflow: %w", err)
				}
			}
		}
	}
	return nil
}

func (api *API) createDataset(hash string, dataset config.Dataset) error {
	filename := filepath.Join(api.dataDir, dataset.ID+".db")

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

	log.Printf("querying `%s`", dataset.DataSource.ID)
	err = api.fetchSingle(hash, db, dataset.DataSource)
	if err != nil {
		return fmt.Errorf("fetch single: %w", err)
	}

	for _, join := range dataset.Joins {
		log.Printf("querying `%s`", join.DataSource.ID)
		err = api.fetchSingle(hash, db, join.DataSource)
		if err != nil {
			return fmt.Errorf("fetch single as part of join: %w", err)
		}
	}

	joinClauses := ""
	for _, join := range dataset.Joins {
		joinColumns := []string{}
		for _, cols := range join.Columns {
			joinColumns = append(joinColumns, fmt.Sprintf(`%s."%s" = %s."%s"`, dataset.DataSource.ID, cols.LeftColumn, join.DataSource.ID, cols.RightColumn))
		}
		joinClauses += fmt.Sprintf(" %s %s ON %s", join.Type, join.DataSource.ID, strings.Join(joinColumns, " AND "))
	}

	log.Println("joining data")
	joinQuery := fmt.Sprintf("CREATE TABLE %s AS SELECT * FROM %s %s", dataset.ID, dataset.DataSource.ID, joinClauses)
	_, err = db.Exec(joinQuery)
	return err
}

func (api *API) fetchSingle(hash string, dest *sql.DB, dataSource *config.DataSource) error {
	dataConnection, err := api.ReadDataConnection(hash, dataSource.DataConnection)
	if err != nil {
		return err
	}
	switch dataConnection.Type {
	case "csv":
		f, err := api.readFile(dataConnection.Path)
		if err != nil {
			return err
		}
		r := csv.NewReader(f)
		firstLine, err := r.Read()
		if err != nil {
			return err
		}
		columns := firstLine
		for i := range columns {
			columns[i] = strconv.Quote(columns[i])
		}

		_, err = dest.Exec(fmt.Sprintf("CREATE TABLE %s (%s)", dataSource.ID, strings.Join(columns, ",")))
		if err != nil {
			return err
		}

		params := []string{}
		for i := range columns {
			params = append(params, fmt.Sprintf("$%d", i+1))
		}
		stmt, err := dest.Prepare(fmt.Sprintf("INSERT INTO %s VALUES (%s)", dataSource.ID, strings.Join(params, ",")))
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

		_, err = dest.Exec(fmt.Sprintf("CREATE TABLE %s (%s)", dataSource.ID, strings.Join(columns, ",")))
		if err != nil {
			return err
		}

		params := []string{}
		for i := range columns {
			params = append(params, fmt.Sprintf("$%d", i+1))
		}
		stmt, err := dest.Prepare(fmt.Sprintf("INSERT INTO %s VALUES (%s)", dataSource.ID, strings.Join(params, ",")))
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

func (api *API) readFile(path string) (io.Reader, error) {
	log.Printf("reading file `%s`", path)
	urlPath, _ := url.Parse(path)
	if urlPath != nil {
		if strings.Contains(path, "api.github.com") {
			contents, err := api.fetchGitHubFile(path)
			if err != nil {
				return nil, fmt.Errorf("read file: %w", err)
			}
			return bytes.NewReader(contents), nil
		}
		return nil, fmt.Errorf("unsupported path: %s", path)
	}
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	return bytes.NewReader(contents), nil
}
