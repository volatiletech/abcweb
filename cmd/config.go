package cmd

import (
	"github.com/volatiletech/sqlboiler/boilingcore"
	"github.com/volatiletech/abcweb/config"
)

const (
	basePackage         = "github.com/volatiletech/abcweb"
	templatesDirectory  = "templates"
	migrationsDirectory = "db/migrations"
)

type buildConfig struct{}
type distConfig struct{}

type modelsConfig struct {
	config.DBConfig
}

type migrateConfig struct {
	config.DBConfig
}

type newConfig struct {
	AppPath        string `toml:"app-path"`
	ImportPath     string `toml:"import-path"`
	AppName        string `toml:"app-name"`
	AppEnvName     string `toml:"app-env-name"`
	TLSCommonName  string `toml:"tls-common-name"`
	ProdStorer     string `toml:"prod-storer"`
	DevStorer      string `toml:"dev-storer"`
	DefaultEnv     string `toml:"default-env"`
	Bootstrap      string `toml:"bootstrap"`
	NoBootstrapJS  bool   `toml:"no-bootstrap-js"`
	NoGulp         bool   `toml:"no-gulp"`
	NoFontAwesome  bool   `toml:"no-font-awesome"`
	NoLiveReload   bool   `toml:"no-live-reload"`
	NoTLSCerts     bool   `toml:"no-tls-certs"`
	NoReadme       bool   `toml:"no-readme"`
	NoConfig       bool   `toml:"no-config"`
	NoSessions     bool   `toml:"no-sessions"`
	ForceOverwrite bool   `toml:"force-overwrite"`
	SkipNPMInstall bool   `toml:"skip-npm-install"`
	SkipDepEnsure  bool   `toml:"skip-dep-ensure"`
	SkipGitInit    bool   `toml:"skip-git-init"`
	Silent         bool   `toml:"silent"`
	Verbose        bool   `toml:"verbose"`
}

// Create some config variables
var (
	migrateCmdConfig migrateConfig
	modelsCmdConfig  *boilingcore.Config
	modelsCmdState   *boilingcore.State
)

// skipDirs are the directories to skip creating for new command
var skipDirs = []string{
	// i18n is not implemented yet
	"i18n",
}

// emptyDirs are the (potentially) empty directories that need to be created
// manually because empty directories cannot be committed to git
var emptyDirs = []string{
	"assets/css",
	"assets/fonts",
	"assets/img",
	"assets/js",
	"assets/vendor/audio",
	"assets/vendor/css",
	"assets/vendor/fonts",
	"assets/vendor/img",
	"assets/vendor/js",
	"assets/vendor/video",
	"public/assets",
	"db/migrations",
}

// Exclude the following files
var bootstrapNone = []string{
	"bootstrap.scss",
	"bootstrap-grid.scss",
	"bootstrap-reboot.scss",
	"_custom.scss",
	"jquery-3.1.1.js",
	"tether.js",
}

var bootstrapRegular = []string{
	"bootstrap-grid.scss",
	"bootstrap-reboot.scss",
}

var bootstrapGridOnly = []string{
	"bootstrap-reboot.scss",
	"bootstrap.scss",
	"jquery-3.1.1.js",
	"tether.js",
}

var bootstrapRebootOnly = []string{
	"bootstrap-grid.scss",
	"bootstrap.scss",
	"jquery-3.1.1.js",
	"tether.js",
}

var bootstrapGridRebootOnly = []string{
	"bootstrap.scss",
	"jquery-3.1.1.js",
	"tether.js",
}

var bootstrapJSFiles = []string{
	"jquery-3.1.1.js",
	"tether.js",
}
