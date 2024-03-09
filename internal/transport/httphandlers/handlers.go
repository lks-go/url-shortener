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

//go:generate go run github.com/vektra/mockery/v2@v2.24.0 --name=Service
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

func (h *Handlers) ShortURL(w http.ResponseWriter, req *http.Request) {
	if http.MethodPost != req.Method {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, err := w.Write([]byte(http.StatusText(http.StatusMethodNotAllowed)))
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
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
	_, err = w.Write([]byte(fmt.Sprintf("http://%s/%s", req.Host, id)))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (h *Handlers) Redirect(w http.ResponseWriter, req *http.Request) {
	if http.MethodGet != req.Method {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, err := w.Write([]byte(http.StatusText(http.StatusMethodNotAllowed)))
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
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
			_, err = w.Write([]byte(http.StatusText(http.StatusNotFound)))
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
