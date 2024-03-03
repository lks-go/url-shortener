package app

import "net/http"

type App struct {
	Addr    string
	Handler http.Handler
}

func (a *App) Run() error {

	mux := http.NewServeMux()
	mux.Handle("/", a.Handler)

	return http.ListenAndServe(a.Addr, mux)
}
