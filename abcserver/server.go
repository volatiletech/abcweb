package abcserver

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"github.com/go-chi/chi"
	"github.com/pkg/errors"
	"github.com/volatiletech/abcweb/abcconfig"
	"go.uber.org/zap"
)

// serverErrLogger allows us to use the zap.Logger as our http.Server ErrorLog
type serverErrLogger struct {
	log *zap.Logger
}

// Implement Write to log server errors using the zap logger
func (s serverErrLogger) Write(b []byte) (int, error) {
	s.log.Debug(string(b))
	return 0, nil
}

// StartServer starts the web server on the specified port, and can be
// gracefully shut down by sending an os.Interrupt signal to the server.
// This is a blocking call.
func StartServer(cfg abcconfig.ServerConfig, router *chi.Mux, logger *zap.Logger) error {
	var err error
	server := http.Server{
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
		ErrorLog:     log.New(serverErrLogger{logger}, "", 0),
		Handler:      router,
	}

	server.TLSConfig = &tls.Config{
		// Causes servers to use Go's default ciphersuite preferences,
		// which are tuned to avoid attacks. Does nothing on clients.
		PreferServerCipherSuites: true,
		// Only use curves which have assembly implementations
		CurvePreferences: []tls.CurveID{
			tls.CurveP256,
			tls.X25519, // Go 1.8 only
		},
	}

	// subscribe to SIGINT signals
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)

	// Start server graceful shutdown goroutine
	go func() {
		<-quit
		log.Println("Shutting down server...")
		if err := server.Shutdown(context.Background()); err != nil {
			log.Fatalf("could not shutdown: %v", err)
		}
	}()

	if len(cfg.TLSBind) > 0 {
		logger.Info("starting https listener", zap.String("bind", cfg.TLSBind))
		server.Addr = cfg.TLSBind

		// Redirect http requests to https
		go Redirect(cfg, logger)

		if err := server.ListenAndServeTLS(cfg.TLSCertFile, cfg.TLSKeyFile); err != nil {
			fmt.Printf("failed to ListenAndServeTLS: %v", err)
			return nil
		}
	} else {
		logger.Info("starting http listener", zap.String("bind", cfg.Bind))
		server.Addr = cfg.Bind
		if err := server.ListenAndServe(); err != nil {
			fmt.Printf("failed to ListenAndServe: %v", err)
			return nil
		}
	}

	return errors.Wrap(err, "failed to StartServer")
}

// Redirect listens on the non-https port, and redirects all requests to https
func Redirect(cfg abcconfig.ServerConfig, logger *zap.Logger) {
	var err error

	// Get https port from TLS Bind
	_, httpsPort, err := net.SplitHostPort(cfg.TLSBind)
	if err != nil {
		log.Fatal("failed to get port from tls bind", zap.Error(err))
	}

	logger.Info("starting http -> https redirect listener", zap.String("bind", cfg.Bind))

	server := http.Server{
		Addr: cfg.Bind,
		// Do not set IdleTimeout, so Go uses ReadTimeout value.
		// IdleTimeout config value too high for redirect listener.
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		ErrorLog:     log.New(serverErrLogger{logger}, "", 0),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Remove port if it exists so we can replace it with https port
			httpHost := r.Host
			if strings.ContainsRune(r.Host, ':') {
				httpHost, _, err = net.SplitHostPort(r.Host)
				if err != nil {
					logger.Error("failed to get http host from request", zap.String("host", r.Host), zap.Error(err))
					w.WriteHeader(http.StatusBadRequest)
					io.WriteString(w, "invalid host header")
					return
				}
			}

			var url string
			if httpsPort != "443" {
				url = fmt.Sprintf("https://%s:%s%s", httpHost, httpsPort, r.RequestURI)
			} else {
				url = fmt.Sprintf("https://%s%s", httpHost, r.RequestURI)
			}

			logger.Info("redirect", zap.String("host", r.Host), zap.String("path", r.URL.String()), zap.String("redirecturl", url))
			http.Redirect(w, r, url, http.StatusMovedPermanently)
		}),
	}

	// Start permanent listener
	err = server.ListenAndServe()
	logger.Fatal("http redirect listener failed", zap.Error(err))
}
