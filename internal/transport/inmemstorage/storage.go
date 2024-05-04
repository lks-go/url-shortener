package inmemstorage

import (
	"context"
	"errors"
	"sync"

	"github.com/lks-go/url-shortener/internal/service"
)

func MustNew(memStoreShortenURLs map[string]string, memUsersURLsCodes map[string][]string) *Storage {
	s, err := New(memStoreShortenURLs, memUsersURLsCodes)
	if err != nil {
		panic(err)
	}

	return s
}

func New(memStoreShortenURLs map[string]string, memUsersURLsCodes map[string][]string) (*Storage, error) {

	if memStoreShortenURLs == nil {
		return nil, errors.New("memory storage of shorten URL must not be nil")
	}

	if memUsersURLsCodes == nil {
		return nil, errors.New("memory storage of user's URL codes  must not be nil")
	}

	return &Storage{
		shortenURLs: memStoreShortenURLs,
		mu:          sync.RWMutex{},
	}, nil
}

type Storage struct {
	shortenURLs map[string]string
	usersURLs   map[string][]string
	mu          sync.RWMutex
}

func (s *Storage) Save(ctx context.Context, url, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.shortenURLs[id] = url

	return nil
}

func (s *Storage) SaveBatch(ctx context.Context, url []service.URL) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, u := range url {
		s.shortenURLs[u.Code] = u.OriginalURL
	}

	return nil
}

func (s *Storage) Exists(ctx context.Context, id string) (bool, error) {

	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.shortenURLs[id]

	return ok, nil
}

func (s *Storage) URL(ctx context.Context, id string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	url, ok := s.shortenURLs[id]
	if !ok {
		return "", service.ErrNotFound
	}

	return url, nil
}

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

func (s *Storage) SaveUsersCode(ctx context.Context, userID string, code string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	codes, ok := s.usersURLs[userID]
	if !ok {
		codes = make([]string, 0)
	}

	for _, c := range codes {
		if c == code {
			return service.ErrRecordAlreadyExists
		}
	}

	s.usersURLs[userID] = append(codes, code)

	return nil
}

func (s *Storage) UsersURLCodes(ctx context.Context, userID string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	codes, ok := s.usersURLs[userID]
	if !ok {
		return nil, service.ErrNotFound
	}

	return codes, nil
}
