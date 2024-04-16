package dbstorage

import (
	"context"
	"database/sql"
	"fmt"
)

func New(db *sql.DB) *Storage {
	return &Storage{
		db: db,
	}
}

type Storage struct {
	db *sql.DB
}

func (s *Storage) Save(ctx context.Context, code, url string) error {
	q := `INSERT INTO url (code, url) VALUES($1, $2)`

	_, err := s.db.ExecContext(ctx, q, code, url)
	if err != nil {
		return fmt.Errorf("failed to exec query: %w", err)
	}

	return nil
}

func (s *Storage) Exists(ctx context.Context, code string) (bool, error) {
	q := "SELECT url FROM url WHERE code = $1"

	row := s.db.QueryRowContext(ctx, q, code)
	url := ""
	if err := row.Scan(&url); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to scan row: %w", err)
	}

	if err := row.Err(); err != nil {
		return false, fmt.Errorf("rows error: %w", err)
	}

	return true, nil
}

func (s *Storage) URL(ctx context.Context, code string) (string, error) {
	q := "SELECT url FROM url WHERE code = $1"

	url := ""
	row := s.db.QueryRowContext(ctx, q, code)
	if err := row.Scan(&url); err != nil {
		return "", fmt.Errorf("failed to scan row: %w", err)
	}

	if err := row.Err(); err != nil {
		return "", fmt.Errorf("row error: %w", err)
	}

	return url, nil
}
