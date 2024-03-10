package app

import (
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"
)

func NewConfig() Config {

	cfg := Config{}

	flag.Var(&cfg.NetAddress, "a", "Net address host:port")
	flag.StringVar(&cfg.RedirectBasePath, "b", "http://localhost:8080", "Base path for short URL")
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
		return ":8080"
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
