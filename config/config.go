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
	"github.com/nullbio/shift"
	"github.com/spf13/afero"
)

// AppFS is a handle to the filesystem in use
var AppFS = afero.NewOsFs()

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
}

// AppConfig holds the relevant generated app config.toml file variables
type AppConfig struct {
	DefaultEnv string `toml:"default_env"`
}

// LoadDBConfig loads the config vars in database.toml into a DBConfig object
func LoadDBConfig(appPath string, env string) *DBConfig {
	cfg := &DBConfig{}
	appName := GetAppName(appPath)
	configPath := filepath.Join(appPath, "database.toml")

	err := shift.Load(cfg, configPath, strmangle.EnvAppName(appName), env)
	if err != nil {
		log.Fatal("unable to load database.toml:", err)
	}

	return cfg
}

// GetActiveEnv attempts to get the config.toml and database.toml environment
// to load by checking the following, in the following order:
// 1. environment variable $APPNAME_ENV (APPNAME is envAppName variable value)
// 2. config.toml "default_env"
func GetActiveEnv(appPath string) string {
	appName := strmangle.EnvAppName(GetAppName(appPath))

	val := os.Getenv(appName + "_ENV")
	if val != "" {
		return val
	}

	contents, err := afero.ReadFile(AppFS, filepath.Join(appPath, "config.toml"))
	if err != nil {
		log.Fatal("unable to read config.toml file:", err)
	}

	var config AppConfig

	_, err = toml.Decode(string(contents), &config)
	if err != nil {
		log.Fatal("unable to decode config.toml file:", err)
	}

	return config.DefaultEnv
}

// GetAppPath executes the git cmd "git rev-parse --show-toplevel" to obtain
// the full path of the current app. The last folder in the path is the app name.
func GetAppPath() string {
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
