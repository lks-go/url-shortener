package httphandlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

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

func New(basePath string, deps Dependencies) *Handlers {
	return &Handlers{redirectBasePath: strings.TrimRight(basePath, "/"), service: deps.Service}
}

type Handlers struct {
	redirectBasePath string
	service          Service
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
	_, err = w.Write([]byte(fmt.Sprintf("%s/%s", h.redirectBasePath, id)))
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

func (h *Handlers) ShortenURL(w http.ResponseWriter, req *http.Request) {
	if http.MethodPost != req.Method {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, err := w.Write([]byte(http.StatusText(http.StatusMethodNotAllowed)))
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	body := struct {
		URL string `json:"url"`
	}{}

	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	id, err := h.service.MakeShortURL(req.Context(), body.URL)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	buf := new(bytes.Buffer)
	resp := struct {
		Result string `json:"result"`
	}{
		Result: fmt.Sprintf("%s/%s", h.redirectBasePath, id),
	}
	if err := json.NewEncoder(buf).Encode(resp); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(buf.Bytes())
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
