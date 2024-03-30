package middleware

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

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
