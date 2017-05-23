package abcmiddleware

import (
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

// Recover middleware recovers panics that occur and gracefully logs their error
func (m Middleware) Recover(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				var protocol string
				if r.TLS == nil {
					protocol = "http"
				} else {
					protocol = "https"
				}

				var log *zap.Logger
				v := r.Context().Value(CtxLoggerKey)
				if v != nil {
					var ok bool
					log, ok = v.(*zap.Logger)
					if !ok {
						panic("cannot get derived request id logger from context object")
					}
					// log with the request_id scoped logger
					log.Error(fmt.Sprintf("%s request error", protocol),
						zap.String("method", r.Method),
						zap.String("uri", r.RequestURI),
						zap.Bool("tls", r.TLS != nil),
						zap.String("protocol", r.Proto),
						zap.String("host", r.Host),
						zap.String("remote_addr", r.RemoteAddr),
						zap.String("error", fmt.Sprintf("%+v", err)),
					)
				} else {
					// log with the logger attached to middleware struct if
					// cannot find request_id scoped logger
					m.Log.Error(fmt.Sprintf("%s request error", protocol),
						zap.String("method", r.Method),
						zap.String("uri", r.RequestURI),
						zap.Bool("tls", r.TLS != nil),
						zap.String("protocol", r.Proto),
						zap.String("host", r.Host),
						zap.String("remote_addr", r.RemoteAddr),
						zap.String("error", fmt.Sprintf("%+v", err)),
					)
				}

				// Return a http 500 with the HTTP body of "Internal Server Error"
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
