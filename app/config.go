package app

import (
	"time"

	"github.com/nullbio/abcweb/rendering"
	"github.com/nullbio/abcweb/sessions"
	"github.com/pressly/chi"
	"github.com/spf13/cobra"
	"github.com/uber-go/zap"
)

// State is the configuration state for the entire app.
// The controllers are passed variables from this object when initialized.
type State struct {
	Config  *Config
	Log     zap.Logger
	Router  *chi.Mux
	Render  rendering.Renderer
	Root    *cobra.Command
	Session sessions.Overseer
}

// Config for the app
type Config struct {
	// LiveReload enabled or disabled
	LiveReload bool `toml:"live_reload"`
	// Log messages in JSON format
	LogJSON bool `toml:"log_json"`
	// Minimum level to log
	LogLevel string `toml:"log_level"`
	// http bind address. ":<port>" for all interfaces
	Bind string `toml:"bind"`
	// https bind address. ":<port>" for all interfaces
	TLSBind string `toml:"tls_bind"`
	// TLS certificate file path
	TLSCertFile string `toml:"tls_cert_file"`
	// TLS key file path
	TLSKeyFile string `toml:"tls_key_file"`
	// Maximum duration before timing out read of the request
	ReadTimeout time.Duration `toml:"read_timeout"`
	// Maximum duration before timing out write of the response
	WriteTimeout time.Duration `toml:"write_timeout"`
	// Templates folder path
	Templates string `toml:"templates"`
	// Static assets input folder path
	AssetsIn string `toml:"assets_in"`
	// Compiled assets output folder path
	AssetsOut string `toml:"assets_out"`
	// Disable precompilation of assets
	AssetsNoCompile bool `toml:"assets_no_compile"`
	// Disable minification of assets
	AssetsNoMinify bool `toml:"assets_no_minify"`
	// Disable fingerprints in compiled asset filenames
	AssetsNoHash bool `toml:"assets_no_hash"`
	// Disable Gzip compression of asset files
	AssetsNoCompress bool `toml:"assets_no_compress"`
	// Disable browsers caching asset files by setting response headers
	AssetsNoCache bool `toml:"assets_no_cache"`
	// RenderRecompile enables recompilation of the template on every render call.
	// This should be used in development mode so no server restart is required
	// on template file changes.
	RenderRecompile bool `toml:"render_recompile"`
	// Use the development mode sessions storer opposed to production mode storer
	SessionsDevStorer bool `toml:"sessions_dev_storer"`
}

// RegisterFlags registers the configuration flag defaults and help strings
func (s State) RegisterFlags() {
	s.Root.PersistentFlags().BoolP("livereload", "", false, "Enable or disable LiveReload")
	s.Root.PersistentFlags().BoolP("logjson", "", true, "Log messages in JSON format")
	s.Root.PersistentFlags().StringP("loglevel", "", "warn", "Minimum level to log")
	s.Root.PersistentFlags().StringP("bind", "", ":80", `HTTP bind address, eg: ":80"`)
	s.Root.PersistentFlags().StringP("tlsbind", "", "", `HTTPS bind address, eg: ":443"`)
	s.Root.PersistentFlags().StringP("tlscertfile", "", "", "TLS certificate file path")
	s.Root.PersistentFlags().StringP("tlskeyfile", "", "", "TLS key file path")
	s.Root.PersistentFlags().DurationP("readtimeout", "", time.Second*30, "Maximum duration before timing out read of the request")
	s.Root.PersistentFlags().DurationP("writetimeout", "", time.Second*30, "Maximum duration before timing out write of the response")
	s.Root.PersistentFlags().StringP("templates", "", "./compiled_templates", "Templates folder path")
	s.Root.PersistentFlags().StringP("assetsin", "", "./assets", "Static assets input folder path")
	s.Root.PersistentFlags().StringP("assetsout", "", "./public/assets", "Static assets output folder path")
	s.Root.PersistentFlags().BoolP("assetsnocompile", "", false, "Disable precompilation of assets")
	s.Root.PersistentFlags().BoolP("assetsnominify", "", false, "Disable minification of assets")
	s.Root.PersistentFlags().BoolP("assetsnohash", "", false, "Disable fingerprints in compiled asset filenames")
	s.Root.PersistentFlags().BoolP("assetsnocompress", "", false, "Disable gzip compression of asset files")
	s.Root.PersistentFlags().BoolP("assetsnocache", "", false, "Disable browsers caching asset files by setting response headers")
	// This should be used in development mode to avoid having to reload the
	// server on every template file modification.
	s.Root.PersistentFlags().BoolP("renderrecompile", "", false, "Enable recompilation of the template on each render")
	s.Root.PersistentFlags().BoolP("sessionsdevstorer", "", false, "Use the development mode sessions storer")
}

// LoadConfig loads the configuration object
func (s State) LoadConfig() error {
	//shift.LoadConfig(&cfg, )

	s.Config = &Config{}
	return nil
}
