package sessions

import (
	"net/http"
	"time"
)

// CookieOptions for the session cookies themselves
type CookieOptions struct {
	// Name for the session cookie, defaults to "id"
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
	cookie := &http.Cookie{
		Name:     c.Name,
		Value:    value,
		MaxAge:   int(c.MaxAge.Seconds()),
		HttpOnly: c.HTTPOnly,
		Secure:   c.Secure,
	}

	if c.MaxAge != 0 {
		cookie.Expires = time.Now().UTC().Add(c.MaxAge)
	}

	return cookie
}

// getCookieValue returns the cookie value (usually the ID of the session)
// stored in the cookies cache. If it does not exist in the cookies cache
// it will attempt to fetch it from the request headers.
// If this fails it will return nil.
func (c CookieOptions) getCookieValue(w http.ResponseWriter, r *http.Request) (string, error) {
	cookie := w.(cookieWriter).GetCookie(c.Name)
	if cookie != nil {
		return cookie.Value, nil
	}

	reqCookie, err := r.Cookie(c.Name)
	if err != nil {
		return "", errNoSession{}
	}

	return reqCookie.Value, nil
}
