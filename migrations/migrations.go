package migrations

import (
	"database/sql"
	"fmt"
)

func RunUp(db *sql.DB) error {

	if err := createTableShorten(db); err != nil {
		return fmt.Errorf("failed to create table 'shorten': %w", err)
	}

	if err := createIndexForURL(db); err != nil {
		return fmt.Errorf("failed to create index for column 'url' in table 'shorten': %w", err)
	}

	return nil
}

func createTableShorten(db *sql.DB) error {
	q := `CREATE TABLE IF NOT EXISTS shorten (
			id SERIAL PRIMARY KEY,
			code VARCHAR UNIQUE NOT NULL,
			url VARCHAR NOT NULL
		)`

	if _, err := db.Exec(q); err != nil {
		return fmt.Errorf("failed to exec query: %w", err)
	}

	return nil
}

func createIndexForURL(db *sql.DB) error {
	q := `CREATE UNIQUE INDEX IF NOT EXISTS shorten_url_key ON shorten (url)`
	if _, err := db.Exec(q); err != nil {
		return fmt.Errorf("failed to exec query: %w", err)
	}

	return nil
}