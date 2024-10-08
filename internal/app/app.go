package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMw "github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"

	"github.com/lks-go/url-shortener/internal/lib/cert"
	"github.com/lks-go/url-shortener/internal/lib/random"
	"github.com/lks-go/url-shortener/internal/service"
	"github.com/lks-go/url-shortener/internal/service/urldeleter"
	"github.com/lks-go/url-shortener/internal/transport/dbstorage"
	"github.com/lks-go/url-shortener/internal/transport/grpchandler"
	"github.com/lks-go/url-shortener/internal/transport/httphandlers"
	"github.com/lks-go/url-shortener/internal/transport/infilestorage"
	"github.com/lks-go/url-shortener/internal/transport/inmemstorage"
	"github.com/lks-go/url-shortener/internal/transport/interceptor"
	"github.com/lks-go/url-shortener/internal/transport/middleware"
	"github.com/lks-go/url-shortener/migrations"
	"github.com/lks-go/url-shortener/pkg/proto"
)

// Service a common interface for app services
type Service interface {
	Start()
	Stop()
}

// App is a struct of the application, contains all necessary dependencies
type App struct {
	Config         Config
	handler        http.Handler
	serviceDeleter Service
	grpcHandler    proto.URLShortenerServer

	pool *sql.DB
}

// Build builds the application
func (a *App) Build() error {
	var (
		storage service.URLStorage
		pool    *sql.DB
		err     error
	)

	switch {
	case a.Config.DatabaseDSN != "":
		pool, err = setupDB(a.Config.DatabaseDSN)
		if err != nil {
			return fmt.Errorf("failed to setup database: %w", err)
		}

		if err := migrations.RunUp(pool); err != nil {
			return fmt.Errorf("failed to run migrations: %w", err)
		}

		storage = dbstorage.New(pool)
	case a.Config.FileStoragePath != "":
		storage = infilestorage.New(a.Config.FileStoragePath)
	default:
		storage = inmemstorage.MustNew(make(map[string]string))
	}

	s := service.New(service.Config{IDSize: 8}, service.Dependencies{
		Storage:      storage,
		RandomString: random.NewString,
	})

	d := urldeleter.NewDeleter(urldeleter.Config{}, urldeleter.Deps{Storage: storage})
	httpHandlers, err := httphandlers.New(httphandlers.Config(a.Config.HTTPHandlerConfig), httphandlers.Dependencies{Service: s, Deleter: d})
	if err != nil {
		return fmt.Errorf("failed to get new http handler: %w", err)
	}

	r := chi.NewRouter()
	r.Use(
		middleware.WithRequestLogger,
		chiMw.Recoverer,
		middleware.WithAuth,
		middleware.WithCompressor,
	)

	if a.Config.ForbiddenAllHandlers {
		r.Use(middleware.WithForbidden)
	}

	r.Post("/", httpHandlers.ShortURL)
	r.Get("/{id}", httpHandlers.Redirect)
	r.Post("/api/shorten", httpHandlers.ShortenURL)
	r.Post("/api/shorten/batch", httpHandlers.ShortenBatchURL)
	r.Get("/api/user/urls", httpHandlers.UsersURLs)
	r.Delete("/api/user/urls", httpHandlers.Delete)
	r.Get("/api/internal/stats", httpHandlers.Stats)

	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		if err := pool.Ping(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	grpcHandler, err := grpchandler.New(grpchandler.Config(a.Config.GRPCHandlerConfig), &grpchandler.Deps{Service: s, Deleter: d})
	if err != nil {
		return fmt.Errorf("failed to get new grpc handler: %w", err)
	}

	a.grpcHandler = grpcHandler
	a.pool = pool
	a.handler = r
	a.serviceDeleter = d

	return nil
}

// StartHTTPServer starts the HTTP server
func (a *App) StartHTTPServer(ctx context.Context) error {
	idleConnsClosed := make(chan struct{})

	srv := http.Server{
		Addr:    a.Config.NetAddress.String(),
		Handler: a.handler,
	}

	go func() {
		<-ctx.Done()
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Printf("failed to shutdown server: %s\n", err)
		}

		close(idleConnsClosed)
	}()

	if a.Config.EnableHTTPS {
		certFile := "cert.pem"
		keyFile := "key.pem"

		err := cert.New(cert.Config{
			CertFileName: certFile,
			KeyFileName:  keyFile,
			Organization: []string{"Yandex.Praktikum"},
			Country:      []string{"RU"},
		})
		if err != nil {
			return fmt.Errorf("failed to get new cert: %w", err)
		}

		if err := srv.ListenAndServeTLS(certFile, keyFile); err != nil {
			return fmt.Errorf("filed to listern and serve TLS: %w", err)
		}

		<-idleConnsClosed

		return nil
	}

	if err := srv.ListenAndServe(); err != nil {
		return fmt.Errorf("filed to listern and serve: %w", err)
	}

	<-idleConnsClosed

	return nil
}

// StartDeleter starts deleter service
func (a *App) StartDeleter(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		a.serviceDeleter.Stop()
	}()

	a.serviceDeleter.Start()

	return nil
}

func (a *App) StartGRPCServer(ctx context.Context) error {
	listen, err := net.Listen("tcp", a.Config.GRPCNetAddress.String())
	if err != nil {
		return fmt.Errorf("filed to start listen address %s: %w", a.Config.GRPCNetAddress.String(), err)
	}

	s := grpc.NewServer(grpc.UnaryInterceptor(interceptor.Auth))
	proto.RegisterURLShortenerServer(s, a.grpcHandler)

	go func() {
		<-ctx.Done()
		listen.Close()
	}()

	if err := s.Serve(listen); err != nil {
		return fmt.Errorf("filed to start serving: %w", err)
	}

	return nil
}

// Exit finishes the app by closing inited db connections and etc
func (a *App) Exit() {
	if err := a.pool.Close(); err != nil {
		log.Printf("failed to close pool: %s", err)
	}
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
