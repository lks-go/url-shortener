package entity

import "time"

const (
	AuthTokenHeader  = "auth_token"
	CookieExpires    = time.Hour * 24 * 30
	UserIDHeaderName = "User-Id"
)
