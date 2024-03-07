package httphandlers_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/lks-go/url-shortener/internal/transport"
	"github.com/lks-go/url-shortener/internal/transport/httphandlers"
	"github.com/lks-go/url-shortener/internal/transport/httphandlers/mocks"
)

func TestHandlers_Redirect(t *testing.T) {
	serviceMock := mocks.NewService(t)

	deps := httphandlers.Dependencies{
		Service: serviceMock,
	}
	h := httphandlers.New(deps)

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
					Return("", transport.ErrNotFound).Once()
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

	host := "test"
	id := "any-rand-id"
	serviceMock := mocks.NewService(t)

	deps := httphandlers.Dependencies{
		Service: serviceMock,
	}
	h := httphandlers.New(deps)

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
			wantResp:     fmt.Sprintf("http://%s/%s", host, id),
			callMocks: func() {
				serviceMock.On("MakeShortURL", mock.Anything, "https://ya.ru").Return(id, nil).Once()
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
				serviceMock.On("MakeShortURL", mock.Anything, "https://ya.ru").Return("", err).Once()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.callMocks()

			w := httptest.NewRecorder()
			r := httptest.NewRequest(tt.method, tt.target, tt.body)
			r.Host = host

			h.ShortURL(w, r)

			assert.Equal(t, tt.wantResp, w.Body.String())
			assert.Equal(t, tt.wantHTTPCode, w.Code)
		})
	}
}
