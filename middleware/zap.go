package middleware

import "net/http"

func Zap(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

	}

	return http.HandlerFunc(fn)
}
