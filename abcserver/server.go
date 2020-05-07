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

	"github.com/pkg/errors"
	"github.com/volatiletech/abcweb/abcconfig"
	"go.uber.org/zap"
)

// ServerErrLogger allows us to use the zap.Logger as our http.Server ErrorLog
type ServerErrLogger struct {
	log *zap.Logger
}

// Implement Write to log server errors using the zap logger
func (s ServerErrLogger) Write(b []byte) (int, error) {
	s.log.Debug(string(b))
	return 0, nil
}

// StartServer starts the web server on the specified port, and can be
// gracefully shut down by sending an os.Interrupt signal to the server.
// This is a blocking call.
func StartServer(cfg abcconfig.ServerConfig, router http.Handler, logger *zap.Logger, kill chan struct{}) error {
	errs := make(chan error)

	// These start in goroutines and converge when we kill them
	primary := mainServer(cfg, router, logger, errs)
	var secondary *http.Server

	if len(cfg.TLSBind) != 0 && len(cfg.Bind) != 0 {
		secondary = redirectServer(cfg, logger, errs)
	}

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, os.Kill)

	select {
	case <-kill:
		logger.Info("internal shutdown initiated")
	case sig := <-quit:
		logger.Info("signal received, shutting down", zap.String("signal", sig.String()))
	case err := <-errs:
		logger.Error("error from server, shutting down", zap.Error(err))
	}

	if err := primary.Shutdown(context.Background()); err != nil {
		logger.Error("error shutting down http(s) server", zap.Error(err))
	}
	if secondary != nil {
		if err := secondary.Shutdown(context.Background()); err != nil {
			logger.Error("error shutting down redirector server", zap.Error(err))
		}
	}

	logger.Info("http(s) server shut down complete")
	return nil
}

func mainServer(cfg abcconfig.ServerConfig, router http.Handler, logger *zap.Logger, errs chan<- error) *http.Server {
	server := basicServer(cfg, logger)
	server.Handler = router

	useTLS := len(cfg.TLSBind) != 0

	if !useTLS {
		server.Addr = cfg.Bind

		logger.Info("starting http listener", zap.String("bind", cfg.Bind))
		go func() {
			if err := server.ListenAndServe(); err != nil {
				errs <- errors.Wrap(err, "http listener died")
			}
		}()

		return server
	}

	server.Addr = cfg.TLSBind
	server.TLSConfig = &tls.Config{
		// Causes servers to use Go's default ciphersuite preferences,
		// which are tuned to avoid attacks. Does nothing on clients.
		PreferServerCipherSuites: true,
		// Only use curves which have assembly implementations
		CurvePreferences: []tls.CurveID{tls.CurveP256, tls.X25519},
	}

	logger.Info("starting https listener", zap.String("bind", cfg.TLSBind))
	go func() {
		if err := server.ListenAndServeTLS(cfg.TLSCertFile, cfg.TLSKeyFile); err != nil {
			errs <- errors.Wrap(err, "https listener died")
		}
	}()

	return server
}

func redirectServer(cfg abcconfig.ServerConfig, logger *zap.Logger, errs chan<- error) *http.Server {
	_, httpsPort, err := net.SplitHostPort(cfg.TLSBind)
	if err != nil {
		errs <- errors.Wrap(err, "http listener died")
		return nil
	}

	server := basicServer(cfg, logger)
	server.Addr = cfg.Bind
	server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		httpHost := r.Host
		// Remove port if it exists so we can replace it with https port
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

		logger.Info("redirect", zap.String("remote", r.RemoteAddr), zap.String("host", r.Host), zap.String("path", r.URL.String()), zap.String("redirecturl", url))
		http.Redirect(w, r, url, http.StatusMovedPermanently)
	})

	logger.Info("starting http listener", zap.String("bind", cfg.Bind))
	go func() {
		if err := server.ListenAndServe(); err != nil {
			errs <- errors.Wrap(err, "http listener died")
		}
	}()

	return server
}

func basicServer(cfg abcconfig.ServerConfig, logger *zap.Logger) *http.Server {
	server := &http.Server{
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
		ErrorLog:     log.New(ServerErrLogger{log: logger}, "", 0),
	}

	return server
}
