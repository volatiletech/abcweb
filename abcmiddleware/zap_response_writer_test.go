package abcmiddleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestZapResponseWriter(t *testing.T) {
	w := httptest.NewRecorder()
	zw := zapResponseWriter{
		ResponseWriter: w,
	}

	a := assert.New(t)

	zw.WriteHeader(http.StatusCreated)
	a.Equal(http.StatusCreated, w.Code)

	zw.Write([]byte("123"))
	a.Equal("123", w.Body.String())

	_, _, err := zw.Hijack()
	a.Error(err)

	a.Equal(http.StatusCreated, zw.status)
	a.Equal(3, zw.size)
	a.False(zw.hijacked)
}
