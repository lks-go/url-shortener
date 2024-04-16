package migrations

import (
	"database/sql"
	"fmt"
)

func RunUp(db *sql.DB) error {

	q := `CREATE TABLE IF NOT EXISTS url (
		id SERIAL PRIMARY KEY,
		code VARCHAR UNIQUE,
		url VARCHAR
	)`

	if _, err := db.Exec(q); err != nil {
		return fmt.Errorf("failed to exec query: %w", err)
	}

	return nil
}
