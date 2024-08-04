package httphandlers_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/lks-go/url-shortener/internal/service"
	"github.com/lks-go/url-shortener/internal/transport/httphandlers"
	"github.com/lks-go/url-shortener/internal/transport/httphandlers/mocks"
	"github.com/lks-go/url-shortener/internal/transport/middleware"
)

func TestHandlers_Redirect(t *testing.T) {
	serviceMock := mocks.NewService(t)

	deps := httphandlers.Dependencies{
		Service: serviceMock,
	}
	h, err := httphandlers.New(httphandlers.Config{RedirectBasePath: "/"}, deps)
	assert.NoError(t, err)

	type header struct {
		key, value string
	}

	tests := []struct {
		name         string
		method       string
		target       string
		wantHTTPCode int
		wantHeader   header
		callMocks    func()
	}{
		{
			name:         "existed id and redirect",
			method:       http.MethodGet,
			target:       "/123456",
			wantHTTPCode: http.StatusTemporaryRedirect,
			wantHeader: header{
				key:   "Location",
				value: "https://ya.ru",
			},
			callMocks: func() {
				serviceMock.On("URL", mock.Anything, "123456").
					Return("https://ya.ru", nil).Once()
			},
		},
		{
			name:         "not found",
			method:       http.MethodGet,
			target:       "/123457",
			wantHTTPCode: http.StatusNotFound,
			wantHeader: header{
				key:   "Location",
				value: "",
			},
			callMocks: func() {
				serviceMock.On("URL", mock.Anything, "123457").
					Return("", service.ErrNotFound).Once()
			},
		},
		{
			name:         "internal server error",
			method:       http.MethodGet,
			target:       "/123456",
			wantHTTPCode: http.StatusInternalServerError,
			wantHeader: header{
				key:   "Location",
				value: "",
			},
			callMocks: func() {
				serviceMock.On("URL", mock.Anything, "123456").
					Return("", errors.New("unexpected error")).Once()
			},
		},
		{
			name:         "method not allowed",
			method:       http.MethodPost,
			target:       "/123456",
			wantHTTPCode: http.StatusMethodNotAllowed,
			wantHeader: header{
				key:   "Location",
				value: "",
			},
			callMocks: func() {},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.callMocks()

			w := httptest.NewRecorder()
			r := httptest.NewRequest(tt.method, tt.target, nil)

			h.Redirect(w, r)

			assert.Equal(t, tt.wantHTTPCode, w.Code)
			assert.Equal(t, tt.wantHeader.value, w.Header().Get(tt.wantHeader.key))
		})
	}
}

func TestHandlers_ShortURL(t *testing.T) {

	basePath := "http://localhost:8080"
	id := "any-rand-id"
	serviceMock := mocks.NewService(t)

	deps := httphandlers.Dependencies{
		Service: serviceMock,
	}
	h, err := httphandlers.New(httphandlers.Config{RedirectBasePath: basePath}, deps)
	assert.NoError(t, err)

	tests := []struct {
		name         string
		method       string
		target       string
		body         io.Reader
		wantHTTPCode int
		wantResp     string
		callMocks    func()
	}{
		{
			name:         "successful request",
			method:       http.MethodPost,
			target:       "/",
			body:         bytes.NewReader([]byte("https://ya.ru")),
			wantHTTPCode: http.StatusCreated,
			wantResp:     fmt.Sprintf("%s/%s", basePath, id),
			callMocks: func() {
				serviceMock.On("MakeShortURL", mock.Anything, mock.Anything, "https://ya.ru").Return(id, nil).Once()
			},
		},
		{
			name:         "method not allowed",
			method:       http.MethodGet,
			target:       "/",
			body:         bytes.NewReader([]byte("https://ya.ru")),
			wantHTTPCode: http.StatusMethodNotAllowed,
			wantResp:     http.StatusText(http.StatusMethodNotAllowed),
			callMocks:    func() {},
		},
		{
			name:         "internal server error",
			method:       http.MethodPost,
			target:       "/",
			body:         bytes.NewReader([]byte("https://ya.ru")),
			wantHTTPCode: http.StatusInternalServerError,
			wantResp:     http.StatusText(http.StatusInternalServerError) + "\n",
			callMocks: func() {
				err := errors.New("any error")
				serviceMock.On("MakeShortURL", mock.Anything, mock.Anything, "https://ya.ru").Return("", err).Once()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.callMocks()

			w := httptest.NewRecorder()
			r := httptest.NewRequest(tt.method, tt.target, tt.body)

			hh := middleware.WithAuth(http.HandlerFunc(h.ShortURL))
			hh.ServeHTTP(w, r)

			assert.Equal(t, tt.wantResp, w.Body.String())
			assert.Equal(t, tt.wantHTTPCode, w.Code)
		})
	}
}

