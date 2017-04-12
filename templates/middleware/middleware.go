package middleware

import (
	"github.com/volatiletech/abcrender"
	"go.uber.org/zap"
)

// Middleware defines the variables that your middleware need access to.
type Middleware struct {
	Log    *zap.Logger
	Render abcrender.Renderer
	// ErrorsMap is a map of errors used in your controllers and
	// validated in the Errors middleware.
	ErrorsMap map[string]error
}
