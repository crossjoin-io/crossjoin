package main

// Copyright 2021 Preetam Jinka
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/crossjoin-io/crossjoin/config"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	configFilePath := flag.String("config", "", "Path to config file")
	flag.Parse()

	if *configFilePath == "" {
		log.Fatalf("missing path to config file (--config)")
	}
	// Parse config file
	log.Println("using config file path", *configFilePath)
	conf := &config.Config{}
	configFile, err := os.ReadFile(*configFilePath)
	if err != nil {
		log.Fatalf("read config file: %v", err)
	}
	err = conf.Parse(configFile)
	if err != nil {
		log.Fatalf("parse config file: %v", err)
	}

	log.Println("starting crossjoin")
	err = run(conf)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("finished crossjoin")
}

func run(conf *config.Config) error {
	for _, dataset := range conf.DataSets {
		log.Printf("creating data set `%s`", dataset.Name)
		err := createDataset(dataset)
		if err != nil {
			return err
		}
	}
	return nil
}

func createDataset(dataset config.DataSet) error {
	filename := "./" + dataset.Name + ".db"

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
	err = fetchSingle(db, dataset.DataSource.Type, dataset.DataSource.ConnectionString, dataset.DataSource.Name, dataset.DataSource.Query)
	if err != nil {
		return err
	}

	for _, join := range dataset.Joins {
		log.Printf("querying `%s`", join.DataSource.Name)
		err = fetchSingle(db, join.DataSource.Type, join.DataSource.ConnectionString, join.DataSource.Name, join.DataSource.Query)
		if err != nil {
			return err
		}
	}

	joinClauses := ""
	for _, join := range dataset.Joins {
		joinColumns := []string{}
		for _, cols := range join.Columns {
			joinColumns = append(joinColumns, fmt.Sprintf("%s.%s = %s.%s", dataset.DataSource.Name, cols.LeftColumn, join.DataSource.Name, cols.RightColumn))
		}
		joinClauses += fmt.Sprintf(" %s %s ON %s", join.Type, join.DataSource.Name, strings.Join(joinColumns, " AND "))
	}

	log.Println("joining data")
	_, err = db.Exec(fmt.Sprintf("CREATE TABLE %s AS SELECT * FROM %s %s", dataset.Name, dataset.DataSource.Name, joinClauses))
	return err
}

func fetchSingle(dest *sql.DB, connType string, connStr string, name string, query string) error {
	db, err := sql.Open(connType, connStr)
	if err != nil {
		return err
	}
	defer db.Close()
	log.Println("query")
	rows, err := db.Query(query)
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

	_, err = dest.Exec(fmt.Sprintf("CREATE TABLE %s (%s)", name, strings.Join(columns, ",")))
	if err != nil {
		return err
	}

	params := []string{}
	for i := range columns {
		params = append(params, fmt.Sprintf("$%d", i+1))
	}
	stmt, err := dest.Prepare(fmt.Sprintf("INSERT INTO %s VALUES (%s)", name, strings.Join(params, ",")))
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

		stmt.Exec(values...)
		if err != nil {
			return err
		}
	}

	return nil
}
