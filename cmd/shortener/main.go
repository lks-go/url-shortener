package main

import (
	"log"

	"github.com/lks-go/url-shortener/internal/app"
)

func main() {

	a := app.App{
		Addr: ":8080",
	}

	if err := a.Run(); err != nil {
		log.Fatal(err)
	}

}
