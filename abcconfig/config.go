package abcconfig

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/kat-co/vala"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var filename = "config.toml"

// AppConfig struct includes the necessary abcweb config components.
// If you'd rather use your own struct so that you can add new values
// to your configuration you can do so, but make sure you include
// *ServerConfig objects and *DBConfig objects like below (if desired).
type AppConfig struct {
	Server *ServerConfig `toml:"server" mapstructure:"server"`
	DB     *DBConfig     `toml:"db" mapstructure:"db"`
}

// ServerConfig is config for the app loaded through environment variables,
// command line, or the config.toml file.
type ServerConfig struct {
	// The active config file environment mode.
	// This sits at the root level of the config file (above the envs).
	ActiveEnv string `toml:"env" mapstructure:"env"`
	// LiveReload enabled or disabled
	LiveReload bool `toml:"live-reload" mapstructure:"live-reload"`
	// Use the production logger (JSON and log level warn) or the
	// development logger (console and log level info)
	ProdLogger bool `toml:"prod-logger" mapstructure:"prod-logger"`
	// http bind address. ":<port>" for all interfaces
	Bind string `toml:"bind" mapstructure:"bind"`
	// https bind address. ":<port>" for all interfaces
	TLSBind string `toml:"tls-bind" mapstructure:"tls-bind"`
	// TLS certificate file path
	TLSCertFile string `toml:"tls-cert-file" mapstructure:"tls-cert-file"`
	// TLS key file path
	TLSKeyFile string `toml:"tls-key-file" mapstructure:"tls-key-file"`
	// Maximum duration before timing out read of the request
	ReadTimeout time.Duration `toml:"read-timeout" mapstructure:"read-timeout"`
	// Maximum duration before timing out write of the response
	WriteTimeout time.Duration `toml:"write-timeout" mapstructure:"write-timeout"`
	// Maximum duration before timing out idle keep-alive connection
	IdleTimeout time.Duration `toml:"idle-timeout" mapstructure:"idle-timeout"`
	// Use manifest.json assets mapping
	AssetsManifest bool `toml:"assets-manifest" mapstructure:"assets-manifest"`
	// Disable browsers caching asset files by setting response headers
	AssetsNoCache bool `toml:"assets-no-cache" mapstructure:"assets-no-cache"`
	// RenderRecompile enables recompilation of the template on every render call.
	// This should be used in development mode so no server restart is required
	// on template file changes.
	RenderRecompile bool `toml:"render-recompile" mapstructure:"render-recompile"`
	// Use the development mode sessions storer opposed to production mode storer
	// defined in app/sessions.go -- Usually a cookie storer for dev
	// and disk storer for prod.
	SessionsDevStorer bool `toml:"sessions-dev-storer" mapstructure:"sessions-dev-storer"`
	// PublicPath defaults to "public" but can be set to something else
	// by the {{.AppEnvName}}_PUBLIC_PATH environment variable.
	// This is set by the "abcweb dev" command to instruct the app to
	// load assets from a /tmp folder instead of the local public folder.
	PublicPath string `toml:"public-path" mapstructure:"public-path"`
}

// DBConfig holds the database config for the app loaded through
// environment variables, or the config.toml file.
type DBConfig struct {
	// DB is the database software; "postgres", "mysql", etc.
	DB string `toml:"db" mapstructure:"db"`
	// The database name
	DBName  string `toml:"dbname" mapstructure:"dbname"`
	Host    string `toml:"host" mapstructure:"host"`
	Port    int    `toml:"port" mapstructure:"port"`
	User    string `toml:"user" mapstructure:"user"`
	Pass    string `toml:"pass" mapstructure:"pass"`
	SSLMode string `toml:"sslmode" mapstructure:"sslmode"`

	// Throw an error when the app starts if the database is not
	// using the latest migration
	EnforceMigration bool `toml:"enforce-migration" mapstructure:"enforce-migration"`
}

// NewAppConfig returns an initialized AppConfig object
func NewAppConfig() *AppConfig {
	return &AppConfig{
		// Server is not optional, so should be initialized.
		// DB *is* optional, so it can be nil and thus is not initialized.
		Server: &ServerConfig{},
	}
}

