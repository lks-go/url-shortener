package service

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrURLAlreadyExists = errors.New("URL already exists")
)

type URL struct {
	СorrelationID string
	OriginalURL   string
	Code          string
}

type URLStorage interface {
	Save(ctx context.Context, code, url string) error
	SaveBatch(ctx context.Context, url []URL) error
	Exists(ctx context.Context, code string) (bool, error)
	URL(ctx context.Context, id string) (string, error)
	CodeByURL(ctx context.Context, url string) (string, error)
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
	code, err := s.generateShort(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to assign short: %w", err)
	}

	err = s.storage.Save(ctx, code, url)
	if err != nil && !errors.Is(err, ErrURLAlreadyExists) {
		return "", fmt.Errorf("filed to save url: %w", err)
	}

	if errors.Is(err, ErrURLAlreadyExists) {
		code, err = s.storage.CodeByURL(ctx, url)
		if err != nil {
			return "", fmt.Errorf("failed to get ID by URL: %w", err)
		}

		return code, ErrURLAlreadyExists
	}

	return code, nil
}

func (s *Service) URL(ctx context.Context, id string) (string, error) {
	url, err := s.storage.URL(ctx, id)
	if err != nil {
		return "", fmt.Errorf("failed to get url: %w", err)
	}

	return url, nil
}

func (s *Service) MakeBatchShortURL(ctx context.Context, urls []URL) ([]URL, error) {

	for i := range urls {
		code, err := s.generateShort(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to assign short: %w", err)
		}

		urls[i].Code = code
	}

	if err := s.storage.SaveBatch(ctx, urls); err != nil {
		return nil, fmt.Errorf("failed to save batch of urls: %w", err)
	}

	return urls, nil
}

func (s *Service) generateShort(ctx context.Context) (string, error) {
	short := ""

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		short = s.randomString(s.cfg.IDSize)
		exists, err := s.storage.Exists(ctx, short)
		if err != nil {
			return "", fmt.Errorf("failed to check url id: %w", err)
		}

		if !exists {
			break
		}
	}

	return short, nil
}
