package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/aarondl/zapcolors"
	"github.com/goware/cors"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine"
	"github.com/labstack/echo/engine/standard"
	"github.com/nullbio/abcweb/middleware"
	"github.com/pressly/chi"
	chimiddleware "github.com/pressly/chi/middleware"
	"github.com/spf13/cobra"
	"github.com/uber-go/zap"
)

func main() {
	// Register the command-line configuration flags (config.go)
	registerFlags(rootCmd)

	// Load the app configuration (config.go)
	config, err := initConfig()
	if err != nil {
		return err
	}

	var log zap.Logger
	// Initialize the zap logger
	if config.logColors {
		// Enable colored logging for development
		log = zap.New(zapcolors.NewColorEncoder(zapcolors.TextTimeFormat("2006-01-02 15:04:05 MST")))
	} else {
		// JSON logging for production, should be coupled with a log analyzer
		// like newrelic, elk, logstash etc.
		log = zap.New(zap.NewJSONEncoder(zap.TextTimeFormat("2006-01-02 15:04:05 MST")))
	}
	// Set the minimum log level, defined in the app configuration
	log.SetLevel(getLevel(config.logLevel))

	rootCmd := &cobra.Command{
		Use:   "{{.AppName}} [flags]",
		Short: "{{.AppName}} web app server",

		// RunE bootstraps the webserver by performing the following steps:
		//
		// 1. Create a new router
		// 2. Hook the middleware up to the router
		// 3. Initialize the renderer configuration and template helper funcs
		// 4. Initialize the controller routes on the router object
		// 5. Start the web server listeners (http and/or https)
		RunE: func(cmd *cobra.Command, args []string) error {

			// Create a new router
			r := chi.NewRouter()

			// Enable middleware for the router
			initMiddleware(r, config)

			// Configure the renderer (render.go)
			rend := initRenderer(config)

			// Initialize the routes with the renderer (routes.go)
			initRoutes(r, rend)

			// Start the web server listener
			if err := startServer(e); err != nil {
				logger.Error(err.Error())
			}
		},
	}

	// Initialize config, router and start server
	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
	}
}

// getLevel returns the zap.Level for the passed in level string
func getLevel(level string) zap.Level {
	switch strings.ToLower(config.logLevel) {
	case "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	case "panic":
		return zap.PanicLevel
	case "fatal":
		return zap.FatalLevel
	default:
		return zap.WarnLevel
	}
}

// initMiddleware enables useful middleware for the router.
// See https://github.com/pressly/chi#middlewares for additional middleware.
func initMiddleware(r *chi.Mux, c appConfig) {
	// Graceful panic recovery that uses zap to log the stack trace
	r.Use(middleware.Recoverer)
	// Use zap logger for all routing
	r.Use(middleware.Zap)

	// Strip and redirect slashes on routing paths
	r.Use(chimiddleware.StripSlashes)
	// Injects a request ID into the context of each request
	r.Use(chimiddleware.RequestID)

	// Sets response headers to prevent clients from caching
	if c.assetsNoCache {
		r.Use(chimiddleware.NoCache)
	}

	// Enable CORS.
	// Configuration documentation at: https://godoc.org/github.com/goware/cors
	//
	// Note: If you're getting CORS related errors you may need to adjust the
	// default settings by calling cors.New() with your own cors.Options struct.
	r.Use(cors.Default().Handler)

	// More available middleware. Uncomment to enable:
	//
	// Monitoring endpoint to check the servers pulse.
	// route := chimiddleware.Route("/ping")
	// r.Use(route)
	//
	// Sets a http.Request's RemoteAddr to either X-Forwarded-For or X-Real-IP
	// r.Use(chimiddleware.RealIP)
	//
	// Signals to the request context when a client has closed their connection.
	// It can be used to cancel long operations on the server when the client
	// disconnects before the response is ready.
	// r.Use(chimiddleware.CloseNotify)
	//
	// Timeout is a middleware that cancels ctx after a given timeout and return
	// a 504 Gateway Timeout error to the client.
	// Generally readTimeout and writeTimeout is all that is required for timeouts.
	// timeout := chimiddleware.Timeout(time.Second * 30)
	// r.Use(timeout)
	//
	// Puts a ceiling on the number of concurrent requests.
	// throttle := chimiddleware.Throttle(100)
	// r.Use(throttle)
	//
	// Easily attach net/http/pprof to your routers.
	// r.Use(chimiddleware.Profiler)

	return
}

// redirect listens on the non-https port, and redirects all requests to https
func redirect(httpAddress string) {
	log.Infoln("starting http redirect listener on port:", *port)
	err := http.ListenAndServe(httpAddress, http.HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {
			url := fmt.Sprintf("https://%s:%d%s", *hostname, *tlsport, req.RequestURI)
			http.Redirect(w, req, url, http.StatusMovedPermanently)
		},
	))
	log.Fatalln("http redirect listener failed:", err)
}

// startServer starts the web server on the specified port
func startServer(e *echo.Echo) error {
	var server *standard.Server
	httpAddress := *hostname + ":" + strconv.Itoa(*port)

	if *tls {
		log.Infoln("starting server in tls mode on port:", *tlsport)
		httpsAddress := *hostname + ":" + strconv.Itoa(*tlsport)
		server = standard.WithConfig(engine.Config{
			Address:     httpsAddress,
			TLSCertFile: *certfile,
			TLSKeyFile:  *keyfile,
		})

		// Redirect http requests to https
		go redirect(httpAddress)
	} else {
		log.Infoln("starting server on port:", *port)
		server = standard.New(httpAddress)
	}

	if err := e.Run(server); err != nil {
		return err
	}

	return nil
}
