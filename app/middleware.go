package app

// initMiddleware enables useful middleware for the router.
import (
	"github.com/goware/cors"
	"github.com/nullbio/abcweb/middleware"
	chimiddleware "github.com/pressly/chi/middleware"
)

// See https://github.com/pressly/chi#middlewares for additional middleware.
func (a AppState) InitMiddleware() {
	m := middleware.Middleware{
		Log: a.Log,
	}

	// Graceful panic recovery that uses zap to log the stack trace
	a.Router.Use(m.Recover)

	// Strip and redirect slashes on routing paths
	a.Router.Use(chimiddleware.StripSlashes)
	// Injects a request ID into the context of each request
	a.Router.Use(chimiddleware.RequestID)

	// Sets response headers to prevent clients from caching
	if a.Config.AssetsNoCache {
		a.Router.Use(chimiddleware.NoCache)
	}

	// Enable CORS.
	// Configuration documentation at: https://godoc.org/github.com/goware/cors
	//
	// Note: If you're getting CORS related errors you may need to adjust the
	// default settings by calling cors.New() with your own cors.Options struct.
	a.Router.Use(cors.Default().Handler)

	// Use zap logger for all routing
	a.Router.Use(m.Zap)

	// More available middleware. Uncomment to enable:
	//
	// Monitoring endpoint to check the servers pulse.
	// route := chimiddleware.Route("/ping")
	// a.router.Use(route)
	//
	// Sets a http.Request's RemoteAddr to either X-Forwarded-For or X-Real-IP
	// a.router.Use(chimiddleware.RealIP)
	//
	// Signals to the request context when a client has closed their connection.
	// It can be used to cancel long operations on the server when the client
	// disconnects before the response is ready.
	// a.router.Use(chimiddleware.CloseNotify)
	//
	// Timeout is a middleware that cancels ctx after a given timeout and return
	// a 504 Gateway Timeout error to the client.
	// Generally readTimeout and writeTimeout is all that is required for timeouts.
	// timeout := chimiddleware.Timeout(time.Second * 30)
	// a.router.Use(timeout)
	//
	// Puts a ceiling on the number of concurrent requests.
	// throttle := chimiddleware.Throttle(100)
	// a.router.Use(throttle)
	//
	// Easily attach net/http/pprof to your routers.
	// a.router.Use(chimiddleware.Profiler)
}
