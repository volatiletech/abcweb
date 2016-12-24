package sessions

import "net/http"

type cookieWriter interface {
	SetCookie(cookie *http.Cookie)
	GetCookie(name string) *http.Cookie
}

// response is a wrapper of the ResponseWriter object used to
// buffer the cookies across session API calls, so that they can be written
// at the very end of the response workflow (opposed to written on every operation)
type response struct {
	http.ResponseWriter
	wroteHeader bool
	cookies     map[string]*http.Cookie
}

// newResponse returns a new response object with a pointer to
// the old ResponseWriter object
func newResponse(w http.ResponseWriter) *response {
	return &response{ResponseWriter: w}
}

// Write calls the underlying ResponseWriter Write func
func (r *response) Write(buf []byte) (int, error) {
	if !r.wroteHeader {
		r.wroteHeader = true
		r.WriteHeader(http.StatusOK)
	}
	return r.ResponseWriter.Write(buf)
}

// WriteHeader sets all cookies in the buffer on the underlying ResponseWriter's
// headers and calls the underlying ResponseWriter WriteHeader func
func (r *response) WriteHeader(code int) {
	// Set all the cookies in the cookie buffer
	for _, c := range r.cookies {
		http.SetCookie(r.ResponseWriter, c)
	}

	r.ResponseWriter.WriteHeader(code)
}

func (r *response) SetCookie(cookie *http.Cookie) {
	if r.cookies == nil {
		r.cookies = make(map[string]*http.Cookie)
	}

	if len(cookie.Name) == 0 {
		panic("cookie name cannot be empty")
	}

	r.cookies[cookie.Name] = cookie
}

func (r *response) GetCookie(name string) *http.Cookie {
	return r.cookies[name]
}

// Middleware converts the ResponseWriter object to a SessionsResponse
// for buffering cookies across session API requests
func Middleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// Convert the response writer to a sessions response, so we can
		// use its cookie buffering and writing capabilities
		w = newResponse(w)

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

type resetExpiryMiddleware struct {
	resetter Resetter
}

// ResetMiddleware resets the users session expiry on each request
func (m resetExpiryMiddleware) ResetMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		err := m.resetter.ResetExpiry(w, r)
		if err != nil {
			panic(err)
		}

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
