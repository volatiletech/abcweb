package sessions

import (
	"net/http"
	"time"
)

// CookieOptions for the session cookies themselves
type CookieOptions struct {
	// Name for the session cookie
	Name string
	// MaxAge sets the max-age and the expires fields of a cookie
	MaxAge time.Duration
	// Secure ensures the cookie is only given on https connections
	Secure bool
	// HTTPOnly means the browser will never allow JS to touch this cookie
	HTTPOnly bool
}

// NewCookieOptions gives healthy defaults for session cookies
func NewCookieOptions() CookieOptions {
	return CookieOptions{
		Name:     "id",
		MaxAge:   0,
		Secure:   true,
		HTTPOnly: true,
	}
}

func (c CookieOptions) makeCookie(value string) *http.Cookie {
	return &http.Cookie{
		Name:     c.Name,
		Value:    value,
		MaxAge:   int(c.MaxAge.Seconds()),
		Expires:  time.Now().UTC().Add(c.MaxAge),
		HttpOnly: c.HTTPOnly,
		Secure:   c.Secure,
	}
}
