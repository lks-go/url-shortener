package app

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMw "github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/lks-go/url-shortener/internal/lib/random"
	"github.com/lks-go/url-shortener/internal/service"
	"github.com/lks-go/url-shortener/internal/transport/dbstorage"
	"github.com/lks-go/url-shortener/internal/transport/httphandlers"
	"github.com/lks-go/url-shortener/internal/transport/infilestorage"
	"github.com/lks-go/url-shortener/internal/transport/inmemstorage"
	"github.com/lks-go/url-shortener/internal/transport/middleware"
	"github.com/lks-go/url-shortener/migrations"
)

type App struct {
	Config Config
}

var (
	storage service.URLStorage
	pool    *sql.DB
)

func (a *App) Run() error {

	switch {
	case a.Config.DatabaseDSN != "":
		pool, err := setupDB(a.Config.DatabaseDSN)
		if err != nil {
			return fmt.Errorf("failed to setup database: %w", err)
		}
		defer pool.Close()

		if err := migrations.RunUp(pool); err != nil {
			return fmt.Errorf("failed to run migrations: %w", err)
		}

		storage = dbstorage.New(pool)
	case a.Config.FileStoragePath != "":
		storage = infilestorage.New(infilestorage.Config{
			UrlsFilename:  a.Config.FileStoragePath,
			UsersURLCodes: a.Config.FileStoragePath,
		})
	default:
		memStoreShortenURLs := make(map[string]string)
		memUsersURLsCodes := make(map[string][]string)
		storage = inmemstorage.MustNew(memStoreShortenURLs, memUsersURLsCodes)
	}

	s := service.New(service.Config{IDSize: 8}, service.Dependencies{
		Storage:      storage,
		RandomString: random.NewString,
	})

	h := httphandlers.New(a.Config.RedirectBasePath, httphandlers.Dependencies{Service: s})

	r := chi.NewRouter()
	r.Use(
		middleware.WithRequestLogger,
		chiMw.Recoverer,
		middleware.WithAuth,
		middleware.WithCompressor,
	)

	r.Get("/{id}", h.Redirect)
	r.Post("/", h.ShortURL)
	r.Post("/api/shorten", h.ShortenURL)
	r.Post("/api/shorten/batch", h.ShortenBatchURL)
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		if err := pool.Ping(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	return http.ListenAndServe(a.Config.NetAddress.String(), r)
}

func setupDB(dsn string) (*sql.DB, error) {
	pool, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := pool.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database after connect: %w", err)
	}

	return pool, nil
}
