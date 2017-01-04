package cmd

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"go/build"
	"go/format"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/nullbio/abcweb/cert"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

const (
	templatesDirectory = "templates"
	basePackage        = "github.com/nullbio/abcweb"
)

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

var newCmdConfig newConfig

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new <import_path> [flags]",
	Short: "Generate a new ABCWeb application.",
	Long: `The 'abcweb new' command generates a new ABCWeb application with a 
default directory structure and configuration at the Go src import path you specify.

The app will generate in $GOPATH/src/<import_path>.
`,
	Example: "abcweb new github.com/yourusername/myapp",
	PreRunE: newCmdPreRun,
	RunE:    newCmdRun,
}

func init() {
	newCmd.Flags().StringP("sessions-prod-storer", "p", "disk", "Session storer to use in production mode")
	newCmd.Flags().StringP("sessions-dev-storer", "d", "cookie", "Session storer to use in development mode")
	newCmd.Flags().StringP("tls-common-name", "", "localhost", "Common Name for generated TLS certificate")
	newCmd.Flags().StringP("default-env", "", "prod", "Default APP_ENV to use when starting server")
	newCmd.Flags().StringP("bootstrap", "b", "flex", "Include Twitter Bootstrap 4 (none|regular|gridonly|rebootonly|gridandrebootonly)")
	newCmd.Flags().BoolP("no-gitignore", "g", false, "Skip .gitignore file")
	newCmd.Flags().BoolP("no-bootstrap-js", "j", false, "Skip Twitter Bootstrap 4 javascript inclusion")
	newCmd.Flags().BoolP("no-font-awesome", "f", false, "Skip Font Awesome inclusion")
	newCmd.Flags().BoolP("no-livereload", "l", false, "Don't support LiveReload")
	newCmd.Flags().BoolP("no-tls-certs", "s", false, "Skip generation of self-signed TLS cert files")
	newCmd.Flags().BoolP("no-readme", "r", false, "Skip README.md files")
	newCmd.Flags().BoolP("no-config", "c", false, "Skip default config.toml file")
	newCmd.Flags().BoolP("force-overwrite", "", false, "Force overwrite of existing files in your import_path")
	newCmd.Flags().BoolP("tls-certs-only", "", false, "Only generate self-signed TLS cert files")
	newCmd.Flags().BoolP("no-http-redirect", "", false, "Disable the http -> https redirect when using TLS")

	RootCmd.AddCommand(newCmd)
	viper.BindPFlags(newCmd.Flags())
}

func newCmdPreRun(cmd *cobra.Command, args []string) error {
	var err error

	newCmdConfig = newConfig{
		NoGitIgnore:    viper.GetBool("no-gitignore"),
		NoBootstrapJS:  viper.GetBool("no-bootstrap-js"),
		NoFontAwesome:  viper.GetBool("no-font-awesome"),
		NoLiveReload:   viper.GetBool("no-livereload"),
		NoTLSCerts:     viper.GetBool("no-tls-certs"),
		TLSCertsOnly:   viper.GetBool("tls-certs-only"),
		NoReadme:       viper.GetBool("no-readme"),
		NoConfig:       viper.GetBool("no-config"),
		ForceOverwrite: viper.GetBool("force-overwrite"),
		NoHTTPRedirect: viper.GetBool("no-http-redirect"),
		ProdStorer:     viper.GetString("sessions-prod-storer"),
		DevStorer:      viper.GetString("sessions-dev-storer"),
		TLSCommonName:  viper.GetString("tls-common-name"),
		DefaultEnv:     viper.GetString("default-env"),
		Bootstrap:      strings.ToLower(viper.GetString("bootstrap")),
	}

	validBootstrap := []string{"none", "flex", "regular", "gridonly", "rebootonly", "gridandrebootonly"}
	found := false
	for _, v := range validBootstrap {
		if newCmdConfig.Bootstrap == v {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("invalid bootstrap option (%q) supplied, valid options are: none|flex|regular|gridonly|rebootonly|gridandrebootonly", newCmdConfig.Bootstrap)
	}

	newCmdConfig.AppPath, newCmdConfig.ImportPath, newCmdConfig.AppName, err = getAppPath(args)
	return err
}

func newCmdRun(cmd *cobra.Command, args []string) error {
	fmt.Println("Generating in:", newCmdConfig.AppPath)

	// Make the app directory if it doesnt already exist.
	// Can get dir not exist errors on --tls-cert-only runs if we don't do this.
	err := os.MkdirAll(newCmdConfig.AppPath, 0755)
	if err != nil {
		return err
	}

	if !newCmdConfig.TLSCertsOnly {
		// Get base path containing templates folder and source files
		p, _ := build.Default.Import(basePackage, "", build.FindOnly)
		if p == nil || len(p.Dir) == 0 {
			return errors.New("cannot locate base path containing templates folder")
		}

		// Walk all files in the templates folder
		basePath := filepath.Join(p.Dir, "templates")
		err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
			return newCmdWalk(basePath, path, info, err)
		})
		if err != nil {
			return err
		}
	}

	// Generate TLS certs if requested
	if !newCmdConfig.NoTLSCerts {
		err := generateTLSCerts()
		if err != nil {
			return err
		}
	}

	fmt.Println("\tresult -> Finished")
	return nil
}

