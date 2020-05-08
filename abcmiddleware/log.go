package abcmiddleware

import (
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type zapLogMiddleware struct {
	logger *zap.Logger
}

// ZapLog returns a logging middleware that outputs details about a request
func ZapLog(logger *zap.Logger) MW {
	return zapLogMiddleware{logger: logger}
}

// Zap middleware handles web request logging using Zap
func (z zapLogMiddleware) Wrap(next http.Handler) http.Handler {
	return zapLogger{mid: z, next: next}
}

type zapLogger struct {
	mid  zapLogMiddleware
	next http.Handler
}

func (z zapLogger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	zw := &zapResponseWriter{ResponseWriter: w}

	// Serve the request
	z.next.ServeHTTP(zw, r)

	// Write the request log line
	z.writeZap(zw, r, startTime)
}

func (z zapLogger) writeZap(zw *zapResponseWriter, r *http.Request, startTime time.Time) {
	elapsed := time.Now().Sub(startTime)
	var protocol string
	if r.TLS == nil {
		protocol = "http"
	} else {
		protocol = "https"
	}

	logger := z.mid.logger
	v := r.Context().Value(CTXKeyLogger)
	if v != nil {
		var ok bool
		logger, ok = v.(*zap.Logger)
		if !ok {
			panic("cannot get derived request id logger from context object")
		}
	}

	// log all the fields
	logger.Info(fmt.Sprintf("%s request", protocol),
		zap.Int("status", zw.status),
		zap.Int("size", zw.size),
		zap.Bool("hijacked", zw.hijacked),
		zap.String("method", r.Method),
		zap.String("uri", r.RequestURI),
		zap.Bool("tls", r.TLS != nil),
		zap.String("protocol", r.Proto),
		zap.String("host", r.Host),
		zap.String("remote_addr", r.RemoteAddr),
		zap.Duration("elapsed", elapsed),
	)
}
