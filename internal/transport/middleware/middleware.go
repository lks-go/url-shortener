package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

func WithCompressor(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ow := w

		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			cw := newCompressWriter(w)
			ow = cw
			defer cw.Close()
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			
			r.Body = cr
			defer cr.Close()
		}

		next.ServeHTTP(ow, r)

	}

	return http.HandlerFunc(fn)
}

func WithRequestLogger(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rp := responseData{}

		lwr := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   &rp,
		}

		defer func() {
			logrus.New().WithFields(logrus.Fields{
				"uri":      r.RequestURI,
				"method":   r.Method,
				"duration": time.Since(start),
				"status":   rp.status,
				"size":     rp.size,
			}).Info("HTTP request")
		}()

		next.ServeHTTP(&lwr, r)
	}

	return http.HandlerFunc(fn)
}

type responseData struct {
	status int
	size   int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size // захватываем размер
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // захватываем код статуса
}
