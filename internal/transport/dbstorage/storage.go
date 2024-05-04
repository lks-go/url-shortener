package dbstorage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

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

	stmt, err := tx.Prepare(`INSERT INTO shorten (code, url) VALUES($1, $2)`)
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
	q := `INSERT INTO shorten (code, url) VALUES($1, $2)`

	_, err := s.db.ExecContext(ctx, q, code, url)
	if err != nil {
		if err, ok := err.(*pgconn.PgError); ok {
			if err.Code == pgerrcode.UniqueViolation {
				return service.ErrURLAlreadyExists
			}
		}

		return fmt.Errorf("failed to exec query: %w", err)
	}

	return nil
}

func (s *Storage) Exists(ctx context.Context, code string) (bool, error) {
	q := "SELECT url FROM shorten WHERE code = $1"

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
	q := "SELECT url FROM shorten WHERE code = $1"

	url := ""
	row := s.db.QueryRowContext(ctx, q, code)
	if err := row.Scan(&url); err != nil {
		if err == sql.ErrNoRows {
			return "", service.ErrNotFound
		}
		return "", fmt.Errorf("failed to scan row: %w", err)
	}

	if err := row.Err(); err != nil {
		return "", fmt.Errorf("row error: %w", err)
	}

	return url, nil
}

func (s *Storage) CodeByURL(ctx context.Context, url string) (string, error) {
	q := "SELECT code FROM shorten WHERE url = $1"

	code := ""
	row := s.db.QueryRowContext(ctx, q, url)
	if err := row.Scan(&code); err != nil {
		return "", fmt.Errorf("failed to scan row: %w", err)
	}

	if err := row.Err(); err != nil {
		return "", fmt.Errorf("row error: %w", err)
	}

	return code, nil
}

func (s *Storage) SaveUsersCode(ctx context.Context, userID string, code string) error {
	q := `INSERT INTO USER_CODES (USER_ID, CODE) VALUES ($1, $2);`

	_, err := s.db.ExecContext(ctx, q, userID, code)
	if err != nil {
		if err, ok := err.(*pgconn.PgError); ok {
			if err.Code == pgerrcode.UniqueViolation {
				return service.ErrRecordAlreadyExists
			}
		}

		return fmt.Errorf("failed to exec query: %w", err)
	}

	return nil
}

func (s *Storage) UsersURLCodes(ctx context.Context, userID string) ([]string, error) {
	q := `SELECT CODE FROM user_codes WHERE USER_ID = $1;`

	rows, err := s.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to make query: %w", err)
	}
	defer rows.Close()

	codes := make([]string, 0)
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, fmt.Errorf("failed to scan code: %w", err)
		}

		codes = append(codes, code)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return codes, nil
}
