package abcmiddleware

import (
	"bufio"
	"net"
	"net/http"

	"github.com/pkg/errors"
)

// zapResponseWriter is a wrapper that includes that http status and size for logging
type zapResponseWriter struct {
	http.ResponseWriter
	status   int
	size     int
	hijacked bool
}

func (z *zapResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := z.ResponseWriter.(http.Hijacker); ok {
		z.hijacked = true
		return hijacker.Hijack()
	}
	return nil, nil, errors.Errorf("%T does not support http hijacking", z.ResponseWriter)
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
