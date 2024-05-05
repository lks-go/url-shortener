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

type Config struct {
	UrlsFilename string
}

func New(filename string) *Storage {
	return &Storage{
		urlsFilename: filename,
		mu:           sync.Mutex{},
	}
}

type Storage struct {
	urlsFilename string
	mu           sync.Mutex
}

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

func (s *Storage) SaveUsersCode(ctx context.Context, userID string, code string) error {
	return nil
}

func (s *Storage) UsersURLCodes(ctx context.Context, userID string) ([]string, error) {
	return []string{}, nil
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
