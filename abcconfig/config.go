package abcconfig

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/kat-co/vala"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Config object used to initialize configuration
type Config struct {
	// The config file path, overwritten in tests to point to a tmp file
	File string
	// Specify which environment to load, empty string means pull the
	// env from the configuration file, cmdline and env vars.
	LoadEnv string
	// Prefix the environment variables with this name so that the config
	// variables don't conflict with other abcweb apps
	EnvPrefix string
}

// NewConfig creates a new Config object used to initialize configuration
func NewConfig(envPrefix string) *Config {
	return &Config{
		File:      "config.toml",
		EnvPrefix: envPrefix,
	}
}

// AppConfig struct includes the necessary abcweb config components.
// If you'd rather use your own struct so that you can add new values
// to your configuration you can do so, but make sure you include
// *ServerConfig objects and *DBConfig objects like below (if desired).
//
// If you do not wish to use a database then you can exclude the DBConfig
// struct in your own struct, but if using this AppConfig struct then
// DBConfig MUST be initialized and database configuration must be present.
type AppConfig struct {
	// The active environment section
	Env string `toml:"env" mapstructure:"env" env:"ENV"`

	Server ServerConfig `toml:"server" mapstructure:"server"`
	DB     DBConfig     `toml:"db" mapstructure:"db"`
}

// ServerConfig is config for the app loaded through environment variables,
// command line, or the config.toml file.
type ServerConfig struct {
	// LiveReload enabled or disabled
	LiveReload bool `toml:"live-reload" mapstructure:"live-reload" env:"SERVER_LIVE_RELOAD"`
	// Use the production logger (JSON and log level warn) or the
	// development logger (console and log level info)
	ProdLogger bool `toml:"prod-logger" mapstructure:"prod-logger" env:"SERVER_PROD_LOGGER"`
	// http bind address. ":<port>" for all interfaces
	Bind string `toml:"bind" mapstructure:"bind" env:"SERVER_BIND"`
	// https bind address. ":<port>" for all interfaces
	TLSBind string `toml:"tls-bind" mapstructure:"tls-bind" env:"SERVER_TLS_BIND"`
	// TLS certificate file path
	TLSCertFile string `toml:"tls-cert-file" mapstructure:"tls-cert-file" env:"SERVER_TLS_CERT_FILE"`
	// TLS key file path
	TLSKeyFile string `toml:"tls-key-file" mapstructure:"tls-key-file" env:"SERVER_TLS_KEY_FILE"`
	// Maximum duration before timing out read of the request
	ReadTimeout time.Duration `toml:"read-timeout" mapstructure:"read-timeout" env:"SERVER_READ_TIMEOUT"`
	// Maximum duration before timing out write of the response
	WriteTimeout time.Duration `toml:"write-timeout" mapstructure:"write-timeout" env:"SERVER_WRITE_TIMEOUT"`
	// Maximum duration before timing out idle keep-alive connection
	IdleTimeout time.Duration `toml:"idle-timeout" mapstructure:"idle-timeout" env:"SERVER_IDLE_TIMEOUT"`
	// Use manifest.json assets mapping
	AssetsManifest bool `toml:"assets-manifest" mapstructure:"assets-manifest" env:"SERVER_ASSETS_MANIFEST"`
	// Disable browsers caching asset files by setting response headers
	AssetsNoCache bool `toml:"assets-no-cache" mapstructure:"assets-no-cache" env:"SERVER_ASSETS_NO_CACHE"`
	// RenderRecompile enables recompilation of the template on every render call.
	// This should be used in development mode so no server restart is required
	// on template file changes.
	RenderRecompile bool `toml:"render-recompile" mapstructure:"render-recompile" env:"SERVER_RENDER_RECOMPILE"`
	// Use the development mode sessions storer opposed to production mode storer
	// defined in app/sessions.go -- Usually a cookie storer for dev
	// and disk storer for prod.
	SessionsDevStorer bool `toml:"sessions-dev-storer" mapstructure:"sessions-dev-storer" env:"SERVER_SESSIONS_DEV_STORER"`
	// PublicPath defaults to "public" but can be set to something else
	// by the {{.AppEnvName}}_SERVER_PUBLIC_PATH environment variable.
	// This is set by the "abcweb dev" command to instruct the app to
	// load assets from a /tmp folder instead of the local public folder.
	PublicPath string `toml:"public-path" mapstructure:"public-path" env:"SERVER_PUBLIC_PATH"`
}

// DBConfig holds the database config for the app loaded through
// environment variables, or the config.toml file.
type DBConfig struct {
	// DebugMode is a flag to toggle the output of SQLBoiler's SQL queries
	DebugMode bool `toml:"debug-mode" mapstructure:"debug-mode" env:"DB_DEBUG_MODE"`
	// DB is the database software; "postgres", "mysql", etc.
	DB string `toml:"db" mapstructure:"db" env:"DB_DB"`
	// The database name
	DBName  string `toml:"dbname" mapstructure:"dbname" env:"DB_DBNAME"`
	Host    string `toml:"host" mapstructure:"host" env:"DB_HOST"`
	Port    int    `toml:"port" mapstructure:"port" env:"DB_PORT"`
	User    string `toml:"user" mapstructure:"user" env:"DB_USER"`
	Pass    string `toml:"pass" mapstructure:"pass" env:"DB_PASS"`
	SSLMode string `toml:"sslmode" mapstructure:"sslmode" env:"DB_SSLMODE"`

	// Throw an error when the app starts if the database is not
	// using the latest migration
	EnforceMigration bool `toml:"enforce-migration" mapstructure:"enforce-migration" env:"DB_ENFORCE_MIGRATION"`
}

