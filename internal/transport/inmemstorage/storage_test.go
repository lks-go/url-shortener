package inmemstorage_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lks-go/url-shortener/internal/lib/random"
	"github.com/lks-go/url-shortener/internal/transport"
	"github.com/lks-go/url-shortener/internal/transport/inmemstorage"
)

func TestStorage_Exists(t *testing.T) {

	id1, id2 := random.NewString(6), random.NewString(6)

	mem := map[string]string{
		id1: "https://ya.ru",
	}

	tests := []struct {
		name    string
		id      string
		want    bool
		wantErr bool
	}{
		{
			name:    "must exist id1",
			id:      id1,
			want:    true,
			wantErr: false,
		},
		{
			name:    "must not exist id2",
			id:      id2,
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := inmemstorage.MustNew(mem)
			got, err := s.Exists(context.Background(), tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Exists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Exists() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStorage_Save(t *testing.T) {

	tests := []struct {
		name    string
		url     string
		id      string
		wantErr bool
	}{
		{
			name:    "new url",
			url:     "https://ya.ru",
			id:      random.NewString(6),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mem := map[string]string{}
			s := inmemstorage.MustNew(mem)

			if err := s.Save(context.Background(), tt.url, tt.id); (err != nil) != tt.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
			}

			url, ok := mem[tt.id]
			assert.Equal(t, true, ok, "Save() id not found, want %v", tt.id)
			assert.Equal(t, tt.url, url, "Save() url not equal, got %v,  want %v", url, tt.id)
		})
	}
}

func TestStorage_URL(t *testing.T) {

	id1, id2 := random.NewString(6), random.NewString(6)
	wantedUrl := "https://ya.ru"

	mem := map[string]string{
		id1: wantedUrl,
	}

	tests := []struct {
		name    string
		id      string
		url     string
		wantErr bool
		err     error
	}{
		{
			name:    "must return url for id1",
			id:      id1,
			url:     wantedUrl,
			wantErr: false,
		},
		{
			name:    "must not return url for id2",
			id:      id2,
			url:     "",
			wantErr: true,
			err:     transport.ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := inmemstorage.MustNew(mem)
			got, err := s.URL(context.Background(), tt.id)
			if tt.wantErr {
				require.ErrorIs(t, err, transport.ErrNotFound)
				return
			}
			if got != tt.url {
				t.Errorf("URL() got = %v, want %v", got, tt.url)
			}
		})
	}
}
