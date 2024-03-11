package app

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	serverAddressDefVal = ":8080"
	baseURLDefVal       = "http://localhost:8080/"
)

func NewConfig() Config {

	cfg := Config{}

	if baseURL, ok := os.LookupEnv("BASE_URL"); ok {
		cfg.RedirectBasePath = baseURL
	} else {
		flag.StringVar(&cfg.RedirectBasePath, "b", baseURLDefVal, "Base path for short URL")
	}

	if srvAddr, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		cfg.NetAddress.Set(srvAddr)
	} else {
		flag.Var(&cfg.NetAddress, "a", "Net address host:port")
	}

	flag.Parse()

	return cfg
}

type Config struct {
	NetAddress       NetAddress
	RedirectBasePath string
}

type NetAddress struct {
	Host string
	Port int
}

func (a *NetAddress) String() string {

	if a.Port == 0 {
		return serverAddressDefVal
	}

	return a.Host + ":" + strconv.Itoa(a.Port)
}

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
