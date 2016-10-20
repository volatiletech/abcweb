package app

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/uber-go/zap"
)

// TO DO
//
// UPDATE
// THE
// THING
// WITH THE
// HTTP.SERVER

// Redirect listens on the non-https port, and redirects all requests to https
func (a AppState) Redirect() {
	var err error

	// Get https port from TLS Bind
	_, httpsPort, err := net.SplitHostPort(a.Config.TLSBind)
	if err != nil {
		a.Log.Fatal("failed to get port from tls bind", zap.Error(err))
	}

	a.Log.Info("starting http -> https redirect listener", zap.String("bind", a.Config.Bind))
	err = http.ListenAndServe(a.Config.Bind, http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// Remove port if it exists so we can replace it with https port
			httpHost, _, err := net.SplitHostPort(r.Host)
			if err != nil {
				a.Log.Fatal("failed to get http host from request", zap.Error(err))
			}

			url := fmt.Sprintf("https://%s:%s%s", httpHost, httpsPort, r.RequestURI)
			http.Redirect(w, r, url, http.StatusMovedPermanently)
		},
	))
	a.Log.Fatal("http redirect listener failed", zap.Error(err))
}

type serverErrLogger struct {
	log zap.Logger
}

func (s serverErrLogger) Write(b []byte) (int, error) {
	s.log.Debug(string(b))
	return 0, nil
}

// StartServer starts the web server on the specified port
func (a AppState) StartServer() error {
	var err error
	server := http.Server{
		ReadTimeout:  a.Config.ReadTimeout,
		WriteTimeout: a.Config.WriteTimeout,
		ErrorLog:     log.New(serverErrLogger{a.Log}, "", 0),
		Handler:      a.Router,
	}

	if len(a.Config.TLSBind) > 0 {
		a.Log.Info("starting https listener", zap.String("bind", a.Config.TLSBind))
		server.Addr = a.Config.TLSBind

		// Redirect http requests to https
		go a.Redirect()
		err = server.ListenAndServeTLS(a.Config.TLSCertFile, a.Config.TLSKeyFile)
	} else {
		a.Log.Info("starting http listener", zap.String("bind", a.Config.Bind))
		server.Addr = a.Config.Bind
		err = server.ListenAndServe()
	}

	return err
}
