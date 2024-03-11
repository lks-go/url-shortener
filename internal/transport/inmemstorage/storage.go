package inmemstorage

import (
	"context"
	"errors"
	"sync"

	"github.com/lks-go/url-shortener/internal/transport"
)

func MustNew(memStorage map[string]string) *Storage {
	s, err := New(memStorage)
	if err != nil {
		panic(err)
	}

	return s
}

func New(memStorage map[string]string) (*Storage, error) {

	if memStorage == nil {
		return nil, errors.New("memory storage must not be nil")
	}

	return &Storage{
		data: memStorage,
		mu:   sync.RWMutex{},
	}, nil
}

type Storage struct {
	data map[string]string
	mu   sync.RWMutex
}

func (s *Storage) Save(ctx context.Context, url, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[id] = url

	return nil
}

func (s *Storage) Exists(ctx context.Context, id string) (bool, error) {

	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.data[id]

	return ok, nil
}

func (s *Storage) URL(ctx context.Context, id string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	url, ok := s.data[id]
	if !ok {
		return "", transport.ErrNotFound
	}

	return url, nil
}
