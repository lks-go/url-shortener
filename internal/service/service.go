package service

import (
	"context"
	"errors"
	"fmt"
)

type URL struct {
	СorrelationID string
	OriginalURL   string
	Code          string
}

type UsersURL struct {
	Code        string
	OriginalURL string
}

//go:generate go run github.com/vektra/mockery/v2@v2.24.0 --name=URLStorage
type URLStorage interface {
	Save(ctx context.Context, code, url string) error
	SaveBatch(ctx context.Context, url []URL) error
	Exists(ctx context.Context, code string) (bool, error)
	URL(ctx context.Context, id string) (string, error)
	CodeByURL(ctx context.Context, url string) (string, error)
	SaveUsersCode(ctx context.Context, userID string, code string) error
	UsersURLCodes(ctx context.Context, userID string) ([]string, error)
	DeleteURLs(ctx context.Context, codes []string) error
	UsersURLs(ctx context.Context, userID string) ([]UsersURL, error)
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

func (s *Service) MakeShortURL(ctx context.Context, userID, url string) (string, error) {
	code, err := s.generateShort(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to assign short: %w", err)
	}

	err = s.storage.Save(ctx, code, url)
	if err != nil && !errors.Is(err, ErrURLAlreadyExists) {
		return "", fmt.Errorf("filed to save url: %w", err)
	}

	if err := s.storage.SaveUsersCode(ctx, userID, code); err != nil {
		return "", fmt.Errorf("failed to save user code: %w", err)
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

func (s *Service) MakeBatchShortURL(ctx context.Context, userID string, urls []URL) ([]URL, error) {

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

	for _, u := range urls {
		if err := s.storage.SaveUsersCode(ctx, userID, u.Code); err != nil {
			return nil, fmt.Errorf("failed to save user code: %w", err)
		}
	}

	return urls, nil
}

func (s *Service) UsersURLs(ctx context.Context, userID string) ([]UsersURL, error) {
	userURLs, err := s.storage.UsersURLs(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get users urls from storage: %w", err)
	}

	return userURLs, nil
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
