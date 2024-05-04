package middleware

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

const (
	cookieName    = "auth_token"
	cookieExpires = time.Hour * 24 * 30
)

func WithAuth(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var userID string
		var emptyCookie bool
		var claims *Claims

		cookie, err := r.Cookie(cookieName)
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
			claims, err = ParseJWTToken(cookie.Value)
			if err != nil && !errors.Is(err, ErrInvalidToken) && !errors.Is(err, ErrTokenExpired) {
				log.Println("failed to parse jwt:", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
				return
			}

			if claims != nil {
				userID = claims.UserID
			}
		}

		if emptyCookie || errors.Is(err, ErrInvalidToken) || errors.Is(err, ErrTokenExpired) {
			userID = uuid.NewString()
			token, err := BuildNewJWTToken(userID)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
			}

			newCookie := http.Cookie{
				Name:    cookieName,
				Value:   token,
				Expires: time.Now().Add(cookieExpires),
			}

			http.SetCookie(w, &newCookie)
		}

		r.Header.Set("user_id", userID)
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

const (
	tokenExp  = time.Hour * 60
	secretKey = "secret"
)

func BuildNewJWTToken(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExp)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to get signed string: %w", err)
	}

	return tokenString, nil
}

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
)

func ParseJWTToken(token string) (*Claims, error) {
	claims := Claims{}
	parsedToken, err := jwt.ParseWithClaims(token, &claims, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return []byte(secretKey), nil
	})
	if err != nil && !errors.Is(err, jwt.ErrTokenExpired) {
		return nil, fmt.Errorf("failed to parse jwt: %w", err)
	}

	if errors.Is(err, jwt.ErrTokenExpired) {
		return nil, ErrTokenExpired
	}

	if !parsedToken.Valid {
		return nil, ErrInvalidToken
	}

	return &claims, nil
}
