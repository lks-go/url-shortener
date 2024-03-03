package service

import (
	"context"
	"fmt"
)

type UrlStorage interface {
	Save(ctx context.Context, id, url string) error
	Exists(ctx context.Context, id string) (bool, error)
	Url(ctx context.Context, id string) (string, error)
}

type Config struct {
	IdSize int
}

type Dependencies struct {
	Storage      UrlStorage
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
	storage      UrlStorage
	randomString func(size int) string
}

func (s *Service) MakeShortUrl(ctx context.Context, url string) (string, error) {

	id := ""
	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		id = s.randomString(s.cfg.IdSize)
		exists, err := s.storage.Exists(ctx, id)
		if err != nil {
			return "", fmt.Errorf("failed to check url id: %w", err)
		}

		if !exists {
			break
		}
	}

	if err := s.storage.Save(ctx, url, id); err != nil {
		return "", fmt.Errorf("filed to save url: %w", err)
	}

	return id, nil
}

func (s *Service) Url(ctx context.Context, id string) (string, error) {
	url, err := s.storage.Url(ctx, id)
	if err != nil {
		return "", fmt.Errorf("failed to get url: %w", err)
	}

	return url, nil
}
