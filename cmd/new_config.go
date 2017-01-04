package cmd

const (
	templatesDirectory = "templates"
	basePackage        = "github.com/nullbio/abcweb"
)

type newConfig struct {
	AppPath        string
	ImportPath     string
	AppName        string
	TLSCommonName  string
	ProdStorer     string
	DevStorer      string
	DefaultEnv     string
	Bootstrap      string
	NoBootstrapJS  bool
	NoGitIgnore    bool
	NoFontAwesome  bool
	NoLiveReload   bool
	NoTLSCerts     bool
	NoReadme       bool
	NoConfig       bool
	ForceOverwrite bool
	TLSCertsOnly   bool
	NoHTTPRedirect bool
}

var skipDirs = []string{
	// i18n is not implemented yet
	"i18n",
}

var fontAwesomeFiles = []string{
	"font-awesome.min.css",
	"FontAwesome.otf",
	"fontawesome-webfont.eot",
	"fontawesome-webfont.svg",
	"fontawesome-webfont.ttf",
	"fontawesome-webfont.woff",
	"fontawesome-webfont.woff2",
}

var bootstrapNone = []string{
	"bootstrap-flex.css",
	"bootstrap-grid.css",
	"bootstrap-reboot.css",
	"bootstrap.css",
	"bootstrap.js",
	"jquery-3.1.1.js",
	"tether.js",
}

var bootstrapRegular = []string{
	"bootstrap-flex.css",
	"bootstrap-grid.css",
	"bootstrap-reboot.css",
}

var bootstrapFlex = []string{
	"bootstrap-grid.css",
	"bootstrap-reboot.css",
	"bootstrap.css",
}

var bootstrapGridOnly = []string{
	"bootstrap-flex.css",
	"bootstrap-reboot.css",
	"bootstrap.css",
	"bootstrap.js",
	"jquery-3.1.1.js",
	"tether.js",
}

var bootstrapRebootOnly = []string{
	"bootstrap-flex.css",
	"bootstrap-grid.css",
	"bootstrap.css",
	"bootstrap.js",
	"jquery-3.1.1.js",
	"tether.js",
}

var bootstrapGridRebootOnly = []string{
	"bootstrap-flex.css",
	"bootstrap.css",
	"bootstrap.js",
	"jquery-3.1.1.js",
	"tether.js",
}

var bootstrapJSFiles = []string{
	"bootstrap.js",
	"jquery-3.1.1.js",
	"tether.js",
}