// InitAppConfig binds your passed in config flags to a new viper
// instance, retrieves the active environment section of your config file using
// that viper instance, and then loads your server and db config into
// the passed in cfg struct and validates the db config is set appropriately.
func InitAppConfig(flags *pflag.FlagSet, cfg interface{}) error {
	v, err := NewSubViper(flags, filename)
	if err != nil {
		return err
	}

	if err := LoadAppConfig(cfg, v); err != nil {
		return err
	}

	val := reflect.Indirect(reflect.ValueOf(cfg))

	// Check if there's a DBConfig object in the cfg struct.
	// If there is one, check if there's a [db] section in the config
	// file by checking if dbCfg is nil. If there is a [db] section
	// then validate all fields on it are set appropriately.
	for i := 0; i < val.NumField(); i++ {
		dbCfg, ok := val.Field(i).Interface().(*DBConfig)
		if !ok {
			continue
		}
		if dbCfg != nil {
			if err := ValidateDBConfig(dbCfg); err != nil {
				return err
			}
		}
		break
	}

	return nil
}

// NewFlagSet creates the set of flags specific to the server config and
// other relevant config (like enforce-migration etc).
func NewFlagSet() *pflag.FlagSet {
	flags := &pflag.FlagSet{}

	flags.BoolP("live-reload", "", false, "Enable or disable LiveReload")
	flags.BoolP("prod-logger", "", true, "Use the production logger, JSON and log level warn")
	flags.StringP("bind", "", ":80", `HTTP bind address, eg: ":80"`)
	flags.StringP("tls-bind", "", "", `HTTPS bind address, eg: ":443"`)
	flags.StringP("tls-cert-file", "", "", "TLS certificate file path")
	flags.StringP("tls-key-file", "", "", "TLS key file path")
	flags.DurationP("read-timeout", "", time.Second*10, "Maximum duration before timing out read of the request")
	flags.DurationP("write-timeout", "", time.Second*15, "Maximum duration before timing out write of the response")
	flags.DurationP("idle-timeout", "", time.Second*120, "Maximum duration before timing out idle keep-alive connection")

	// manifest.json is created as a part of the gulp production "build" task,
	// it maps fingerprinted asset names to regular asset names, for example:
	// {"js/main.css": "js/e2a3ff9-main.css"}.
	// This should only be set to true if doing asset fingerprinting.
	flags.BoolP("assets-manifest", "", true, "Use manifest.json for mapping asset names to fingerprinted assets")

	// This should be used in development mode to prevent browser caching of assets
	flags.BoolP("assets-no-cache", "", false, "Disable browsers caching asset files by setting response headers")
	// This should be used in development mode to avoid having to reload the
	// server on every template file modification.
	flags.BoolP("render-recompile", "", false, "Enable recompilation of the template on each render")
	// Defined in app/sessions.go -- Usually cookie storer for dev and disk storer for prod.
	flags.BoolP("sessions-dev-storer", "", false, "Use the development mode sessions storer (defined in app/sessions.go)")
	flags.BoolP("enforce-migration", "", true, "Throw error on app start if database is not using latest migration")
	flags.BoolP("version", "", false, "Display the build version hash")
	flags.StringP("env", "e", "prod", "The config files environment to load")

	return flags
}

// NewSubViper returns a viper instance activated against the active environment
// configuration subsection and initialized with the config.toml
// configuration file and the environment variable prefix.
func NewSubViper(flags *pflag.FlagSet, path string) (*viper.Viper, error) {
	v := viper.New()

	if flags != nil {
		if err := v.BindPFlags(flags); err != nil {
			return nil, err
		}
	}

	if err := ConfigureViper(v, path); err != nil {
		return nil, err
	}

	env := v.GetString("env")

	v = v.Sub(env)
	if v == nil {
		return nil, fmt.Errorf("unable to load environment %q from %q", env, path)
	}

	if flags != nil {
		if err := v.BindPFlags(flags); err != nil {
			return nil, err
		}
	}

	v.Set("env", env)
	return v, nil
}

// ConfigureViper sets the viper object to use the passed in config toml file
// and also configures the configuration environment variables.
func ConfigureViper(v *viper.Viper, path string) error {
	v.SetConfigType("toml")
	v.SetConfigFile(path)
	v.SetEnvPrefix("THINGYMABOB")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	if err := v.ReadInConfig(); err != nil {
		return err
	}
	v.AutomaticEnv()

	return nil
}

