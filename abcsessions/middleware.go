package abcsessions

import "net/http"

type cookieWriter interface {
	SetCookie(cookie *http.Cookie)
	GetCookie(name string) *http.Cookie
}

// sessionsResponseWriter is a wrapper of the ResponseWriter object used to
// buffer the cookies across session API calls, so that they can be written
// at the very end of the response workflow (opposed to written on every operation)
type sessionsResponseWriter struct {
	http.ResponseWriter
	wroteHeader  bool
	wroteCookies bool
	cookies      map[string]*http.Cookie
}

// newSessionsResponseWriter returns a new sessionsResponseWriter object with a pointer to
// the old ResponseWriter object
func newSessionsResponseWriter(w http.ResponseWriter) *sessionsResponseWriter {
	return &sessionsResponseWriter{ResponseWriter: w}
}

// Write calls the underlying ResponseWriter Write func
func (s *sessionsResponseWriter) Write(buf []byte) (int, error) {
	if !s.wroteHeader {
		s.WriteHeader(http.StatusOK)
	}
	return s.ResponseWriter.Write(buf)
}

// WriteHeader sets all cookies in the buffer on the underlying ResponseWriter's
// headers and calls the underlying ResponseWriter WriteHeader func
func (s *sessionsResponseWriter) WriteHeader(code int) {
	s.wroteHeader = true

	// Set all the cookies in the cookie buffer
	if !s.wroteCookies {
		s.wroteCookies = true
		for _, c := range s.cookies {
			http.SetCookie(s.ResponseWriter, c)
		}
	}

	s.ResponseWriter.WriteHeader(code)
}

func (s *sessionsResponseWriter) SetCookie(cookie *http.Cookie) {
	if s.cookies == nil {
		s.cookies = make(map[string]*http.Cookie)
	}

	if len(cookie.Name) == 0 {
		panic("cookie name cannot be empty")
	}

	s.cookies[cookie.Name] = cookie
}

func (s *sessionsResponseWriter) GetCookie(name string) *http.Cookie {
	return s.cookies[name]
}

// Middleware converts the ResponseWriter object to a sessionsResponseWriter
// for buffering cookies across session API requests.
// The sessionsResponseWriter implements cookieWriter.
//
// If you would also like to reset the users session expiry on each
// request (recommended), then use MiddlewareWithReset instead.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Convert the response writer to a sessions response, so we can
		// use its cookie buffering and writing capabilities
		w = newSessionsResponseWriter(w)

		next.ServeHTTP(w, r)
	})
}

type resetExpiryMiddleware struct {
	resetter Resetter
}

// Middleware converts the ResponseWriter object to a sessionsResponseWriter
// for buffering cookies across session API requests.
// The sessionsResponseWriter implements cookieWriter.
//
// MiddlewareWithReset also resets the users session expiry on each request.
// If you do not want this added functionality use Middleware instead.
func (m resetExpiryMiddleware) MiddlewareWithReset(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Convert the response writer to a sessions response, so we can
		// use its cookie buffering and writing capabilities
		w = newSessionsResponseWriter(w)

		err := m.resetter.ResetExpiry(w, r)
		// It's possible that the session hasn't been created yet
		// so there's nothing to reset. In that case, do not explode.
		if err != nil && !IsNoSessionError(err) {
			panic(err)
		}

		next.ServeHTTP(w, r)
	})
}

// ResetMiddleware resets the users session expiry on each request.
//
// Note: Generally you will want to use Middleware or MiddlewareWithReset instead
// of this middleware, but if you have a requirement to insert a middleware
// between the sessions Middleware and the sessions ResetMiddleware then you can
// use the Middleware first, and this one second, as a two-step process, instead
// of the combined MiddlewareWithReset middleware.
//
// It's also important to note that the sessions Middleware must come BEFORE
// this middleware in the chain, or you will get a panic.
func (m resetExpiryMiddleware) ResetMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := m.resetter.ResetExpiry(w, r)
		// It's possible that the session hasn't been created yet
		// so there's nothing to reset. In that case, do not explode.
		if err != nil && !IsNoSessionError(err) {
			panic(err)
		}

		next.ServeHTTP(w, r)
	})
}
