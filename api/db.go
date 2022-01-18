package api

import (
	"database/sql"
	"log"
)

// setupDatabase sets up the database and schema migrations.
func setupDatabase(db *sql.DB) error {
	migrations := []string{
		/* 000 */ `CREATE TABLE IF NOT EXISTS schema_version (version INT PRIMARY KEY, timestamp TIMESTAMP)`,
		/* 001 */ `
		CREATE TABLE IF NOT EXISTS workflows (id TEXT NOT NULL PRIMARY KEY, text TEXT NOT NULL);
		CREATE TABLE IF NOT EXISTS workflow_runs (id TEXT NOT NULL PRIMARY KEY, workflow_id TEXT NOT NULL, started_at TIMESTAMP, completed_at TIMESTAMP, success BOOL);
		CREATE TABLE tasks (
			id TEXT NOT NULL PRIMARY KEY,
			workflow_run_id TEXT NOT NULL,
			workflow_task_id TEXT NOT NULL,
			input JSON NOT NULL DEFAULT '{}',
			output JSON NOT NULL DEFAULT '{}',
			created_at TIMESTAMP NOT NULL,
			started_at TIMESTAMP,
			timeout_at TIMESTAMP,
			completed_at TIMESTAMP,
			attempts_left NOT NULL DEFAULT 3,
			stdout TEXT,
			stderr TEXT,
			success BOOL
		);
		CREATE TABLE IF NOT EXISTS data_connections (
			id TEXT NOT NULL PRIMARY KEY,
			type TEXT NOT NULL,
			path TEXT,
			connection_string TEXT
		);
		CREATE TABLE IF NOT EXISTS datasets (
			id TEXT NOT NULL PRIMARY KEY,
			text TEXT NOT NULL
		)
		`,
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Always run migration 0 as a setup.
	_, err = tx.Exec(migrations[0])
	if err != nil {
		tx.Rollback()
		return err
	}
	_, err = tx.Exec("INSERT INTO schema_version VALUES (0, datetime('now')) ON CONFLICT DO NOTHING")
	if err != nil {
		tx.Rollback()
		return err
	}

	maxVersion := 0
	err = tx.QueryRow("SELECT version FROM schema_version ORDER BY version DESC").Scan(&maxVersion)
	if err != nil {
		tx.Rollback()
		return err
	}

	log.Printf("Current schema version is %d. Latest available is %d.", maxVersion, len(migrations)-1)

	for i := maxVersion + 1; i < len(migrations); i++ {
		migrationSQL := migrations[i]

		log.Printf("Running migration %d (%s...)", i, migrationSQL[:40])
		_, err = tx.Exec(migrationSQL)
		if err != nil {
			tx.Rollback()
			return err
		}

		_, err = tx.Exec("INSERT INTO schema_version VALUES ($1, datetime('now'))", i)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
