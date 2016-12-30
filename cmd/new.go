package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type newConfig struct {
	appPath        string
	appName        string
	prodStorer     string
	devStorer      string
	noGitIgnore    bool
	noBootstrap    bool
	noFontAwesome  bool
	noLiveReload   bool
	noSSHKeys      bool
	noReadme       bool
	noConfig       bool
	forceOverwrite bool
}

var newCmdConfig newConfig

var skipDirs = []string{
	// i18n is not implemented yet
	"i18n",
}

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new <app_path> [flags]",
	Short: "Generate a new ABCWeb application.",
	Long: `The 'abcweb new' command generates a new ABCWeb application with a 
default directory structure and configuration at the path you specify.`,
	Example: "abcweb new ~/go/src/github.com/john/awesomeapp",
	PreRunE: newCmdPreRun,
	RunE:    newCmdRun,
}

func newCmdRun(cmd *cobra.Command, args []string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	return filepath.Walk(filepath.Join(wd, "templates"), newCmdWalk)
}

func newCmdWalk(path string, info os.FileInfo, err error) error {
	// Ignore the root folder
	if path == "templates" && info.IsDir() {
		return nil
	}

	// Skip directories defined in skipDirs slice
	if info.IsDir() {
		for _, skipDir := range skipDirs {
			if info.Name() == skipDir {
				return filepath.SkipDir
			}
		}
	}

	chunks := strings.Split(path, string(os.PathSeparator))

	// Make cleanPath
	chunks[0] = newCmdConfig.appName
	chunks[len(chunks)-1] = strings.TrimSuffix(chunks[len(chunks)-1], ".tmpl")
	cleanPath := strings.Join(chunks, string(os.PathSeparator))

	// Make fullPath
	chunks[0] = newCmdConfig.appPath
	fullPath := strings.Join(chunks, string(os.PathSeparator))

	var fileContents *bytes.Buffer

	if info.IsDir() {
		os.MkdirAll(fullPath, 0755)
	} else {
		rawTmplContents, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		t, err := template.New("").Parse(string(rawTmplContents))
		if err != nil {
			return err
		}

		err = t.Execute(fileContents, templateFuncs)
		if err != nil {
			return err
		}

		// if overwrites are allowed, call WriteFile directly
		if newCmdConfig.forceOverwrite {
			ioutil.WriteFile(fullPath, fileContents.Bytes(), 0644)
		} else { // otherwise only create file if it doesn't already exist
			_, err := os.Stat(fullPath)
			if os.IsNotExist(err) {
				ioutil.WriteFile(fullPath, fileContents.Bytes(), 0644)
			} else if err != nil {
				return err
			} else { // If file already exists return nil instantly
				return nil
			}
		}
	}

	fmt.Printf("\tcreate -> %s\n", cleanPath)
	return nil
}

func init() {
	newCmd.Flags().StringP("sessions-prod-storer", "p", "disk", "Session storer to use in production mode")
	newCmd.Flags().StringP("sessions-dev-storer", "d", "cookie", "Session storer to use in development mode")
	newCmd.Flags().BoolP("no-gitignore", "g", false, "Skip .gitignore file")
	newCmd.Flags().BoolP("no-twitter-bootstrap", "t", false, "Skip Twitter Bootstrap 4 inclusion")
	newCmd.Flags().BoolP("no-font-awesome", "f", false, "Skip Font Awesome inclusion")
	newCmd.Flags().BoolP("no-livereload", "l", false, "Don't support LiveReload")
	newCmd.Flags().BoolP("no-ssh-keys", "s", false, "Skip generation of server.key and server.pem files")
	newCmd.Flags().BoolP("no-readme", "r", false, "Skip README.md files")
	newCmd.Flags().BoolP("no-config", "c", false, "Skip default config.toml file")
	newCmd.Flags().BoolP("force-overwrite", "", false, "Force overwrite of existing files in app_path")

	RootCmd.AddCommand(newCmd)
	viper.BindPFlags(newCmd.Flags())
}

func newCmdPreRun(cmd *cobra.Command, args []string) error {
	var err error

	newCmdConfig = newConfig{
		noGitIgnore:    viper.GetBool("no-gitignore"),
		noBootstrap:    viper.GetBool("no-twitter-bootstrap"),
		noFontAwesome:  viper.GetBool("no-font-awesome"),
		noLiveReload:   viper.GetBool("no-livereload"),
		noSSHKeys:      viper.GetBool("no-ssh-keys"),
		noReadme:       viper.GetBool("no-readme"),
		noConfig:       viper.GetBool("no-config"),
		forceOverwrite: viper.GetBool("force-overwrite"),
		prodStorer:     viper.GetString("sessions-prod-storer"),
		devStorer:      viper.GetString("sessions-dev-storer"),
	}

	newCmdConfig.appPath, newCmdConfig.appName, err = getAppPath(args)
	return err
}

func getAppPath(args []string) (string, string, error) {
	if len(args) == 0 {
		return "", "", errors.New("must provide an app path")
	}

	appPath := filepath.Clean(args[0])
	_, appName := filepath.Split(appPath)

	if appName == "" || appName == "." || appName == "/" {
		return appPath, "", errors.New("app path must contain an output folder name")
	}

	return appPath, appName, nil
}
