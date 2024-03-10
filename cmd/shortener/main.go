package main

import (
	"log"

	"github.com/lks-go/url-shortener/internal/app"
)

func main() {

	a := app.App{
		Config: app.NewConfig(),
	}

	log.Println("Starting server")
	log.Printf("Listen and serve on %s", a.Config.NetAddress.String())
	log.Printf("Base path for short URL '%s'", a.Config.BasePath)
	if err := a.Run(); err != nil {
		log.Fatal(err)
	}

}