// Bind your passed in config flags to a new viper
// instance, retrieves the active environment section of your config file using
// that viper instance, and then loads your server and db config into
// the passed in cfg struct and validates the db config is set appropriately.
func (c *Config) Bind(flags *pflag.FlagSet, cfg interface{}) (*viper.Viper, error) {
	v, err := c.NewSubViper(flags, cfg)
	if err != nil {
		return v, err
	}

	if err := UnmarshalAppConfig(cfg, v); err != nil {
		return v, err
	}

	val := reflect.Indirect(reflect.ValueOf(cfg))

	// Check if there's a DBConfig object in the cfg struct.
	// If found, then validate all fields on it are set appropriately.
	for i := 0; i < val.NumField(); i++ {
		dbCfg, ok := val.Field(i).Interface().(DBConfig)
		if !ok {
			continue
		}
		if err := ValidateDBConfig(dbCfg); err != nil {
			return v, err
		}
		break
	}

	return v, nil
}

// NewSubViper returns a viper instance activated against the active environment
// configuration subsection and initialized with the config.toml
// configuration file and the environment variable prefix.
// It also takes in the configuration struct so that it can generate the env
// mappings.
func (c *Config) NewSubViper(flags *pflag.FlagSet, cfg interface{}) (*viper.Viper, error) {
	v := viper.New()

	if flags != nil {
		if err := v.BindPFlags(flags); err != nil {
			return nil, err
		}
	}

	if err := c.ConfigureViper(v); err != nil {
		return nil, err
	}

	// Use the env from the config if it's not explicitly set
	env := c.LoadEnv
	if env == "" {
		env = v.GetString("env")
	}

	v = v.Sub(env)
	if v == nil {
		return nil, fmt.Errorf("cannot find env section named %s", env)
	}

	mappings, err := GetTagMappings(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get tag mappings for config struct")
	}

	if c.EnvPrefix != "" {
		for _, m := range mappings {
			v.BindEnv(m.chain, strings.Join([]string{c.EnvPrefix, m.env}, "_"))
		}
	} else {
		for _, m := range mappings {
			v.BindEnv(m.chain, m.env)
		}
	}

	if v == nil {
		return nil, fmt.Errorf("unable to load environment %q from %q", env, c.File)
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
func (c *Config) ConfigureViper(v *viper.Viper) error {
	v.SetConfigType("toml")
	v.SetConfigFile(c.File)
	v.SetEnvPrefix(c.EnvPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	if err := v.ReadInConfig(); err != nil {
		return err
	}
	v.AutomaticEnv()

	return nil
}

// UnmarshalAppConfig unmarshals the viper's configured config file
// into the passed in cfg object containing an AppConfig
func UnmarshalAppConfig(cfg interface{}, v *viper.Viper) error {
	err := v.Unmarshal(cfg)
	if err != nil {
		return err
	}

	val := reflect.Indirect(reflect.ValueOf(cfg))

	// if cfg has an imbedded AppConfig then we need to unmarshal
	// directly into that and overwrite it in the parent struct,
	// since its another layer of indirection and viper
	// can't handle it magically.
	for i := 0; i < val.NumField(); i++ {
		appCfg, ok := val.Field(i).Interface().(AppConfig)
		if !ok {
			continue
		}

		v.Unmarshal(&appCfg)
		val.Field(i).Set(reflect.ValueOf(appCfg))
		// overwrite val to point to the AppConfig so DBConfig can be set below.
		val = val.Field(i)
		break
	}

	// Find *DBConfig and set object appropriately
	for i := 0; i < val.NumField(); i++ {
		dbCfg, ok := val.Field(i).Interface().(DBConfig)
		if !ok {
			continue
		}

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

		val.Field(i).Set(reflect.ValueOf(dbCfg))

		// Finished working on the db cfg struct, so break out
		break
	}

	return nil
}

// ValidateDBConfig returns an error if any of the required db config
// fields are not set to their appropriate values.
func ValidateDBConfig(cfg DBConfig) error {
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

// Mapping represents a chain which is a list of nested object mapstructures
// joined together and seperated by dots (i.e. one.two.three), and the
// accompanying environment variable tag value for the last item in the chain
type Mapping struct {
	chain string
	env   string
}

// Mappings is a slice of mapping
type Mappings []Mapping

func getTagMappingsRecursive(chain string, v reflect.Value) (Mappings, error) {
	mappings := Mappings{}

	for i := 0; i < v.NumField(); i++ {
		cv := v.Field(i)
		tag := v.Type().Field(i).Tag

		ms := tag.Get("mapstructure")
		env := tag.Get("env")

		if cv.Kind() == reflect.Ptr {
			nv := reflect.Indirect(cv)
			// If it has no mapstructure set then fail gracefully,
			// because it's probably not a field that should be read by viper.
			// For example, a pointer to something that is late-initialized
			// and isn't loaded by Bind or present in the config file.
			if !nv.IsValid() && ms == "" {
				continue
			} else if !nv.IsValid() {
				return nil, fmt.Errorf("cannot access non-initialized pointer %#v", cv)
			}
			// Only indirect struct types, if they're valid
			if nv.Kind() == reflect.Struct {
				cv = nv
			}
		}

		// nc = newchain
		var nc string
		if chain != "" {
			nc = strings.Join([]string{chain, ms}, ".")
		} else {
			nc = ms
		}

		switch cv.Kind() {
		case reflect.Struct:
			m, err := getTagMappingsRecursive(nc, cv)
			if err != nil {
				return nil, err
			}
			mappings = append(mappings, m...)
		default:
			if env != "" && ms != "" {
				mappings = append(mappings, Mapping{chain: nc, env: env})
			}
		}
	}

	return mappings, nil
}

// GetTagMappings returns the viper .BindEnv mappings for an entire config
// struct.
func GetTagMappings(cfg interface{}) (Mappings, error) {
	return getTagMappingsRecursive("", reflect.Indirect(reflect.ValueOf(cfg)))
}

// NewFlagSet creates the set of flags specific to the server and db config
// and the root level config (like --version, --env)
func NewFlagSet() *pflag.FlagSet {
	flags := &pflag.FlagSet{}

	flags.AddFlagSet(NewRootFlagSet())
	flags.AddFlagSet(NewServerFlagSet())
	flags.AddFlagSet(NewDBFlagSet())

	return flags
}

// NewRootFlagSet returns a list of top level flags (flags that arent contained
// inside an environment section in the config)
func NewRootFlagSet() *pflag.FlagSet {
	flags := &pflag.FlagSet{}

	// root level flags
	flags.BoolP("version", "", false, "Display the build version hash")
	flags.StringP("env", "e", "prod", "The config files environment to load")

	return flags
}

// NewServerFlagSet returns a list of flags contained within the [server]
// section of a config
func NewServerFlagSet() *pflag.FlagSet {
	flags := &pflag.FlagSet{}

	// server subsection flags
	flags.BoolP("server.live-reload", "", false, "Enable or disable LiveReload")
	flags.BoolP("server.prod-logger", "", true, "Use the production logger, JSON and log level warn")
	flags.StringP("server.bind", "", ":80", `HTTP bind address, eg: ":80"`)
	flags.StringP("server.tls-bind", "", "", `HTTPS bind address, eg: ":443"`)
	flags.StringP("server.tls-cert-file", "", "", "TLS certificate file path")
	flags.StringP("server.tls-key-file", "", "", "TLS key file path")
	flags.StringP("server.public-path", "", "public", "The path to the public folder containing assets")
	flags.DurationP("server.read-timeout", "", time.Second*10, "Maximum duration before timing out read of the request")
	flags.DurationP("server.write-timeout", "", time.Second*15, "Maximum duration before timing out write of the response")
	flags.DurationP("server.idle-timeout", "", time.Second*120, "Maximum duration before timing out idle keep-alive connection")
	// manifest.json is created as a part of the gulp production "build" task,
	// it maps fingerprinted asset names to regular asset names, for example:
	// {"js/main.css": "js/e2a3ff9-main.css"}.
	// This should only be set to true if doing asset fingerprinting.
	flags.BoolP("server.assets-manifest", "", true, "Use manifest.json for mapping asset names to fingerprinted assets")
	// This should be used in development mode to prevent browser caching of assets
	flags.BoolP("server.assets-no-cache", "", false, "Disable browsers caching asset files by setting response headers")
	// This should be used in development mode to avoid having to reload the
	// server on every template file modification.
	flags.BoolP("server.render-recompile", "", false, "Enable recompilation of the template on each render")
	// Defined in app/sessions.go -- Usually cookie storer for dev and disk storer for prod.
	flags.BoolP("server.sessions-dev-storer", "", false, "Use the development mode sessions storer (defined in app/sessions.go)")

	return flags
}

// NewDBFlagSet returns a list of flags contained within the [db] section
// of a config
func NewDBFlagSet() *pflag.FlagSet {
	flags := &pflag.FlagSet{}

	// db subsection flags
	flags.BoolP("db.debug-mode", "", false, "Output database model SQL queries")
	flags.StringP("db.db", "", "", "The database software (postgres|mysql)")
	flags.StringP("db.dbname", "", "", "The database name to connect to")
	flags.StringP("db.host", "", "", "The database hostname, e.g localhost")
	flags.IntP("db.port", "", 0, "The database port")
	flags.StringP("db.user", "", "", "The database username")
	flags.StringP("db.pass", "", "", "The database password")
	flags.StringP("db.sslmode", "", "", "The database sslmode")
	flags.BoolP("db.enforce-migrations", "", true, "Throw error on app start if database is not using latest migration")

	return flags
}
