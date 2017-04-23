package cmd

import "github.com/volatiletech/abcweb/config"

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
	AppPath          string
	ImportPath       string
	AppName          string
	AppEnvName       string
	TLSCommonName    string
	ProdStorer       string
	DevStorer        string
	DefaultEnv       string
	Bootstrap        string
	NoBootstrapJS    bool
	NoGulp           bool
	NoFontAwesome    bool
	NoLiveReload     bool
	NoTLSCerts       bool
	NoReadme         bool
	NoConfig         bool
	NoSessions       bool
	NoRequestID      bool
	ForceOverwrite   bool
	TLSCertsOnly     bool
	NoHTTPRedirect   bool
	SkipNPMInstall   bool
	SkipGovendorSync bool
	SkipGitInit      bool
	Silent           bool
}

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
