package main

import (
	"log"

	"github.com/lks-go/url-shortener/internal/app"
	"github.com/lks-go/url-shortener/internal/lib/random"
	"github.com/lks-go/url-shortener/internal/service"
	"github.com/lks-go/url-shortener/internal/transport/http-handlers"
	inmem_storage "github.com/lks-go/url-shortener/internal/transport/inmem-storage"
)

func main() {

	s := service.New(service.Config{IdSize: 8}, service.Dependencies{
		Storage:      inmem_storage.New(),
		RandomString: random.NewString,
	})

	a := app.App{
		Addr:    ":8088",
		Handler: http_handlers.New(http_handlers.Dependencies{Service: s}),
	}

	if err := a.Run(); err != nil {
		log.Fatal(err)
	}

}
