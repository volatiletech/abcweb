package abcmiddleware

import (
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

// ZapRecover will attempt to log a panic, as well as produce a reasonable
// error for the client by calling the passed in errorHandler function.
//
// It uses the zap logger and attempts to look up a request-scoped logger
// created with this package before using the passed in logger.
//
// The zap logger that's used here should be careful to enable stacktrace
// logging for any levels that they require it for.
func ZapRecover(fallback *zap.Logger, errorHandler http.HandlerFunc) MW {
	return zapRecoverMiddleware{
		fallback: fallback,
		eh:       errorHandler,
	}
}

type zapRecoverMiddleware struct {
	fallback *zap.Logger
	eh       http.HandlerFunc
}

func (z zapRecoverMiddleware) Wrap(next http.Handler) http.Handler {
	return zapRecoverer{
		zr:   z,
		next: next,
	}
}

type zapRecoverer struct {
	zr   zapRecoverMiddleware
	next http.Handler
}

// recoverPanic was mostly adapted from abcweb
func (z zapRecoverer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer z.recoverNicely(w, r)
	z.next.ServeHTTP(w, r)
}

func (z zapRecoverer) recoverNicely(w http.ResponseWriter, r *http.Request) {
	err := recover()
	if err == nil {
		return
	}

	var protocol string
	if r.TLS == nil {
		protocol = "http"
	} else {
		protocol = "https"
	}

	if z.zr.eh != nil {
		z.zr.eh(w, r)
	}

	logger := z.zr.fallback
	v := r.Context().Value(CTXKeyLogger)
	if v != nil {
		var ok bool
		logger, ok = v.(*zap.Logger)
		if !ok {
			panic("cannot get derived request id logger from context object")
		}
	}

	logger.Error(fmt.Sprintf("%s request error", protocol),
		zap.String("method", r.Method),
		zap.String("uri", r.RequestURI),
		zap.Bool("tls", r.TLS != nil),
		zap.String("protocol", r.Proto),
		zap.String("host", r.Host),
		zap.String("remote_addr", r.RemoteAddr),
		zap.String("panic", fmt.Sprintf("%+v", err)),
		zap.Stack("stacktrace"),
	)
}
