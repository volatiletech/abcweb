package middleware

import "github.com/uber-go/zap"

// Middleware exposes useful variables to every middleware handler.
// If you wanted to pass additional objects to each middleware handler
// you can add them here, and then assign them in main.go where the
// instance of middleware is created.
type Middleware struct {
	Log zap.Logger
}
