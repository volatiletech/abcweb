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
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

// The configuration flags for the web server.
// The default values are development mode values.
// For production mode all flags will need to be supplied appropriately.
var (
	// The HTTP port to listen on. If TLS is enabled it will redirect to TLS port.
	hostname = kingpin.Flag("hostname", "The domain name, server hostname or IP address.").Default("localhost").Short('h').String()
	port     = kingpin.Flag("port", "The HTTP port to listen on.").Default("3000").Short('p').Int()

	// TLS configuration variables. If --tls=false they won't be used.
	tls      = kingpin.Flag("tls", "Enable TLS & HTTP2 mode.").Default("false").Short('s').Bool()
	tlsport  = kingpin.Flag("tlsport", "The TLS port to listen on.").Default("3001").Int()
	certfile = kingpin.Flag("certfile", "The TLS cert file path.").Default("./server.pem").String()
	keyfile  = kingpin.Flag("keyfile", "The TLS key file path.").Default("./server.key").String()

	assets    = kingpin.Flag("assets", "The static assets path.").Default("./assets").Short('a').String()
	templates = kingpin.Flag("templates", "The dynamic templates path.").Default("./templates").Short('t').String()

	// If no log file is provided, os.Stdout will be used
	logfile  = kingpin.Flag("log", "The optional log file path.").Short('l').String()
	loglevel = kingpin.Flag("level", "The minimum level to log.").Default("INFO").String()
)

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

func main() {
	// Initialize flags
	kingpin.Version("1.0.0")
	kingpin.Parse()

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
