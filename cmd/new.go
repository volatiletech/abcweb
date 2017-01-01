package cmd

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/gob"
	"encoding/pem"
	"errors"
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type newConfig struct {
	AppPath        string
	ImportPath     string
	AppName        string
	ProdStorer     string
	DevStorer      string
	NoGitIgnore    bool
	NoBootstrap    bool
	NoFontAwesome  bool
	NoLiveReload   bool
	NoSSLCerts     bool
	NoReadme       bool
	NoConfig       bool
	ForceOverwrite bool
	SSLCertsOnly   bool
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

var bootstrapFiles = []string{
	"bootstrap.css",
	"bootstrap-flex.css",
	"bootstrap-grid.css",
	"bootstrap-reboot.css",
	"bootstrap.js",
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
	newCmd.Flags().BoolP("no-gitignore", "g", false, "Skip .gitignore file")
	newCmd.Flags().BoolP("no-twitter-bootstrap", "t", false, "Skip Twitter Bootstrap 4 inclusion")
	newCmd.Flags().BoolP("no-font-awesome", "f", false, "Skip Font Awesome inclusion")
	newCmd.Flags().BoolP("no-livereload", "l", false, "Don't support LiveReload")
	newCmd.Flags().BoolP("no-ssl-certs", "s", false, "Skip generation of self-signed SSL cert files")
	newCmd.Flags().BoolP("no-readme", "r", false, "Skip README.md files")
	newCmd.Flags().BoolP("no-config", "c", false, "Skip default config.toml file")
	newCmd.Flags().BoolP("force-overwrite", "", false, "Force overwrite of existing files in your import_path")
	newCmd.Flags().BoolP("ssl-certs-only", "", false, "Only generate self-signed SSL cert files")

	RootCmd.AddCommand(newCmd)
	viper.BindPFlags(newCmd.Flags())
}

func newCmdPreRun(cmd *cobra.Command, args []string) error {
	var err error

	newCmdConfig = newConfig{
		NoGitIgnore:    viper.GetBool("no-gitignore"),
		NoBootstrap:    viper.GetBool("no-twitter-bootstrap"),
		NoFontAwesome:  viper.GetBool("no-font-awesome"),
		NoLiveReload:   viper.GetBool("no-livereload"),
		NoSSLCerts:     viper.GetBool("no-ssl-certs"),
		SSLCertsOnly:   viper.GetBool("ssl-certs-only"),
		NoReadme:       viper.GetBool("no-readme"),
		NoConfig:       viper.GetBool("no-config"),
		ForceOverwrite: viper.GetBool("force-overwrite"),
		ProdStorer:     viper.GetString("sessions-prod-storer"),
		DevStorer:      viper.GetString("sessions-dev-storer"),
	}

	newCmdConfig.AppPath, newCmdConfig.ImportPath, newCmdConfig.AppName, err = getAppPath(args)
	return err
}

func newCmdRun(cmd *cobra.Command, args []string) error {
	fmt.Println("Generating in:", newCmdConfig.AppPath)

	// Make the app directory if it doesnt already exist.
	// Can get dir not exist errors on --ssl-cert-only runs if we don't do this.
	err := os.MkdirAll(newCmdConfig.AppPath, 0755)
	if err != nil {
		return err
	}

	if !newCmdConfig.SSLCertsOnly {
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

	// Generate SSL certs if requested
	if !newCmdConfig.NoSSLCerts {
		err := generateSSLCerts()
		if err != nil {
			return err
		}
	}

	fmt.Println("\tresult -> Finished")
	return nil
}

func generateSSLCerts() error {
	pemFilePath := filepath.Join(newCmdConfig.AppPath, "private.pem")
	if !newCmdConfig.SSLCertsOnly {
		_, err := os.Stat(pemFilePath)
		if err == nil || (err != nil && !os.IsNotExist(err)) {
			return nil
		}
	}

	fmt.Println("\trun -> SSL Certificate Generator")
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return err
	}

	publicKey := &privateKey.PublicKey

	privateKeyPath := filepath.Join(newCmdConfig.AppPath, "private.key")
	privateKeyFile, err := os.Create(privateKeyPath)
	if err != nil {
		return err
	}

	// Write gob encoded private key to file
	privateKeyEncoder := gob.NewEncoder(privateKeyFile)
	if err := privateKeyEncoder.Encode(privateKey); err != nil {
		return err
	}
	privateKeyFile.Close()
	fmt.Printf("\tcreate -> %s\n", filepath.Join(newCmdConfig.AppName, "private.key"))

	publicKeyPath := filepath.Join(newCmdConfig.AppPath, "public.key")
	publicKeyFile, err := os.Create(publicKeyPath)
	if err != nil {
		return err
	}

	// Write gob encoded public key to file
	publicKeyEncoder := gob.NewEncoder(publicKeyFile)
	if err := publicKeyEncoder.Encode(publicKey); err != nil {
		return err
	}
	publicKeyFile.Close()
	fmt.Printf("\tcreate -> %s\n", filepath.Join(newCmdConfig.AppName, "public.key"))

	pemFile, err := os.Create(pemFilePath)
	if err != nil {
		return err
	}

	pemKey := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	if err := pem.Encode(pemFile, pemKey); err != nil {
		return err
	}
	fmt.Printf("\tcreate -> %s\n", filepath.Join(newCmdConfig.AppName, "private.pem"))

	return pemFile.Close()
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
		}

		err = ioutil.WriteFile(fullPath, fileContents.Bytes(), 0644)
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

	// Skip FontAwesome files if requested
	if newCmdConfig.NoFontAwesome {
		for _, faFile := range fontAwesomeFiles {
			if info.Name() == faFile || info.Name() == faFile+".tmpl" {
				return true, nil
			}
		}
	}

	// Skip Twitter Bootstrap files if requested
	if newCmdConfig.NoBootstrap {
		for _, bsFile := range bootstrapFiles {
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
