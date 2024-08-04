package middleware

import "net/http"

// WithForbidden перекрывает доступ ко всем ручкам
func WithForbidden(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}

	return http.HandlerFunc(fn)
}
