package main

import (
	"time"

	"github.com/spf13/cobra"
)

type appConfig struct {
	// LiveReload enabled or disabled
	liveReload bool `toml:"live_reload"`
	// Minimum level to log
	logLevel string `toml:"log_level"`
	// Enable log colors
	logColors bool `toml:"log_colors"`
	// http bind address. ":<port>" for all interfaces
	bind string `toml:"bind"`
	// https bind address. ":<port>" for all interfaces
	tlsBind int `toml:"tls_bind"`
	// TLS certificate file path
	tlsCertFile string `toml:"tls_cert_file"`
	// TLS key file path
	tlsKeyFile string `toml:"tls_key_file"`
	// Maximum duration before timing out read of the request
	readTimeout time.Duration `toml:"read_timeout"`
	// Maximum duration before timing out write of the response
	writeTimeout time.Duration `toml:"write_timeout"`
	// Templates folder path
	templates string `toml:"templates"`
	// Static assets input folder path
	assetsIn string `toml:"assets_in"`
	// Compiled assets output folder path
	assetsOut string `toml:"assets_out"`
	// Disable precompilation of assets
	assetsNoCompile bool `toml:"assets_no_compile"`
	// Disable minification of assets
	assetsNoMinify bool `toml:"assets_no_minify"`
	// Disable fingerprints in compiled asset filenames
	assetsNoHash bool `toml:"assets_no_hash"`
	// Disable Gzip compression of asset files
	assetsNoCompress bool `toml:"assets_no_compress"`
	// Disable browsers caching asset files by setting response headers
	assetsNoCache bool `toml:"assets_no_cache"`
	// RenderRecompile enables recompilation of the template on every render call.
	// This should be used in development mode so no server restart is required
	// on template file changes.
	renderRecompile bool `toml:"render_recompile"`
}

func registerFlags(rootCmd *cobra.Command) {
	rootCmd.PersistentFlags().BoolP("livereload", "", false, "Enable or disable LiveReload")
	rootCmd.PersistentFlags().StringP("loglevel", "", "warn", "Minimum level to log")
	rootCmd.PersistentFlags().StringP("bind", "", ":80", `HTTP bind address, eg: ":80"`)
	rootCmd.PersistentFlags().StringP("tlsbind", "", "", `HTTPS bind address, eg: ":443"`)
	rootCmd.PersistentFlags().StringP("tlscertfile", "", "", "TLS certificate file path")
	rootCmd.PersistentFlags().StringP("tlskeyfile", "", "", "TLS key file path")
	rootCmd.PersistentFlags().DurationP("readtimeout", "", time.Second*30, "Maximum duration before timing out read of the request")
	rootCmd.PersistentFlags().DurationP("writetimeout", "", time.Second*30, "Maximum duration before timing out write of the response")
	rootCmd.PersistentFlags().StringP("templates", "", "./compiled_templates", "Templates folder path")
	rootCmd.PersistentFlags().StringP("assetsin", "", "./assets", "Static assets input folder path")
	rootCmd.PersistentFlags().StringP("assetsout", "", "./public/assets", "Static assets output folder path")
	rootCmd.PersistentFlags().BoolP("assetsnocompile", "", false, "Disable precompilation of assets")
	rootCmd.PersistentFlags().BoolP("assetsnominify", "", false, "Disable minification of assets")
	rootCmd.PersistentFlags().BoolP("assetsnohash", "", false, "Disable fingerprints in compiled asset filenames")
	rootCmd.PersistentFlags().BoolP("assetsnocompress", "", false, "Disable gzip compression of asset files")
	rootCmd.PersistentFlags().BoolP("assetsnocache", "", false, "Disable browsers caching asset files by setting response headers")
	// This should be used in development mode to avoid having to reload the
	// server on every template file modification.
	rootCmd.PersistentFlags().BoolP("renderrecompile", "", false, "Enable recompilation of the template on each render")
}

func initConfig() (appConfig, error) {
	//shift.LoadConfig(&cfg, )

	return appConfig{}, nil
}