// LoadServerConfig loads the config.toml server configuration object
func LoadServerConfig(cfg interface{}, v *viper.Viper) error {
	val := reflect.Indirect(reflect.ValueOf(cfg))

	env := v.GetString("env")

	vs := v.Sub("server")
	vs.Unmarshal(cfg)
	return nil

	// Find *ServerConfig and set object appropriately
	for i := 0; i < val.NumField(); i++ {
		serverCfg, ok := val.Field(i).Interface().(*ServerConfig)
		if !ok {
			continue
		}

		sv := v.Sub("server")
		// sv.BindPFlags(v.)
		v.Debug()
		if sv == nil {
			return errors.New("unable to find \"server\" subsection in config")
		}
		fmt.Printf("BIND BEFORE: %s\n", v.GetString("bind"))
		fmt.Printf("BIND AFTER: %s\n", sv.GetString("bind"))

		serverCfg.ActiveEnv = env
		serverCfg.LiveReload = sv.GetBool("live-reload")
		serverCfg.ProdLogger = sv.GetBool("prod-logger")
		serverCfg.Bind = sv.GetString("bind")
		serverCfg.TLSBind = sv.GetString("tls-bind")
		serverCfg.TLSCertFile = sv.GetString("tls-cert-file")
		serverCfg.TLSKeyFile = sv.GetString("tls-key-file")
		serverCfg.ReadTimeout = sv.GetDuration("read-timeout")
		serverCfg.WriteTimeout = sv.GetDuration("write-timeout")
		serverCfg.IdleTimeout = sv.GetDuration("idle-timeout")
		serverCfg.AssetsManifest = sv.GetBool("assets-manifest")
		serverCfg.AssetsNoCache = sv.GetBool("assets-no-cache")
		serverCfg.RenderRecompile = sv.GetBool("render-recompile")
		serverCfg.SessionsDevStorer = sv.GetBool("sessions-dev-storer")
		serverCfg.PublicPath = sv.GetString("public-path")

		// Finished working on the server cfg struct, so break out
		break
	}

	return nil
}

// LoadDBConfig loads the config.toml db configuration object
func LoadDBConfig(cfg interface{}, v *viper.Viper) error {
	val := reflect.Indirect(reflect.ValueOf(cfg))

	env := v.GetString("env")

	vs := v.Sub("server")
	vs.Unmarshal(cfg)
	return nil

	// Find *DBConfig and set object appropriately
	for i := 0; i < val.NumField(); i++ {
		dbCfg, ok := val.Field(i).Interface().(*DBConfig)
		if !ok {
			continue
		}

		// if sv is nil it means that there was no [db] section in the toml
		// file, and so db loading should be skipped
		sv := v.Sub("db")
		if sv == nil {
			break
		}

		dbCfg.DB = sv.GetString("db")
		dbCfg.DBName = sv.GetString("dbname")
		dbCfg.EnforceMigration = sv.GetBool("enforce-migration")
		dbCfg.Host = sv.GetString("host")
		dbCfg.Pass = sv.GetString("pass")
		dbCfg.Port = sv.GetInt("port")
		dbCfg.SSLMode = sv.GetString("sslmode")
		dbCfg.User = sv.GetString("user")

		if dbCfg.DB == "postgres" {
			if dbCfg.Port == 0 {
				dbCfg.Port = 5432
			}
			if dbCfg.SSLMode == "" {
				dbCfg.SSLMode = "require"
			}
		} else if dbCfg.DB == "mysql" {
			if dbCfg.Port == 0 {
				dbCfg.Port = 3306
			}
			if dbCfg.SSLMode == "" {
				dbCfg.SSLMode = "true"
			}
		}

		// Finished working on the db cfg struct, so break out
		break
	}

	return nil
}

// ValidateDBConfig returns an error if any of the required db config
// fields are not set to their appropriate values.
func ValidateDBConfig(cfg *DBConfig) error {

	err := vala.BeginValidation().Validate(
		vala.StringNotEmpty(cfg.DB, "db"),
		vala.StringNotEmpty(cfg.User, "user"),
		vala.StringNotEmpty(cfg.Host, "host"),
		vala.Not(vala.Equals(cfg.Port, 0, "port")),
		vala.StringNotEmpty(cfg.DBName, "dbname"),
		vala.StringNotEmpty(cfg.SSLMode, "sslmode"),
	).Check()
	if err != nil {
		return err
	}

	if cfg.DB != "postgres" && cfg.DB != "mysql" {
		return errors.New("not a valid driver name")
	}

	return nil
}
