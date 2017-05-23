package abcmiddleware

import "go.uber.org/zap"

// CtxLoggerKey is the http.Request Context lookup key for the request ID logger
const CtxLoggerKey = "request_id_logger"

// Middleware exposes useful variables to every abcmiddleware handler
type Middleware struct {
	// Log is used for logging in your middleware and to
	// create a derived logger that includes the request ID.
	Log *zap.Logger
}
