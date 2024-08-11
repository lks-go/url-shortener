package middleware

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/lks-go/url-shortener/internal/entity"
	"github.com/lks-go/url-shortener/internal/lib/jwt"
)

// WithAuth checks user's cookie and jwt
// if cookie is empty generates new jwt and set new cooker to headers
func WithAuth(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var userID string
		var emptyCookie bool
		var claims *jwt.Claims

		cookie, err := r.Cookie(entity.AuthTokenHeader)
		if err != nil {
			switch {
			case errors.Is(err, http.ErrNoCookie):
				emptyCookie = true
			default:
				log.Println("cookie error:", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
				return
			}
		}

		if !emptyCookie {
			claims, err = jwt.ParseJWTToken(cookie.Value)
			if err != nil && !errors.Is(err, jwt.ErrInvalidToken) && !errors.Is(err, jwt.ErrTokenExpired) {
				log.Println("failed to parse jwt:", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
				return
			}

			if claims != nil && claims.UserID == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if claims != nil {
				userID = claims.UserID
			}
		}

		if emptyCookie || errors.Is(err, jwt.ErrInvalidToken) || errors.Is(err, jwt.ErrTokenExpired) {
			userID = uuid.NewString()
			token, err := jwt.BuildNewJWTToken(userID)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
			}

			newCookie := http.Cookie{
				Name:    entity.AuthTokenHeader,
				Value:   token,
				Expires: time.Now().Add(entity.CookieExpires),
			}

			http.SetCookie(w, &newCookie)
		}

		r.Header.Set(entity.UserIDHeaderName, userID)
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
