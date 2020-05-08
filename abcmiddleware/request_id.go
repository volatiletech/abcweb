package abcmiddleware

import (
	"context"
	"net/http"

	chimiddleware "github.com/go-chi/chi/middleware"
	"go.uber.org/zap"
)

type ctxKey int

const (
	// CTXKeyLogger is the key under which the request scoped logger is placed
	CTXKeyLogger ctxKey = iota
)

// RequestIDHeader sets the X-Request-ID header to the chi request id
// This must be used after the chi request id middleware.
func RequestIDHeader(next http.Handler) http.Handler {
	return reqIDInserter{next: next}
}

type reqIDInserter struct {
	next http.Handler
}

func (re reqIDInserter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if reqID := chimiddleware.GetReqID(r.Context()); len(reqID) != 0 {
		w.Header().Set("X-Request-ID", reqID)
	}
	re.next.ServeHTTP(w, r)
}

// ZapRequestIDLogger returns a request id logger middleware. This only works
// if chi has inserted a request id into the stack first.
func ZapRequestIDLogger(logger *zap.Logger) MW {
	return zapReqLoggerMiddleware{logger: logger}
}

type zapReqLoggerMiddleware struct {
	logger *zap.Logger
}

func (z zapReqLoggerMiddleware) Wrap(next http.Handler) http.Handler {
	return zapReqLoggerInserter{logger: z.logger, next: next}
}

type zapReqLoggerInserter struct {
	logger *zap.Logger
	next   http.Handler
}

func (z zapReqLoggerInserter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := chimiddleware.GetReqID(r.Context())

	derivedLogger := z.logger.With(zap.String("request_id", requestID))

	r = r.WithContext(context.WithValue(r.Context(), CTXKeyLogger, derivedLogger))
	z.next.ServeHTTP(w, r)
}

// Logger returns the Request ID scoped logger from the request Context
// and panics if it cannot be found. This function is only ever used
// by your controllers if your app uses the RequestID middlewares,
// otherwise you should use the controller's receiver logger directly.
func Logger(r *http.Request) *zap.Logger {
	return LoggerCTX(r.Context())
}

// LoggerCTX retrieves a logger from a context.
func LoggerCTX(ctx context.Context) *zap.Logger {
	v := ctx.Value(CTXKeyLogger)
	log, ok := v.(*zap.Logger)
	if !ok {
		panic("cannot get derived request id logger from context object")
	}
	return log
}
