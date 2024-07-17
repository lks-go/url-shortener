package app

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"database/sql"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	chiMw "github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/lks-go/url-shortener/internal/lib/random"
	"github.com/lks-go/url-shortener/internal/service"
	"github.com/lks-go/url-shortener/internal/service/urldeleter"
	"github.com/lks-go/url-shortener/internal/transport/dbstorage"
	"github.com/lks-go/url-shortener/internal/transport/httphandlers"
	"github.com/lks-go/url-shortener/internal/transport/infilestorage"
	"github.com/lks-go/url-shortener/internal/transport/inmemstorage"
	"github.com/lks-go/url-shortener/internal/transport/middleware"
	"github.com/lks-go/url-shortener/migrations"
)

// App is a struct of the application, contains all necessary dependencies
type App struct {
	Config Config
}

var (
	storage service.URLStorage
	pool    *sql.DB
)

// Run builds and starts the application
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
		storage = infilestorage.New(a.Config.FileStoragePath)
	default:
		storage = inmemstorage.MustNew(make(map[string]string))
	}

	s := service.New(service.Config{IDSize: 8}, service.Dependencies{
		Storage:      storage,
		RandomString: random.NewString,
	})

	d := urldeleter.NewDeleter(urldeleter.Config{}, urldeleter.Deps{Storage: storage})
	d.Start()
	defer d.Stop()

	h := httphandlers.New(a.Config.RedirectBasePath, httphandlers.Dependencies{Service: s, Deleter: d})

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
	r.Get("/api/user/urls", h.UsersURLs)
	r.Delete("/api/user/urls", h.Delete)
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		if err := pool.Ping(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	return listenAndServe(a, r)
}

func listenAndServe(a *App, r http.Handler) error {
	if a.Config.EnableHTTPS {
		certFile, keyFile, err := setupCert()
		if err != nil {
			return fmt.Errorf("failed got setup cert: %w", err)
		}

		return http.ListenAndServeTLS(a.Config.NetAddress.String(), certFile, keyFile, r)
	}

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

func setupCert() (certFile string, keyFile string, err error) {

	certPEMFileName := "cert.pem"
	keyPEMFileName := "key.pem"

	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			Organization: []string{"Yandex.Praktikum"},
			Country:      []string{"RU"},
		},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate rsa key: %w", err)
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &privateKey.PublicKey, privateKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to create certificate: %w", err)
	}

	var certPEMBuf bytes.Buffer
	pem.Encode(&certPEMBuf, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	certPEMFile, err := os.Create(certPEMFileName)
	if err != nil {
		return "", "", fmt.Errorf("failed to create cert.pem: %w", err)
	}

	_, err = certPEMFile.Write(certPEMBuf.Bytes())
	if err != nil {
		return "", "", fmt.Errorf("failed to write cert.pem: %w", err)
	}

	var privateKeyPEMBuf bytes.Buffer
	pem.Encode(&privateKeyPEMBuf, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	privateKeyPEMFile, err := os.Create(keyPEMFileName)
	if err != nil {
		return "", "", fmt.Errorf("failed to create key.pem: %w", err)
	}

	_, err = privateKeyPEMFile.Write(privateKeyPEMBuf.Bytes())
	if err != nil {
		return "", "", fmt.Errorf("failed to write key.pem: %w", err)
	}

	return certPEMFileName, keyPEMFileName, nil
}
