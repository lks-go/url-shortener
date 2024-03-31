package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMw "github.com/go-chi/chi/v5/middleware"

	"github.com/lks-go/url-shortener/internal/lib/random"
	"github.com/lks-go/url-shortener/internal/service"
	"github.com/lks-go/url-shortener/internal/transport/httphandlers"
	"github.com/lks-go/url-shortener/internal/transport/inmemstorage"
	"github.com/lks-go/url-shortener/internal/transport/middleware"
)

type App struct {
	Config Config
}

func (a *App) Run() error {

	memStore := make(map[string]string)

	s := service.New(service.Config{IDSize: 8}, service.Dependencies{
		Storage:      inmemstorage.MustNew(memStore),
		RandomString: random.NewString,
	})

	h := httphandlers.New(a.Config.RedirectBasePath, httphandlers.Dependencies{Service: s})

	r := chi.NewRouter()
	r.Use(
		middleware.WithRequestLogger,
		chiMw.Recoverer,
	)

	r.Get("/{id}", h.Redirect)
	r.Post("/", h.ShortURL)
	r.Post("/api/shorten", h.ShortenURL)

	return http.ListenAndServe(a.Config.NetAddress.String(), r)
}
