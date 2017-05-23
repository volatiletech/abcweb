package abcsessions

import (
	"net/http"
	"time"
)

// CookieOptions for the session cookies themselves.
// See https://tools.ietf.org/html/rfc6265 for details.
type CookieOptions struct {
	// Domain is the domain name the cookie is for
	Domain string
	// Path is the URI path the cookie is for
	Path string
	// Name for the session cookie, defaults to "id"
	Name string
	// MaxAge sets the max-age and the expires fields of a cookie
	// A value of 0 means the browser will expire the session on browser close
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
		Path:     "/",
		MaxAge:   0,
		Secure:   true,
		HTTPOnly: true,
	}
}

func (c CookieOptions) makeCookie(value string) *http.Cookie {
	cookie := &http.Cookie{
		Domain:   c.Domain,
		Path:     c.Path,
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

// deleteCookie sets the cookie to a deleted value to force the client to delete
func (c CookieOptions) deleteCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		// If the browser refuses to delete it, set value to "" so subsequent
		// requests replace it when it does not point to a valid session id.
		Path:     c.Path,
		Domain:   c.Domain,
		Value:    "",
		Name:     c.Name,
		MaxAge:   -1,
		Expires:  time.Now().UTC().AddDate(-1, 0, 0),
		HttpOnly: c.HTTPOnly,
		Secure:   c.Secure,
	}

	w.(cookieWriter).SetCookie(cookie)
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