func TestHandlers_ShortenURL(t *testing.T) {

	basePath := "http://localhost:8080"
	id := "any-rand-id"
	serviceMock := mocks.NewService(t)

	deps := httphandlers.Dependencies{
		Service: serviceMock,
	}
	h, err := httphandlers.New(httphandlers.Config{RedirectBasePath: basePath}, deps)
	assert.NoError(t, err)

	tests := []struct {
		name         string
		method       string
		target       string
		body         io.Reader
		wantHTTPCode int
		wantResp     string
		callMocks    func()
	}{
		{
			name:         "successful request",
			method:       http.MethodPost,
			target:       "/api/shorten",
			body:         bytes.NewReader([]byte(`{"url": "https://ya.ru"}`)),
			wantHTTPCode: http.StatusCreated,
			wantResp:     fmt.Sprintf("{\"result\":\"%s/%s\"}\n", basePath, id),
			callMocks: func() {
				serviceMock.On("MakeShortURL", mock.Anything, mock.Anything, "https://ya.ru").Return(id, nil).Once()
			},
		},
		{
			name:         "method not allowed",
			method:       http.MethodGet,
			target:       "/api/shorten",
			body:         bytes.NewReader([]byte(`{"url": "https://ya.ru"}`)),
			wantHTTPCode: http.StatusMethodNotAllowed,
			wantResp:     http.StatusText(http.StatusMethodNotAllowed),
			callMocks:    func() {},
		},
		{
			name:         "internal server error",
			method:       http.MethodPost,
			target:       "/api/shorten",
			body:         bytes.NewReader([]byte(`{"url": "https://ya.ru"}`)),
			wantHTTPCode: http.StatusInternalServerError,
			wantResp:     http.StatusText(http.StatusInternalServerError) + "\n",
			callMocks: func() {
				err := errors.New("any error")
				serviceMock.On("MakeShortURL", mock.Anything, mock.Anything, "https://ya.ru").Return("", err).Once()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.callMocks()

			w := httptest.NewRecorder()
			r := httptest.NewRequest(tt.method, tt.target, tt.body)

			hh := middleware.WithAuth(http.HandlerFunc(h.ShortenURL))
			hh.ServeHTTP(w, r)

			assert.Equal(t, tt.wantResp, w.Body.String())
			assert.Equal(t, tt.wantHTTPCode, w.Code)
		})
	}
}

func TestHandlers_Stats(t *testing.T) {
	basePath := "http://localhost:8080"
	serviceMock := mocks.NewService(t)

	deps := httphandlers.Dependencies{
		Service: serviceMock,
	}

	cfg := httphandlers.Config{RedirectBasePath: basePath, TrustedSubnet: "248.133.71.0/24"}
	h, err := httphandlers.New(cfg, deps)
	assert.NoError(t, err)

	expectedRespBody := `{"urls": 23,"users": 10}`
	tests := []struct {
		name         string
		ip           string
		wantHTTPCode int
		wantResp     string
		callMocks    func()
	}{
		{
			name:         "successful request",
			ip:           "248.133.71.33",
			wantHTTPCode: http.StatusOK,
			wantResp:     expectedRespBody,
			callMocks: func() {
				serviceMock.On("Stats", mock.Anything).
					Return(&service.StatsInfo{URLCount: 23, UserCount: 10}, nil).Once()
			},
		},
		{
			name:         "internal error",
			ip:           "248.133.71.33",
			wantHTTPCode: http.StatusInternalServerError,
			wantResp:     "",
			callMocks: func() {
				serviceMock.On("Stats", mock.Anything).
					Return(nil, errors.New("any error")).Once()
			},
		},
		{
			name:         "forbidden",
			ip:           "248.133.72.1",
			wantHTTPCode: http.StatusForbidden,
			wantResp:     "",
			callMocks:    func() {},
		},
		{
			name:         "forbidden",
			ip:           "248.134.71.5",
			wantHTTPCode: http.StatusForbidden,
			wantResp:     "",
			callMocks:    func() {},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.callMocks()

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.Header.Add("X-Real-IP", tt.ip)

			hh := middleware.WithAuth(http.HandlerFunc(h.Stats))
			hh.ServeHTTP(w, r)

			if tt.wantResp != "" {
				assert.JSONEq(t, tt.wantResp, w.Body.String())
			}
			assert.Equal(t, tt.wantHTTPCode, w.Code)
		})
	}

}

func TestHandlers_StatsTrustedSubnet(t *testing.T) {

	ip := net.ParseIP("248.133.72.1")

	_, ipNet, err := net.ParseCIDR("248.133.71.0/24")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(ipNet.Contains(ip))
}
