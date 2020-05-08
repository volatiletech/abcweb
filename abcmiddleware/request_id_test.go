package abcmiddleware

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap/zapcore"

	"github.com/go-chi/chi/middleware"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestRequestID(t *testing.T) {
	t.Parallel()

	called := false
	var reqID string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID = middleware.GetReqID(r.Context())
		called = true
	})
	server := middleware.RequestID(RequestIDHeader(handler))

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, r)

	a := assert.New(t)
	a.True(called)
	a.Equal(reqID, w.Header().Get("X-Request-ID"))
}

func TestZapRequestIDLogger(t *testing.T) {
	t.Parallel()

	buf := bufSyncer{new(bytes.Buffer)}
	leveler := zap.NewAtomicLevelAt(zap.InfoLevel)
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(encoder, buf, leveler)
	logger := zap.New(core)

	var reqID string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID = middleware.GetReqID(r.Context())

		reqlogger := Logger(r)
		reqlogger.Info("test")
	})

	reqIDMW := ZapRequestIDLogger(logger)
	server := middleware.RequestID(RequestIDHeader(reqIDMW.Wrap(handler)))

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, r)

	str := buf.String()
	a := assert.New(t)
	a.Contains(str, fmt.Sprintf(`"request_id":"%s"`, reqID))
	a.Contains(str, "test")
}
