package abcmiddleware

import (
	"errors"
	"net/http"
	"reflect"

	chimiddleware "github.com/go-chi/chi/middleware"
	"github.com/volatiletech/abcweb/abcrender"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// These are the errors that can be used in your controllers and matched
// against in the Errors middleware to perform special actions.
var (
	ErrUnauthorized = errors.New("not authorized")
	ErrForbidden    = errors.New("access is forbidden")
)

// ErrorContainer holds all of the relevant variables for a users custom error
type ErrorContainer struct {
	// The error that will be returned by the controller
	Err error
	// The template file top render (e.g "errors/500")
	Template string
	// HTTP status code to respond with
	Code int
	// A custom handler to perform additional operations on this error
	Handler ErrorHandler
}

type errorManager struct {
	render abcrender.Renderer
	errors []ErrorContainer
}

// NewErrorManager creates an error manager that can be used to
// create an error handler to wrap your controllers with
func NewErrorManager(render abcrender.Renderer) *errorManager {
	return &errorManager{
		render: render,
		errors: []ErrorContainer{},
	}
}

// NewError creates a new ErrorContainer that can be added to an errorManager.
// If you provide a handler here (instead of nil) then the Errors middleware
// will use your handler opposed to taking the default route of logging
// and rendering. You must handle logging and rendering yourself.
func NewError(err error, code int, template string, handler ErrorHandler) ErrorContainer {
	if err == nil {
		panic("cannot supply nil error")
	}

	// template and code must be set if handler is nil
	if handler == nil && (len(template) == 0 || code == 0) {
		panic("template and code must be set if handler is nil")
	}

	return ErrorContainer{
		Err:      err,
		Code:     code,
		Template: template,
		Handler:  handler,
	}
}

// Remove a ErrorContainer from the error manager
func (m *errorManager) Remove(e ErrorContainer) {
	for i, v := range m.errors {
		if reflect.DeepEqual(v, e) {
			m.errors = append(m.errors[:i], m.errors[i+1:]...)
		}
	}
}

// Add a new ErrorContainer to the error manager
func (m *errorManager) Add(e ErrorContainer) {
	m.errors = append(m.errors, e)
}

// AppHandler is the function signature for controllers that return errors.
type AppHandler func(w http.ResponseWriter, r *http.Request) error

// ErrorHandler is the function signature for user supplied error handlers.
type ErrorHandler func(w http.ResponseWriter, r *http.Request, e ErrorContainer, render abcrender.Renderer) error

// Errors is a middleware to handle controller errors and error page rendering.
// The benefit of using this middleware opposed to logging and rendering
// errors directly in your controller is that it's all centralized to one
// location which simplifies adding notifiers (like slack and email).
// It also reduces a lot of controller boilerplate.
func (m *errorManager) Errors(ctrl AppHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := ctrl(w, r)
		if err == nil {
			return
		}

		var container ErrorContainer
		var template string
		var code int

		found := false
		for _, e := range m.errors {
			if e.Err == err {
				container = e
				found = true
				break
			}
		}

		if !found { // no error containers/handlers found, default path
			code = http.StatusInternalServerError
			template = "errors/500"
		} else if container.Handler != nil { // container and handler are set
			err := container.Handler(w, r, container, m.render)
			if err != nil {
				panic(err)
			}
			// Users handlers should handle EVERYTHING, so return here
			// once the handler has been called successfully above.
			return
		} else { // container is set and handler is nil
			code = container.Code
			template = container.Template
		}

		// Get the Request ID scoped logger
		log := Log(r)

		fields := []zapcore.Field{
			zap.String("method", r.Method),
			zap.String("uri", r.RequestURI),
			zap.Bool("tls", r.TLS != nil),
			zap.String("protocol", r.Proto),
			zap.String("host", r.Host),
			zap.String("remote_addr", r.RemoteAddr),
			zap.Error(err),
		}

		// log with the request_id scoped logger
		switch code {
		case http.StatusInternalServerError:
			log.Error("request error", fields...)
			requestID := chimiddleware.GetReqID(r.Context())
			m.render.HTML(w, code, template, requestID)
		default: // warn does not log stacktrace in prod, but error and above does
			log.Warn("request failed", fields...)
			m.render.HTML(w, code, template, nil)
		}
	}
}
