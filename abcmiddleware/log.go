package abcmiddleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	chimiddleware "github.com/go-chi/chi/middleware"
	"go.uber.org/zap"
)

// zapResponseWriter is a wrapper that includes that http status and size for logging
type zapResponseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

// Zap middleware handles web request logging using Zap
func (m Middleware) Zap(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		zw := &zapResponseWriter{ResponseWriter: w}

		// Serve the request
		next.ServeHTTP(zw, r)

		// Write the request log line
		writeZap(m.Log, r, t, zw.status, zw.size)
	}

	return http.HandlerFunc(fn)
}

// RequestIDLogger middleware creates a derived logger to include logging of the
// Request ID, and inserts it into the context object
func (m Middleware) RequestIDLogger(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		requestID := chimiddleware.GetReqID(r.Context())
		derivedLogger := m.Log.With(zap.String("request_id", requestID))
		r = r.WithContext(context.WithValue(r.Context(), CtxLoggerKey, derivedLogger))
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

// Log returns the Request ID scoped logger from the request Context
// and panics if it cannot be found. This function is only ever used
// by your controllers if your app uses the RequestID middlewares,
// otherwise you should use the controller's receiver logger directly.
func Log(r *http.Request) *zap.Logger {
	v := r.Context().Value(CtxLoggerKey)
	log, ok := v.(*zap.Logger)
	if !ok {
		panic("cannot get derived request id logger from context object")
	}
	return log
}

func writeZap(log *zap.Logger, r *http.Request, t time.Time, status int, size int) {
	elapsed := time.Now().Sub(t)
	var protocol string
	if r.TLS == nil {
		protocol = "http"
	} else {
		protocol = "https"
	}

	v := r.Context().Value(CtxLoggerKey)
	if v != nil {
		var ok bool
		log, ok = v.(*zap.Logger)
		if !ok {
			panic("cannot get derived request id logger from context object")
		}
	}

	// log all the fields
	log.Info(fmt.Sprintf("%s request", protocol),
		zap.Int("status", status),
		zap.String("method", r.Method),
		zap.String("uri", r.RequestURI),
		zap.Bool("tls", r.TLS != nil),
		zap.String("protocol", r.Proto),
		zap.String("host", r.Host),
		zap.String("remote_addr", r.RemoteAddr),
		zap.Int("size", size),
		zap.Duration("elapsed", elapsed),
	)
}

func (z *zapResponseWriter) WriteHeader(code int) {
	z.status = code
	z.ResponseWriter.WriteHeader(code)
}

func (z *zapResponseWriter) Write(b []byte) (int, error) {
	size, err := z.ResponseWriter.Write(b)
	z.size += size
	return size, err
}
