package inmem_storage

import (
	"context"
	"sync"

	"github.com/lks-go/url-shortener/internal/transport"
)

func New() *Storage {
	return &Storage{
		data: make(map[string]string),
		mu:   sync.RWMutex{},
	}
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

func (s *Storage) Url(ctx context.Context, id string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	url, ok := s.data[id]
	if !ok {
		return "", transport.ErrNotFound
	}

	return url, nil
}
