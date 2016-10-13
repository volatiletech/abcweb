package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine"
	"github.com/labstack/echo/engine/standard"
	echologrus "github.com/nullbio/echo-logrus"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "{{.AppName}} [flags]",
		Short: "{{.AppName}} web app server",
		RunE:  run,
	}

	registerFlags(rootCmd)

	// Initialize the logger
	initLogger()

	e := echo.New()

	// Log echo activity using echo logrus middleware
	e.Use(echologrus.New())

	// Bind the template renderer
	e.SetRenderer(NewRender())

	// Initialize the routes
	initRoutes(e)

	// Start the web server listener
	if err := startServer(e); err != nil {
		log.Error(err.Error())
	}
}

func run(cmd *cobra.Command, args []string) error {

}

// initLogger initializes Logrus using an optional log file or Stdout.
func initLogger() {
	fmter := &log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05 MST",
	}

	log.SetFormatter(fmter)

	if *logfile == "" {
		log.SetOutput(os.Stdout)
	} else {
		f, err := os.OpenFile(*logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalln("cannot open log file:", err)
		}
		log.SetOutput(f)
	}

	level, err := log.ParseLevel(*loglevel)
	if err != nil {
		log.Fatalln("cannot parse log level:", err)
	}
	log.SetLevel(level)
}

// redirect listens on the non-HTTPS port, and redirects all requests to HTTPS
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
