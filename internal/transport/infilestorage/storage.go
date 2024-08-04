package infilestorage

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"sync"

	"github.com/lks-go/url-shortener/internal/service"
	"github.com/lks-go/url-shortener/pkg/fs"
)

// Config of file storage
type Config struct {
	UrlsFilename string
}

// New creates a new instance of Storage
func New(filename string) *Storage {
	return &Storage{
		urlsFilename: filename,
		mu:           sync.Mutex{},
	}
}

// Storage the main struct
type Storage struct {
	urlsFilename string
	mu           sync.Mutex
}

// Save stores a new URL to file storage
func (s *Storage) Save(ctx context.Context, id, url string) error {

	l, err := s.recordList(s.urlsFilename)
	if err != nil {
		return fmt.Errorf("failed to get url list: %w", err)
	}

	r := fs.Record{
		UUID:        strconv.Itoa(len(l) + 1),
		ShortURL:    id,
		OriginalURL: url,
	}

	if err := s.append(s.urlsFilename, &r); err != nil {
		return fmt.Errorf("failed to append row: %w", err)
	}

	return nil
}

// Exists checks if URL already exists
func (s *Storage) Exists(ctx context.Context, id string) (bool, error) {

	l, err := s.recordList(s.urlsFilename)
	if err != nil {
		return false, fmt.Errorf("failed to get url list: %w", err)
	}

	for _, row := range l {
		if row.ShortURL == id {
			return true, nil
		}
	}

	return false, nil
}

// URL returns URL by code
func (s *Storage) URL(ctx context.Context, id string) (string, error) {

	l, err := s.recordList(s.urlsFilename)
	if err != nil {
		return "", fmt.Errorf("failed to get url list: %w", err)
	}

	for _, row := range l {
		if row.ShortURL == id {
			return row.OriginalURL, nil
		}
	}

	return "", service.ErrNotFound
}

// SaveBatch stores array of URLs to file storage
func (s *Storage) SaveBatch(ctx context.Context, url []service.URL) error {
	l, err := s.recordList(s.urlsFilename)
	if err != nil {
		return fmt.Errorf("failed to get url list: %w", err)
	}

	for _, u := range url {
		r := fs.Record{
			UUID:        strconv.Itoa(len(l) + 1),
			ShortURL:    u.Code,
			OriginalURL: u.OriginalURL,
		}

		if err := s.append(s.urlsFilename, &r); err != nil {
			return fmt.Errorf("failed to append row: %w", err)
		}
	}

	return nil
}

// CodeByURL returns URLs code by URL
func (s *Storage) CodeByURL(ctx context.Context, url string) (string, error) {
	l, err := s.recordList(s.urlsFilename)
	if err != nil {
		return "", fmt.Errorf("failed to get url list: %w", err)
	}

	for _, row := range l {
		if row.OriginalURL == url {
			return row.ShortURL, nil
		}
	}

	return "", service.ErrNotFound
}

// SaveUsersCode stores owner and code of URL to file storage
func (s *Storage) SaveUsersCode(ctx context.Context, userID string, code string) error {
	return nil
}

// UsersURLCodes returns codes of user's URLs
func (s *Storage) UsersURLCodes(ctx context.Context, userID string) ([]string, error) {
	return []string{}, nil
}

// DeleteURLs removes URLs from storage by codes
func (s *Storage) DeleteURLs(ctx context.Context, codes []string) error {
	return nil
}

// UsersURLs returns list of user's URLs
func (s *Storage) UsersURLs(ctx context.Context, userID string) ([]service.UsersURL, error) {
	return []service.UsersURL{}, nil
}

func (s *Storage) recordList(fileName string) ([]fs.Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	consumer, err := fs.NewConsumer(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to get consumer: %w", err)
	}
	defer consumer.Close()

	rows := make([]fs.Record, 0)
	for {
		rec := fs.Record{}
		err := consumer.ReadRow(&rec)
		if err != nil {
			if err == io.EOF {
				break
			}

			return nil, fmt.Errorf("failed to read record: %w", err)
		}

		rows = append(rows, rec)
	}

	return rows, nil
}

func (s *Storage) append(fileName string, r *fs.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	producer, err := fs.NewProducer(fileName)
	if err != nil {
		return fmt.Errorf("failed to get producer: %w", err)
	}
	defer producer.Close()

	if err := producer.WriteRow(r); err != nil {
		return fmt.Errorf("filed to write row: %w", err)
	}

	return nil
}

// URLCount blank
func (s *Storage) URLCount(ctx context.Context) (int, error) {
	return 0, nil
}

// UserCount blank
func (s *Storage) UserCount(ctx context.Context) (int, error) {
	return 0, nil
}
