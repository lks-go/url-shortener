package httphandlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/lks-go/url-shortener/internal/service"
)

type Config struct {
	RedirectBasePath string
	TrustedSubnet    string
}

// Service это интерфейс сервиса отвечающего за обратоку входящих http запросов
type Service interface {
	MakeBatchShortURL(ctx context.Context, userID string, urls []service.URL) ([]service.URL, error)
	MakeShortURL(ctx context.Context, userID, url string) (string, error)
	URL(ctx context.Context, id string) (string, error)
	UsersURLs(ctx context.Context, userID string) ([]service.UsersURL, error)
	Stats(ctx context.Context) (*service.StatsInfo, error)
}

// Deleter это интерфейс сервиса отвечающего за получение запроса на удаление
type Deleter interface {
	Delete(ctx context.Context, userID string, codes []string) error
}

// Dependencies основные зависимости
type Dependencies struct {
	Service
	Deleter
}

// New is a constructor of *Handlers
func New(cfg Config, deps Dependencies) (*Handlers, error) {
	var ipNet *net.IPNet
	var err error

	if cfg.TrustedSubnet != "" {
		_, ipNet, err = net.ParseCIDR(cfg.TrustedSubnet)
		if err != nil {
			return nil, fmt.Errorf("failed to parse trusted subnet: %w", err)
		}
	}

	return &Handlers{
		redirectBasePath: strings.TrimRight(cfg.RedirectBasePath, "/"),
		service:          deps.Service,
		deleter:          deps.Deleter,
		ipNet:            ipNet,
	}, nil
}

// Handlers is a main structure of httphandlers
type Handlers struct {
	redirectBasePath string
	service          Service
	deleter          Deleter
	ipNet            *net.IPNet
}

// ShortURL ручка для создания короткой ссылки
func (h *Handlers) ShortURL(w http.ResponseWriter, req *http.Request) {
	if http.MethodPost != req.Method {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, err := w.Write([]byte(http.StatusText(http.StatusMethodNotAllowed)))
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	userID, ok := req.Header["User-Id"]
	if !ok || len(userID) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	b, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	id, err := h.service.MakeShortURL(req.Context(), userID[0], string(b))
	if err != nil && !errors.Is(err, service.ErrURLAlreadyExists) {
		logrus.Errorf("failed to make short url: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if errors.Is(err, service.ErrURLAlreadyExists) {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}
	_, err = w.Write([]byte(fmt.Sprintf("%s/%s", h.redirectBasePath, id)))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// Redirect запрашивает в сервисе оригинальный урл по короткой ссылке
// и если такой урл есть, то возвращает клиенту http код ответа 307
// и оригинальный урл в заголовке Location
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

	code := matches[1]
	url, err := h.service.URL(req.Context(), code)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			w.WriteHeader(http.StatusNotFound)
			_, err = w.Write([]byte(http.StatusText(http.StatusNotFound)))
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		case errors.Is(err, service.ErrDeleted):
			w.WriteHeader(http.StatusGone)
		default:
			logrus.Errorf("failed to get url by code [%s]: %s", code, err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// ShortenBatchURL возвращает короткие ссылки пачку урлов
// тело запроса должно содержать массив объектов
//
//	Пример:
//	 [
//			{"correlation_id": "example_id", "original_url": "https://ya.ru"}
//	 ]
func (h *Handlers) ShortenBatchURL(w http.ResponseWriter, req *http.Request) {
	userID, ok := req.Header["User-Id"]
	if !ok || len(userID) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

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

	shortURLList, err := h.service.MakeBatchShortURL(req.Context(), userID[0], urlList)
	if err != nil {
		logrus.Errorf("failed to make batch short urls: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	type respURL struct {
		CorrelationID string `json:"correlation_id"`
		ShortURL      string `json:"short_url"`
	}

	resp := make([]respURL, 0, len(shortURLList))
	for _, u := range shortURLList {
		resp = append(resp, respURL{
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

// ShortenURL создает короткую ссылку для урла
func (h *Handlers) ShortenURL(w http.ResponseWriter, req *http.Request) {
	if http.MethodPost != req.Method {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, err := w.Write([]byte(http.StatusText(http.StatusMethodNotAllowed)))
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	userID, ok := req.Header["User-Id"]
	if !ok || len(userID) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	body := struct {
		URL string `json:"url"`
	}{}

	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	isConflict := false

	code, err := h.service.MakeShortURL(req.Context(), userID[0], body.URL)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrURLAlreadyExists):
			logrus.Warnf("url [%s] already exists: %s", body.URL, err)
			isConflict = true
		default:
			logrus.Errorf("failed to make short url: %s", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	resp, err := h.shortenURLResponse(code)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	if isConflict {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}

	_, err = w.Write(resp)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// UsersURLs возращает списко ссылок, добавленных для пользователя
func (h *Handlers) UsersURLs(w http.ResponseWriter, req *http.Request) {
	userID, ok := req.Header["User-Id"]
	if !ok || len(userID) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	urls, err := h.service.UsersURLs(req.Context(), userID[0])
	if err != nil {
		logrus.Errorf("failed to get users urls: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	type respURL struct {
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
	}

	resp := make([]respURL, 0, len(urls))
	for _, u := range urls {
		resp = append(resp, respURL{
			ShortURL:    fmt.Sprintf("%s/%s", h.redirectBasePath, u.Code),
			OriginalURL: u.OriginalURL,
		})
	}

	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(resp); err != nil {
		logrus.Errorf("failed encode response to json: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	if len(resp) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(buf.Bytes())
	if err != nil {
		logrus.Errorf("failed write response: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (h *Handlers) shortenURLResponse(code string) ([]byte, error) {
	buf := new(bytes.Buffer)
	resp := struct {
		Result string `json:"result"`
	}{
		Result: fmt.Sprintf("%s/%s", h.redirectBasePath, code),
	}

	if err := json.NewEncoder(buf).Encode(resp); err != nil {
		return nil, fmt.Errorf("failed to encode response: %w", err)
	}

	return buf.Bytes(), nil
}

// Delete принимает запрос на удаление уролов
// в теле запроса передается список кодов коротких ссылок
// хендлер не дожидается фактического удаления урлов и возвращает http код 202
func (h *Handlers) Delete(w http.ResponseWriter, req *http.Request) {
	userID, ok := req.Header["User-Id"]
	if !ok || len(userID) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	b, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer req.Body.Close()

	var codes []string
	if err = json.Unmarshal(b, &codes); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	go func() {
		if err := h.deleter.Delete(context.Background(), userID[0], codes); err != nil {
			logrus.Errorf("failed to delete urls (userId = %s, codes = [%v]): %s", userID, codes, err)
		}
	}()

	w.WriteHeader(http.StatusAccepted)
}

// Stats возвращает количество сокращённых URL и количество пользователей в сервисе
func (h *Handlers) Stats(w http.ResponseWriter, req *http.Request) {
	ip := req.Header.Get("X-Real-IP")
	if h.ipNet != nil && !h.ipNet.Contains(net.ParseIP(ip)) {
		logrus.Errorf("ip %s is not in trusted subnet", ip)
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	statsInfo, err := h.service.Stats(req.Context())
	if err != nil {
		logrus.Errorf("failed to get stats in handler: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp := struct {
		URLS  int `json:"urls"`
		USERS int `json:"users"`
	}{
		URLS:  statsInfo.URLCount,
		USERS: statsInfo.UserCount,
	}

	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(resp); err != nil {
		logrus.Errorf("failed encode response to json: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(buf.Bytes())
	if err != nil {
		logrus.Errorf("failed to write response: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
