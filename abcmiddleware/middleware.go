package abcmiddleware

import (
	"net/http"
)

// MiddlewareFunc is the function signature for Chi's Use() middleware
type MiddlewareFunc func(http.Handler) http.Handler

// MW is an interface defining middleware wrapping
type MW interface {
	Wrap(http.Handler) http.Handler
}
