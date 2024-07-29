package inmemstorage

import (
	"context"
	"errors"
	"sync"

	"github.com/lks-go/url-shortener/internal/service"
)

// MustNew returns instance of Storage
// if an errors occurs then panic happens
func MustNew(memStoreShortenURLs map[string]string) *Storage {
	s, err := New(memStoreShortenURLs)
	if err != nil {
		panic(err)
	}

	return s
}

// New constructor of new instance of Storage
func New(memStoreShortenURLs map[string]string) (*Storage, error) {

	if memStoreShortenURLs == nil {
		return nil, errors.New("memory storage of shorten URL must not be nil")
	}

	return &Storage{
		shortenURLs: memStoreShortenURLs,
		mu:          sync.RWMutex{},
	}, nil
}

// Storage the main struct implementing the storage
type Storage struct {
	shortenURLs map[string]string
	mu          sync.RWMutex
}

// Save stores a new URL to file storage
func (s *Storage) Save(ctx context.Context, url, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.shortenURLs[id] = url

	return nil
}

// SaveBatch stores array of URLs to file storage
func (s *Storage) SaveBatch(ctx context.Context, url []service.URL) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, u := range url {
		s.shortenURLs[u.Code] = u.OriginalURL
	}

	return nil
}

// Exists checks if URL already exists
func (s *Storage) Exists(ctx context.Context, id string) (bool, error) {

	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.shortenURLs[id]

	return ok, nil
}

// URL returns URL by code
func (s *Storage) URL(ctx context.Context, id string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	url, ok := s.shortenURLs[id]
	if !ok {
		return "", service.ErrNotFound
	}

	return url, nil
}

// CodeByURL returns URLs code by URL
func (s *Storage) CodeByURL(ctx context.Context, url string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for code, u := range s.shortenURLs {
		if u == url {
			return code, nil
		}
	}

	return "", service.ErrNotFound
}

// DeleteURLs removes URLs from storage by codes
func (s *Storage) DeleteURLs(ctx context.Context, codes []string) error {
	return nil
}

// SaveUsersCode stores owner and code of URL to file storage
func (s *Storage) SaveUsersCode(ctx context.Context, userID string, code string) error {

	return nil
}

// UsersURLCodes returns codes of user's URLs
func (s *Storage) UsersURLCodes(ctx context.Context, userID string) ([]string, error) {

	return []string{}, nil
}

// UsersURLs returns list of user's URLs
func (s *Storage) UsersURLs(ctx context.Context, userID string) ([]service.UsersURL, error) {
	return []service.UsersURL{}, nil
}

func (s *Storage) URLCount(ctx context.Context) (int, error) {
	return 0, nil
}

func (s *Storage) UserCount(ctx context.Context) (int, error) {
	return 0, nil
}
