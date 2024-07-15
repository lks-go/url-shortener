package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"

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

	a := app.App{
		Config: app.NewConfig(),
	}

	log.Println("Starting server")
	log.Printf("Listen and serve on %s", a.Config.NetAddress.String())
	log.Printf("Base path for short URL '%s'", a.Config.RedirectBasePath)

	go func() {
		err := http.ListenAndServe(":8082", nil)
		log.Fatalf("failed to run profiler http server: %s", err)
	}()

	if err := a.Run(); err != nil {
		log.Fatalf("failed to run application: %s", err)
	}
}
