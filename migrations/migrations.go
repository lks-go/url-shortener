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

	if err := createTableUserCodes(db); err != nil {
		return fmt.Errorf("failed to create table 'user_codes': %w", err)
	}

	if err := createIndexForUserCodes(db); err != nil {
		return fmt.Errorf("failed to create index for 'user_codes': %w", err)
	}

	if err := addColumnDeleteToShorten(db); err != nil {
		return fmt.Errorf("failed to add column to 'shorten': %w", err)
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

func createTableUserCodes(db *sql.DB) error {
	q := `CREATE TABLE IF NOT EXISTS USER_CODES (
		USER_ID UUID  NOT NULL,
		CODE VARCHAR NOT NULL
	);`

	if _, err := db.Exec(q); err != nil {
		return fmt.Errorf("failed to exec query: %w", err)
	}

	return nil
}

func createIndexForUserCodes(db *sql.DB) error {
	q := `CREATE UNIQUE INDEX IF NOT EXISTS USER_ID_CODE_KEY ON USER_CODES(USER_ID, CODE);`
	if _, err := db.Exec(q); err != nil {
		return fmt.Errorf("failed to exec query: %w", err)
	}

	return nil
}

func addColumnDeleteToShorten(db *sql.DB) error {
	q := `ALTER TABLE shorten ADD COLUMN IF NOT EXISTS deleted BOOL;`
	if _, err := db.Exec(q); err != nil {
		return fmt.Errorf("failed to exec query: %w", err)
	}

	return nil
}
