package sessions

import "time"

// CookieOptions for the session cookies themselves.
type CookieOptions struct {
	// Name for the session cookie
	Name string
	// ClientExpiry is set as well as max-age, max-age is determined from
	// ClientExpiry
	ClientExpiry time.Duration
	// Secure ensures the cookie is only given on https connections
	Secure bool
	// HTTPOnly means the browser will never allow JS to touch this cookie
	HTTPOnly bool
}

// NewCookieOptions gives healthy defaults for session cookies
func NewCookieOptions() CookieOptions {
	return CookieOptions{
		Name:         "id",
		ClientExpiry: 0,
		Secure:       true,
		HTTPOnly:     true,
	}
}
