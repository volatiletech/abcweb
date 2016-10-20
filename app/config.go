package app

import (
	"time"

	"github.com/pressly/chi"
	"github.com/spf13/cobra"
	"github.com/uber-go/zap"
	"github.com/unrolled/render"
)

type AppState struct {
	Config *AppConfig
	Log    zap.Logger
	Router *chi.Mux
	Render *render.Render
	Root   *cobra.Command
}

type AppConfig struct {
	// LiveReload enabled or disabled
	LiveReload bool `toml:"live_reload"`
	// Log messages in JSON format
	LogJSON bool `toml:"log_JSON"`
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
}

func (a AppState) RegisterFlags() {
	a.Root.PersistentFlags().BoolP("livereload", "", false, "Enable or disable LiveReload")
	a.Root.PersistentFlags().BoolP("logjson", "", true, "Log messages in JSON format")
	a.Root.PersistentFlags().StringP("loglevel", "", "warn", "Minimum level to log")
	a.Root.PersistentFlags().StringP("bind", "", ":80", `HTTP bind address, eg: ":80"`)
	a.Root.PersistentFlags().StringP("tlsbind", "", "", `HTTPS bind address, eg: ":443"`)
	a.Root.PersistentFlags().StringP("tlscertfile", "", "", "TLS certificate file path")
	a.Root.PersistentFlags().StringP("tlskeyfile", "", "", "TLS key file path")
	a.Root.PersistentFlags().DurationP("readtimeout", "", time.Second*30, "Maximum duration before timing out read of the request")
	a.Root.PersistentFlags().DurationP("writetimeout", "", time.Second*30, "Maximum duration before timing out write of the response")
	a.Root.PersistentFlags().StringP("templates", "", "./compiled_templates", "Templates folder path")
	a.Root.PersistentFlags().StringP("assetsin", "", "./assets", "Static assets input folder path")
	a.Root.PersistentFlags().StringP("assetsout", "", "./public/assets", "Static assets output folder path")
	a.Root.PersistentFlags().BoolP("assetsnocompile", "", false, "Disable precompilation of assets")
	a.Root.PersistentFlags().BoolP("assetsnominify", "", false, "Disable minification of assets")
	a.Root.PersistentFlags().BoolP("assetsnohash", "", false, "Disable fingerprints in compiled asset filenames")
	a.Root.PersistentFlags().BoolP("assetsnocompress", "", false, "Disable gzip compression of asset files")
	a.Root.PersistentFlags().BoolP("assetsnocache", "", false, "Disable browsers caching asset files by setting response headers")
	// This should be used in development mode to avoid having to reload the
	// server on every template file modification.
	a.Root.PersistentFlags().BoolP("renderrecompile", "", false, "Enable recompilation of the template on each render")
}

func (a AppState) LoadConfig() error {
	//shift.LoadConfig(&cfg, )

	a.Config = &AppConfig{}
	return nil
}
