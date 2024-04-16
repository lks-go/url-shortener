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

	"github.com/sirupsen/logrus"

	"github.com/lks-go/url-shortener/internal/service"
	"github.com/lks-go/url-shortener/internal/transport"
)

//go:generate go run github.com/vektra/mockery/v2@v2.24.0 --name=Service
type Service interface {
	MakeBatchShortURL(ctx context.Context, urls []service.URL) ([]service.URL, error)
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
		fmt.Println(err)
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

func (h *Handlers) ShortenBatchURL(w http.ResponseWriter, req *http.Request) {
	type url struct {
		CorrelationID string `json:"correlation_id"`
		OriginalURL   string `json:"original_url"`
	}

	body := make([]url, 0)
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	urlList := make([]service.URL, 0, len(body))
	for _, u := range body {
		urlList = append(urlList, service.URL{
			СorrelationID: u.CorrelationID,
			OriginalURL:   u.OriginalURL,
		})
	}

	shortURLList, err := h.service.MakeBatchShortURL(req.Context(), urlList)
	if err != nil {
		logrus.Errorf("failed to make batch short urls: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	type respUrl struct {
		CorrelationID string `json:"correlation_id"`
		ShortURL      string `json:"short_url"`
	}

	resp := make([]respUrl, 0, len(shortURLList))
	for _, u := range shortURLList {
		resp = append(resp, respUrl{
			CorrelationID: u.СorrelationID,
			ShortURL:      fmt.Sprintf("%s/%s", h.redirectBasePath, u.Code),
		})
	}

	buf := new(bytes.Buffer)
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
