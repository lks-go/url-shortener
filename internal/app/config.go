package app

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Default settings
const (
	DefaultServerAddress = ":8080"
	DefaultBaseURL       = "http://localhost:8080"
	DefaultFSPath        = "/tmp/short-url-db.json"
)

// NewConfig builds and returns application config
func NewConfig() Config {

	cfg := Config{}
	flag.Var(&cfg.NetAddress, "a", "Net address host:port")
	flag.StringVar(&cfg.RedirectBasePath, "b", DefaultBaseURL, "Base path for short URL")
	flag.StringVar(&cfg.FileStoragePath, "f", DefaultFSPath, "Path for file storage")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "Database connection string")
	flag.BoolVar(&cfg.EnableHTTPS, "s", false, "")
	flag.Parse()

	if baseURL, ok := os.LookupEnv("BASE_URL"); ok {
		cfg.RedirectBasePath = baseURL
	}

	if srvAddr, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		cfg.NetAddress.Set(srvAddr)
	}

	if fsPath, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		cfg.FileStoragePath = fsPath
	}

	if dbConnString, ok := os.LookupEnv("DATABASE_DSN"); ok {
		cfg.DatabaseDSN = dbConnString
	}

	if enableHTTPS, ok := os.LookupEnv("ENABLE_HTTPS"); ok {
		cfg.EnableHTTPS = enableHTTPS == "true" || enableHTTPS == "1"
	}

	return cfg
}

// Config contains application config
type Config struct {
	NetAddress       NetAddress
	RedirectBasePath string
	FileStoragePath  string
	DatabaseDSN      string
	EnableHTTPS      bool
}

// NetAddress contains net config
type NetAddress struct {
	Host string
	Port int
}

// String builds and returns address which the application listens
func (a *NetAddress) String() string {

	if a.Port == 0 {
		return DefaultServerAddress
	}

	return a.Host + ":" + strconv.Itoa(a.Port)
}

// Set pareses address and sets Host and Port
func (a *NetAddress) Set(s string) error {
	addr := strings.Split(s, ":")

	if len(addr) < 2 {
		return errors.New("invalid address value")
	}

	p, err := strconv.Atoi(addr[1])
	if err != nil {
		return fmt.Errorf("failed to parse port: %w", err)
	}

	a.Host = addr[0]
	a.Port = p

	return nil
}
