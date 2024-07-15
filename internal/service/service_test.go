package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/lks-go/url-shortener/internal/lib/random"
	"github.com/lks-go/url-shortener/internal/service"
	"github.com/lks-go/url-shortener/internal/service/mocks"
	"github.com/lks-go/url-shortener/internal/transport/inmemstorage"
)

func TestService_MakeShortURL(t *testing.T) {

	wantedID := "abcdef"
	cfg := service.Config{IDSize: 6}
	deps := service.Dependencies{
		Storage: inmemstorage.MustNew(map[string]string{}),
		RandomString: func(size int) string {
			return wantedID
		},
	}

	s := service.New(cfg, deps)

	tests := []struct {
		name    string
		url     string
		wantID  string
		wantErr bool
	}{
		{
			name:    "make new short url",
			url:     "http://ya.ru",
			wantID:  wantedID,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.MakeShortURL(context.Background(), "", tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("MakeShortURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantID {
				t.Errorf("MakeShortURL() got = %v, want %v", got, tt.wantID)
			}
		})
	}
}

func TestService_URL(t *testing.T) {

	id := "abcdef"
	url := "http://ya.ru"

	deps := service.Dependencies{
		Storage: inmemstorage.MustNew(map[string]string{id: url}),
	}

	s := service.New(service.Config{}, deps)
	tests := []struct {
		name    string
		id      string
		wantURL string
		wantErr bool
	}{
		{
			name:    "get existed url",
			id:      id,
			wantURL: url,
			wantErr: false,
		},
		{
			name:    "get not existed url",
			id:      "any",
			wantURL: "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.URL(context.Background(), tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("URL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantURL {
				t.Errorf("URL() got = %v, want %v", got, tt.wantURL)
			}
		})
	}
}

func BenchmarkService_MakeShortURL(b *testing.B) {
	cfg := service.Config{IDSize: 6}
	URLStorageMock := mocks.NewURLStorage(b)
	URLStorageMock.On("Exists", mock.Anything, mock.Anything).Return(false, nil)
	URLStorageMock.On("Save", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	URLStorageMock.On("SaveUsersCode", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	deps := service.Dependencies{
		Storage:      URLStorageMock,
		RandomString: random.NewString,
	}

	s := service.New(cfg, deps)
	for i := 0; i < b.N; i++ {
		s.MakeShortURL(context.Background(), "", "")
	}
}

func BenchmarkService_URL(b *testing.B) {
	cfg := service.Config{}
	URLStorageMock := mocks.NewURLStorage(b)
	URLStorageMock.On("URL", mock.Anything, mock.Anything).Return("", nil)

	deps := service.Dependencies{
		Storage: URLStorageMock,
	}

	s := service.New(cfg, deps)
	for i := 0; i < b.N; i++ {
		s.URL(context.Background(), "")
	}
}
