package infilestorage_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lks-go/url-shortener/internal/lib/random"
	"github.com/lks-go/url-shortener/internal/service"
	"github.com/lks-go/url-shortener/internal/transport/infilestorage"
	"github.com/lks-go/url-shortener/pkg/fs"
)

const (
	testFileName = "./test_storage_file"
)

func deleteFile(t *testing.T) {
	t.Helper()

	require.NoError(t, os.Remove(testFileName))
}

func createFile(t *testing.T, data string) {
	t.Helper()

	f, err := os.OpenFile(testFileName, os.O_WRONLY|os.O_CREATE, 0666)
	require.NoError(t, err)

	_, err = f.Write([]byte(data))
	require.NoError(t, err)
}

func TestStorage_Exists(t *testing.T) {
	defer deleteFile(t)

	id1, id2 := "test1", "test2"

	data := fmt.Sprintf(`{"uuid":"xxx1", "short_url":"%s", "original_url":"https://test1.ru"}`, id1)
	createFile(t, data)

	tests := []struct {
		name    string
		id      string
		want    bool
		wantErr bool
	}{
		{
			name:    "must exist test1",
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

			s := infilestorage.New(testFileName)
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
	defer deleteFile(t)
	createFile(t, "")

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
			s := infilestorage.New(testFileName)

			if err := s.Save(context.Background(), tt.id, tt.url); (err != nil) != tt.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
			}

			b, err := os.ReadFile(testFileName)
			require.NoError(t, err)

			rec := fs.Record{}
			require.NoError(t, json.Unmarshal(b, &rec))

			assert.Equal(t, tt.url, rec.OriginalURL, "Save() id not found, want %v", tt.id)
		})
	}
}

func TestStorage_URL(t *testing.T) {
	defer deleteFile(t)

	id1, id2, wantedURL := "test1", "", "https://ya.ru"
	data := fmt.Sprintf(`{"uuid":"xxx1", "short_url":"%s", "original_url":"%s"}`, id1, wantedURL)
	createFile(t, data)

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
			url:     wantedURL,
			wantErr: false,
		},
		{
			name:    "must not return url for id2",
			id:      id2,
			url:     "",
			wantErr: true,
			err:     service.ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := infilestorage.New(testFileName)
			got, err := s.URL(context.Background(), tt.id)
			if tt.wantErr {
				require.ErrorIs(t, err, service.ErrNotFound)
				return
			}
			if got != tt.url {
				t.Errorf("URL() got = %v, want %v", got, tt.url)
			}
		})
	}
}