func generateTLSCerts() error {
	certFilePath := filepath.Join(newCmdConfig.AppPath, "cert.pem")
	privateKeyPath := filepath.Join(newCmdConfig.AppPath, "private.key")

	if !newCmdConfig.TLSCertsOnly {
		_, err := os.Stat(certFilePath)
		if err == nil || (err != nil && !os.IsNotExist(err)) {
			return nil
		}
	}

	fmt.Println("\trun -> TLS Certificate Generator")
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	err = cert.WriteCertFile(certFilePath, newCmdConfig.AppName,
		newCmdConfig.TLSCommonName, &privateKey.PublicKey, privateKey)
	if err != nil {
		return err
	}
	fmt.Printf("\tcreate -> %s\n", filepath.Join(newCmdConfig.AppName, "cert.pem"))

	if err := cert.WritePrivateKey(privateKeyPath, privateKey); err != nil {
		return err
	}
	fmt.Printf("\tcreate -> %s\n", filepath.Join(newCmdConfig.AppName, "private.key"))

	return nil
}

func newCmdWalk(basePath string, path string, info os.FileInfo, err error) error {
	// Skip files and dirs depending on command line args
	if skip, err := processSkips(basePath, path, info); skip {
		return err
	}

	// Get the path for command line output, and the output target fullPath
	cleanPath, fullPath := getProcessedPaths(path, string(os.PathSeparator), newCmdConfig)

	fileContents := &bytes.Buffer{}

	// Check if the output file or folder already exists
	_, err = os.Stat(fullPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	outputExists := err == nil

	// Dirs only get created if they don't already exist
	if info.IsDir() {
		// Don't bother trying to create the folder if it already exists.
		// Return now so we don't get "created" output
		if outputExists {
			return nil
		}

		err = os.MkdirAll(fullPath, 0755)
		if err != nil {
			return err
		}
	} else {
		// Files only get created if they don't already exist, or force overwrite is enabled
		if !newCmdConfig.ForceOverwrite && outputExists {
			return nil
		}

		rawFileContents, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		// Only process the file as a template if it has the .tmpl extension
		if strings.HasSuffix(path, ".tmpl") {
			t, err := template.New("").Funcs(templateFuncs).Parse(string(rawFileContents))
			if err != nil {
				return err
			}

			err = t.Execute(fileContents, newCmdConfig)
			if err != nil {
				return err
			}
		} else {
			if _, err := fileContents.Write(rawFileContents); err != nil {
				return err
			}
		}

		// Gofmt go files before save.
		if strings.HasSuffix(fullPath, ".go") {
			res, err := format.Source(fileContents.Bytes())
			if err != nil {
				return err
			}
			fileContents.Reset()
			if _, err := fileContents.Write(res); err != nil {
				return err
			}
		}

		err = ioutil.WriteFile(fullPath, fileContents.Bytes(), 0664)
		if err != nil {
			return err
		}
	}

	fmt.Printf("\tcreate -> %s\n", cleanPath)
	return nil
}

func processSkips(basePath string, path string, info os.FileInfo) (skip bool, err error) {
	// Ignore the root folder
	if path == basePath && info.IsDir() {
		return true, nil
	}

	// Skip directories defined in skipDirs slice
	if info.IsDir() {
		for _, skipDir := range skipDirs {
			if info.Name() == skipDir {
				return true, filepath.SkipDir
			}
		}
	}

	// Skip readme files if requested
	if newCmdConfig.NoReadme {
		if info.Name() == "README.md" || info.Name() == "README.md.tmpl" {
			return true, nil
		}
	}

	// Skip gitignore if requested
	if newCmdConfig.NoGitIgnore {
		if info.Name() == ".gitignore" || info.Name() == ".gitignore.tmpl" {
			return true, nil
		}
	}

	// Skip default config.toml if requested
	if newCmdConfig.NoConfig {
		if info.Name() == "config.toml" || info.Name() == "config.toml.tmpl" {
			return true, nil
		}
	}

	// Skip FontAwesome files if requested
	if newCmdConfig.NoFontAwesome {
		for _, faFile := range fontAwesomeFiles {
			if info.Name() == faFile || info.Name() == faFile+".tmpl" {
				return true, nil
			}
		}
	}

	var bsArr []string
	if newCmdConfig.Bootstrap == "none" {
		bsArr = bootstrapNone
	} else if newCmdConfig.Bootstrap == "flex" {
		bsArr = bootstrapFlex
	} else if newCmdConfig.Bootstrap == "regular" {
		bsArr = bootstrapRegular
	} else if newCmdConfig.Bootstrap == "gridonly" {
		bsArr = bootstrapGridOnly
	} else if newCmdConfig.Bootstrap == "rebootonly" {
		bsArr = bootstrapRebootOnly
	} else if newCmdConfig.Bootstrap == "gridandrebootonly" {
		bsArr = bootstrapGridRebootOnly
	}

	// Skip files contained within bsArr
	for _, bsFile := range bsArr {
		if info.Name() == bsFile || info.Name() == bsFile+".tmpl" {
			return true, nil
		}
	}

	// Skip Twitter Bootstrap JS files if requested
	if newCmdConfig.NoBootstrapJS {
		for _, bsFile := range bootstrapJSFiles {
			if info.Name() == bsFile || info.Name() == bsFile+".tmpl" {
				return true, nil
			}
		}
	}

	return false, nil
}

func getAppPath(args []string) (appPath string, importPath string, appName string, err error) {
	if len(args) == 0 {
		return "", "", "", errors.New("must provide an app path")
	}

	appPath = filepath.Clean(args[0])
	importPath = appPath

	// Somewhat validate provided app path, valid paths will have at least 2 components
	appPathChunks := strings.Split(appPath, string(os.PathSeparator))
	if len(appPathChunks) < 2 {
		return "", "", "", errors.New("invalid app path provided, see --help for example")
	}

	_, appName = filepath.Split(appPath)

	if appName == "" || appName == "." || appName == "/" {
		return "", "", "", errors.New("app path must contain an output folder name")
	}

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return "", "", "", errors.New("cannot get GOPATH from environment variables")
	}

	// If GOPATH has more than one directory, prompt user to choose which one
	goPathChunks := strings.Split(gopath, string(os.PathListSeparator))
	if len(goPathChunks) > 1 {
		fmt.Println("Your GOPATH has multiple paths, select your desired GOPATH:")
		for pos, chunk := range goPathChunks {
			fmt.Printf("[%d] %s\n", pos+1, chunk)
		}

		num := 0
		for num < 1 || num > len(goPathChunks) {
			fmt.Printf("Select GOPATH number: ")
			fmt.Scanln(&num)
			fmt.Println(num)
		}

		gopath = goPathChunks[num-1]
	}

	// Target directory is $GOPATH/src/<import_path>
	appPath = filepath.Join(gopath, "src", appPath)
	return appPath, importPath, appName, nil
}

func getProcessedPaths(path string, pathSeparator string, config newConfig) (cleanPath string, fullPath string) {
	chunks := strings.Split(path, pathSeparator)
	var newChunks []string

	var found int
	for i := 0; i < len(chunks); i++ {
		if chunks[i] == templatesDirectory {
			found = i
			break
		}
	}

	newChunks = append(newChunks, chunks[found:]...)

	// Make cleanPath for results output
	newChunks[0] = config.AppName
	newChunks[len(newChunks)-1] = strings.TrimSuffix(newChunks[len(newChunks)-1], ".tmpl")
	cleanPath = strings.Join(newChunks, string(os.PathSeparator))

	// Make fullPath for destination save
	newChunks[0] = config.AppPath
	fullPath = strings.Join(newChunks, string(os.PathSeparator))

	return cleanPath, fullPath
}
