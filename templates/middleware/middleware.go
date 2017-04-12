package middleware

import (
	"github.com/volatiletech/abcrender"
	"go.uber.org/zap"
)

// Middleware defines the variables that your middleware need access to.
type Middleware struct {
	Log    *zap.Logger
	Render abcrender.Renderer
}
