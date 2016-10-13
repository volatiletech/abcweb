package main

import (
	"time"

	"github.com/spf13/cobra"
)

type appConfig struct {
	// LiveReload enabled or disabled
	liveReload bool
	// Host name to listen on, blank for all addresses
	host string
	// Port to listen on for HTTP binding
	port int
	// Port to listen on for HTTPS binding
	tlsPort int
	// TLS certificate file path
	tlsCertFile string
	// TLS key file path
	tlsKeyFile string
	// Maximum duration before timing out read of the request
	readTimeout time.Duration
	// Maximum duration before timing out write of the response
	writeTimeout time.Duration
	// Static assets folder path
	staticAssets string
	// Templates folder path
	templates string
	// Minimum level to log
	logLevel string
}

func registerFlags(rootCmd *cobra.Command) {
	rootCmd.PersistentFlags().BoolP("livereload", "", false, "Enable or disable LiveReload")
	rootCmd.PersistentFlags().StringP("bind", "", ":80", `HTTP bind address, eg: ":80"`)
	rootCmd.PersistentFlags().StringP("tlsbind", "", "", `HTTPS bind address, eg: ":443"`)
	rootCmd.PersistentFlags().StringP("tlscertfile", "", "", "TLS certificate file path")
	rootCmd.PersistentFlags().StringP("tlskeyfile", "", "", "TLS key file path")
	rootCmd.PersistentFlags().DurationP("readtimeout", "", time.Second*30, "Maximum duration before timing out read of the request")
	rootCmd.PersistentFlags().DurationP("writetimeout", "", time.Second*30, "Maximum duration before timing out write of the response")
	rootCmd.PersistentFlags().StringP("staticassets", "", "./public/assets", "Static assets folder path")
	rootCmd.PersistentFlags().StringP("templates", "", "./compiled_templates", "Templates folder path")
	rootCmd.PersistentFlags().StringP("loglevel", "", "WARN", "Minimum level to log")
}

func loadConfig() (appConfig, error) {
	//shift.LoadConfig(&cfg, )

	return cfg, nil
}
