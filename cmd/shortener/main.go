package main

import (
	"log"

	"github.com/lks-go/url-shortener/internal/app"
	"github.com/lks-go/url-shortener/internal/lib/random"
	"github.com/lks-go/url-shortener/internal/service"
	"github.com/lks-go/url-shortener/internal/transport/httphandlers"
	"github.com/lks-go/url-shortener/internal/transport/inmemstorage"
)

func main() {

	memStore := make(map[string]string)

	s := service.New(service.Config{IDSize: 8}, service.Dependencies{
		Storage:      inmemstorage.MustNew(memStore),
		RandomString: random.NewString,
	})

	a := app.App{
		Addr:    ":8080",
		Handler: httphandlers.New(httphandlers.Dependencies{Service: s}),
	}

	if err := a.Run(); err != nil {
		log.Fatal(err)
	}

}
