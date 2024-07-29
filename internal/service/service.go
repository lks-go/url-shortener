// The main service
package service

import (
	"context"
	"errors"
	"fmt"
)

// URL a main domain struct of URL
type URL struct {
	СorrelationID string
	OriginalURL   string
	Code          string
}

// UsersURL a domain struct describes which shorten code belongs to URL
type UsersURL struct {
	Code        string
	OriginalURL string
}

// URLStorage is an interface of URL storage
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
	URLCount(ctx context.Context) (int, error)
	UserCount(ctx context.Context) (int, error)
}

// Config is a service config
type Config struct {
	IDSize int
}

// Dependencies is a struct contains main service dependencies
type Dependencies struct {
	Storage      URLStorage
	RandomString func(size int) string
}

// New is a service constructor
// to declare Service use only the constructor recommended
func New(cfg Config, deps Dependencies) *Service {
	return &Service{
		cfg:          cfg,
		storage:      deps.Storage,
		randomString: deps.RandomString,
	}
}

// Service is a main service structure
type Service struct {
	cfg          Config
	storage      URLStorage
	randomString func(size int) string
}

// MakeShortURL generates code and save generated code with URL
// returns generated code
// if code or URL already exist returns the error ErrURLAlreadyExists
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

// URL find and return URL by id
func (s *Service) URL(ctx context.Context, id string) (string, error) {
	url, err := s.storage.URL(ctx, id)
	if err != nil {
		return "", fmt.Errorf("failed to get url: %w", err)
	}

	return url, nil
}

// MakeBatchShortURL generates codes for batch of URLs
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

// UsersURLs reruns list of URLs added by user
func (s *Service) UsersURLs(ctx context.Context, userID string) ([]UsersURL, error) {
	userURLs, err := s.storage.UsersURLs(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get users urls from storage: %w", err)
	}

	return userURLs, nil
}

type StatsInfo struct {
	URLCount  int
	UserCount int
}

// Stats gets user and URL count from DB
func (s *Service) Stats(ctx context.Context) (*StatsInfo, error) {
	urlCount, err := s.storage.URLCount(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get URL count: %w", err)
	}

	userCount, err := s.storage.UserCount(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user count: %w", err)
	}

	return &StatsInfo{URLCount: urlCount, UserCount: userCount}, nil
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
