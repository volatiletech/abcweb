package middleware

import (
	"bytes"
	"fmt"
	"net/http"

	chimiddleware "github.com/pressly/chi/middleware"
)

func (m Middleware) Recover(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); r != nil {
				recover()
				buf := &bytes.Buffer{}
				requestID := chimiddleware.GetReqID(r.Context())

				if requestID != "" {
					fmt.Fprintf(buf, "[%s] ", requestID)
				}

				fmt.Fprintf(buf, `"%s `, r.Method)

				if r.TLS == nil {
					buf.WriteString(`http`)
				} else {
					buf.WriteString(`https`)
				}

				fmt.Fprintf(buf, "://%s%s %s\" from %s -- panic:\n%+v", r.Host, r.RequestURI, r.Proto, r.RemoteAddr, err)

				m.Log.Error(buf.String())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
