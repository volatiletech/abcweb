package config

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/nullbio/abcweb/strmangle"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

const (
	// DBConfigFilename is the filename for the database config file
	DBConfigFilename = "database.toml"
	// AppConfigFilename is the filename for the app config file
	AppConfigFilename = "config.toml"
)

// AppFS is a handle to the filesystem in use
var AppFS = afero.NewOsFs()

// AppPath is the path to the project, set using the init function
var AppPath string

// ActiveEnv is the environment mode currently set by "env" in config.toml
// or APPNAME_ENV environment variable. This mode indicates what section of
// config variables to to load into the config structs.
var ActiveEnv string

// ModeViper is a *viper.Viper that has been initialized to:
// Load the active environment section of the AppPath/database.toml file
// Load environment variables with a prefix of APPNAME
// Replace "-" with "_" in environment variable names
var ModeViper *viper.Viper

func init() {
	AppPath = getAppPath()
	ActiveEnv = getActiveEnv(AppPath)
	ModeViper = NewModeViper(AppPath, ActiveEnv)
}

// DBConfig holds the configuration variables contained in the database.toml
// file for the environment currently loaded (obtained from GetDatabaseEnv())
type DBConfig struct {
	DB      string
	Host    string
	Port    int
	DBName  string
	User    string
	Pass    string
	SSLMode string
	// Other SQLBoiler flags
	Blacklist        []string
	Whitelist        []string
	Tag              []string
	BaseDir          string
	Output           string
	PkgName          string
	Schema           string
	TinyintNotBool   bool
	NoAutoTimestamps bool
	Debug            bool
	NoHooks          bool
	NoTests          bool

	MigrationsSQL bool   `toml:"migrations.sql"`
	MigrationsDir string `toml:"migrations.dir"`
}

// AppConfig holds the relevant generated app config.toml file variables
type AppConfig struct {
	DefaultEnv string `toml:"env"`
}

var testHarnessViperReadConfig = func(newViper *viper.Viper) error {
	return newViper.ReadInConfig()
}

// NewModeViper creates a viper.Viper with config path and environment prefixes
// set. It also specifies a Sub of the active environment (the chosen env mode)
// and reads in the config file.
func NewModeViper(appPath string, env string) *viper.Viper {
	newViper := viper.New()
	newViper.SetConfigType("toml")
	newViper.SetConfigFile(filepath.Join(appPath, DBConfigFilename))
	newViper.SetEnvPrefix(strmangle.EnvAppName(GetAppName(appPath)))
	newViper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	if env == "" {
		return newViper
	}

	// Only give a warning on errors here, so we can fallback to other validation
	// methods. Users can use env vars or cmd line flags if a config is not found.
	err := testHarnessViperReadConfig(newViper)
	if err != nil {
		return newViper
	}

	modeViper := newViper.Sub(env)
	if modeViper == nil {
		return newViper
	}

	return modeViper
}

// getActiveEnv attempts to get the config.toml and database.toml environment
// to load by checking the following, in the following order:
// 1. environment variable $APPNAME_ENV (APPNAME is envAppName variable value)
// 2. config.toml default environment field "env"
func getActiveEnv(appPath string) string {
	appName := strmangle.EnvAppName(GetAppName(appPath))

	val := os.Getenv(appName + "_ENV")
	if val != "" {
		return val
	}

	contents, err := afero.ReadFile(AppFS, filepath.Join(appPath, AppConfigFilename))
	if err != nil {
		return ""
	}

	var config AppConfig

	_, err = toml.Decode(string(contents), &config)
	if err != nil {
		return ""
	}

	return config.DefaultEnv
}

// getAppPath executes the git cmd "git rev-parse --show-toplevel" to obtain
// the full path of the current app. The last folder in the path is the app name.
func getAppPath() string {
	gitCmd := exec.Command("git", "rev-parse", "--show-toplevel")

	b := &bytes.Buffer{}
	gitCmd.Stdout = b

	err := gitCmd.Run()
	if err != nil {
		log.Fatal("Cannot execute git command:", err)
	}

	output := b.String()

	if len(output) == 0 {
		log.Fatalln("No output for git command")
	}

	return strings.TrimSpace(output)
}

// GetAppName gets the appname portion of a project path
func GetAppName(appPath string) string {
	split := strings.Split(appPath, string(os.PathSeparator))
	return split[len(split)-1]
}
