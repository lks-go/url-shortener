package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	"github.com/lks-go/url-shortener/internal/app"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	log.Printf("Build version: %s\n", buildVersion)
	log.Printf("Build date: %s\n", buildDate)
	log.Printf("Build commit: %s\n", buildCommit)

	cfg, err := app.NewConfig()
	if err != nil {
		log.Fatalf("failed to get new config: %s", err)
	}

	a := app.App{
		Config: cfg,
	}

	log.Println("Starting server")
	log.Printf("Listen and serve on %s", a.Config.NetAddress.String())
	log.Printf("Base path for short URL '%s'", a.Config.HandlerConfig.RedirectBasePath)

	go func() {
		err := http.ListenAndServe(":8083", nil)
		log.Fatalf("failed to run profiler http server: %s", err)
	}()

	if err := a.Init(); err != nil {
		log.Fatalf("failed to init application: %s", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		if err := a.StartHTTPServer(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("HTTP server error: %w", err)
		}

		return nil
	})

	g.Go(func() error {
		if err := a.StartDeleter(ctx); err != nil {
			return fmt.Errorf("service deleter error: %w", err)
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		log.Fatalf("group error: %s", err)
	}

	a.Exit()

	log.Println("Application successfully stopped")
}
