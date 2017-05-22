package middleware

import "github.com/volatiletech/abcrender"

// Middleware defines the variables that your middleware need access to.
type Middleware struct {
	Render abcrender.Renderer
}
