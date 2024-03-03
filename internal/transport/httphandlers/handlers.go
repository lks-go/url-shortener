package httphandlers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/lks-go/url-shortener/internal/transport"
)

type Service interface {
	MakeShortURL(ctx context.Context, url string) (string, error)
	URL(ctx context.Context, id string) (string, error)
}

type Dependencies struct {
	Service
}

func New(deps Dependencies) *Handlers {
	return &Handlers{service: deps.Service}
}

type Handlers struct {
	service Service
}

func (h *Handlers) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	defer func() {
		if rec := recover(); rec != nil {
			http.Error(w, "panic", 500)
		}
	}()

	switch {
	case match(r.URL.Path, `^/$`):
		h.ShortURL(w, r)
	case match(r.URL.Path, `^/\w+$`):
		h.Redirect(w, r)
	default:
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
	}
}

func match(path string, pattern string) bool {
	regExp := regexp.MustCompile(pattern)
	return regExp.MatchString(path)
}

func (h *Handlers) ShortURL(w http.ResponseWriter, req *http.Request) {
	if http.MethodPost != req.Method {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(http.StatusText(http.StatusMethodNotAllowed)))
		return
	}

	b, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	id, err := h.service.MakeShortURL(req.Context(), string(b))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("http://%s/%s", req.Host, id)))
}

func (h *Handlers) Redirect(w http.ResponseWriter, req *http.Request) {
	if http.MethodGet != req.Method {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(http.StatusText(http.StatusMethodNotAllowed)))
		return
	}

	matches := regexp.MustCompile(`/(\w+)`).FindStringSubmatch(req.URL.Path)
	if len(matches) < 1 {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	id := matches[1]
	url, err := h.service.URL(req.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, transport.ErrNotFound):
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(http.StatusText(http.StatusNotFound)))
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
