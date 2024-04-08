package service

import (
	"context"
	"fmt"
)

type URLStorage interface {
	Save(ctx context.Context, id, url string) error
	Exists(ctx context.Context, id string) (bool, error)
	URL(ctx context.Context, id string) (string, error)
}

type Config struct {
	IDSize int
}

type Dependencies struct {
	Storage      URLStorage
	RandomString func(size int) string
}

func New(cfg Config, deps Dependencies) *Service {
	return &Service{
		cfg:          cfg,
		storage:      deps.Storage,
		randomString: deps.RandomString,
	}
}

type Service struct {
	cfg          Config
	storage      URLStorage
	randomString func(size int) string
}

func (s *Service) MakeShortURL(ctx context.Context, url string) (string, error) {

	id := ""
	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		id = s.randomString(s.cfg.IDSize)
		exists, err := s.storage.Exists(ctx, id)
		if err != nil {
			return "", fmt.Errorf("failed to check url id: %w", err)
		}

		if !exists {
			break
		}
	}

	if err := s.storage.Save(ctx, id, url); err != nil {
		return "", fmt.Errorf("filed to save url: %w", err)
	}

	return id, nil
}

func (s *Service) URL(ctx context.Context, id string) (string, error) {
	url, err := s.storage.URL(ctx, id)
	if err != nil {
		return "", fmt.Errorf("failed to get url: %w", err)
	}

	return url, nil
}
