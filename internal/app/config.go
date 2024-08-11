package app

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
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
func NewConfig() (Config, error) {
	cfg := Config{}
	var redirectBasePath string

	flag.Var(&cfg.NetAddress, "a", "HTTP net address host:port")
	flag.Var(&cfg.GRPCNetAddress, "ga", "GRPC net address host:port")
	flag.StringVar(&redirectBasePath, "b", DefaultBaseURL, "Base path for short URL")
	flag.StringVar(&cfg.FileStoragePath, "f", DefaultFSPath, "Path for file storage")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "Database connection string")
	flag.BoolVar(&cfg.EnableHTTPS, "s", false, "Enable HTTPS")
	flag.StringVar(&cfg.HTTPHandlerConfig.TrustedSubnet, "t", "", "Trusted subnet")

	cfg.HTTPHandlerConfig.RedirectBasePath, cfg.GRPCHandlerConfig.RedirectBasePath = redirectBasePath, redirectBasePath

	var configFile string
	flag.StringVar(&configFile, "c", "", "Config json file path")

	flag.Parse()

	if baseURL, ok := os.LookupEnv("BASE_URL"); ok {
		cfg.HTTPHandlerConfig.RedirectBasePath = baseURL
		cfg.GRPCHandlerConfig.RedirectBasePath = baseURL
	}

	if srvAddr, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		cfg.NetAddress.Set(srvAddr)
	}

	if srvAddr, ok := os.LookupEnv("GRPC_SERVER_ADDRESS"); ok {
		cfg.GRPCNetAddress.Set(srvAddr)
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

	if trustedSubnet, ok := os.LookupEnv("TRUSTED_SUBNET"); ok {
		cfg.HTTPHandlerConfig.TrustedSubnet = trustedSubnet
	}

	if configFile != "" {
		jsonCfg, err := parseJSONConfig(configFile)
		if err != nil {
			return Config{}, fmt.Errorf("failed to parse json config: %w", err)
		}

		mapJSONConfig(&cfg, jsonCfg)
	}

	return cfg, nil
}

// Config contains application config
type Config struct {
	NetAddress           NetAddress
	GRPCNetAddress       NetAddress
	FileStoragePath      string
	DatabaseDSN          string
	EnableHTTPS          bool
	HTTPHandlerConfig    HTTPHandlerConfig
	GRPCHandlerConfig    GRPCHandlerConfig
	ForbiddenAllHandlers bool
}

// HTTPHandlerConfig конфиг для HTTP хендлеров
type HTTPHandlerConfig struct {
	RedirectBasePath string
	TrustedSubnet    string
}

// GRPCHandlerConfig конфиг для GRPC хендлеров
type GRPCHandlerConfig struct {
	RedirectBasePath string
	TrustedSubnet    string
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

type jsonConfig struct {
	ServerAddress     string `json:"server_address"`
	GRPCServerAddress string `json:"grpc_server_address"`
	BaseURL           string `json:"base_url"`
	FileStoragePath   string `json:"file_storage_path"`
	DatabaseDSN       string `json:"database_dsn"`
	EnableHTTPS       bool   `json:"enable_https"`
	TrustedSubnet     string `json:"trusted_subnet"`
}

func parseJSONConfig(file string) (*jsonConfig, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("filed to open file: %w", err)
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read bytes: %w", err)
	}

	cfg := jsonConfig{}
	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

func mapJSONConfig(cfg *Config, jsonCfg *jsonConfig) {
	if cfg.NetAddress.String() == "" {
		cfg.NetAddress.Set(jsonCfg.ServerAddress)
	}

	if cfg.GRPCNetAddress.String() == "" {
		cfg.GRPCNetAddress.Set(jsonCfg.GRPCServerAddress)
	}

	if cfg.HTTPHandlerConfig.RedirectBasePath == "" {
		cfg.HTTPHandlerConfig.RedirectBasePath = jsonCfg.BaseURL
		cfg.GRPCHandlerConfig.RedirectBasePath = jsonCfg.BaseURL
	}

	if cfg.FileStoragePath == "" {
		cfg.FileStoragePath = jsonCfg.FileStoragePath
	}

	if cfg.DatabaseDSN == "" {
		cfg.DatabaseDSN = jsonCfg.DatabaseDSN
	}

	if !cfg.EnableHTTPS {
		cfg.EnableHTTPS = jsonCfg.EnableHTTPS
	}

	if cfg.HTTPHandlerConfig.TrustedSubnet == "" {
		cfg.HTTPHandlerConfig.TrustedSubnet = jsonCfg.TrustedSubnet
		if jsonCfg.TrustedSubnet == "" {
			cfg.ForbiddenAllHandlers = true
		}
	}
}
