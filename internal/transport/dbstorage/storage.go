package dbstorage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lks-go/url-shortener/internal/service"
)

func New(db *sql.DB) *Storage {
	return &Storage{
		db: db,
	}
}

type Storage struct {
	db *sql.DB
}

func (s *Storage) SaveBatch(ctx context.Context, urls []service.URL) error {

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT INTO url (code, url) VALUES($1, $2)`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}

	for _, u := range urls {
		_, err = stmt.ExecContext(ctx, u.Code, u.OriginalURL)
		if err != nil {
			return fmt.Errorf("failed to exec query: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
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
