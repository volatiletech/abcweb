package middleware

import (
	"github.com/volatiletech/abcrender"
	"go.uber.org/zap"
)

type Middleware struct {
	Log       *zap.Logger
	Render    abcrender.Renderer
	ErrorsMap map[string]error
}
