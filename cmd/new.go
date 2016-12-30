package cmd

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type newConfig struct {
	appPath       string
	appName       string
	prodStorer    string
	devStorer     string
	noGitIgnore   bool
	noBootstrap   bool
	noFontAwesome bool
	noLiveReload  bool
	noSSHKeys     bool
	noReadme      bool
	noConfig      bool
	forceOverride bool
}

var newCmdConfig newConfig

var rgxIgnore = []*regexp.Regexp{
	// i18n isn't implemented yet
	regexp.MustCompile(`.*/templates/i18n.*`),
}

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new <app_path> [flags]",
	Short: "Generate a new ABCWeb application.",
	Long: `The 'abcweb new' command generates a new ABCWeb application with a 
default directory structure and configuration at the path you specify.`,
	Example: "abcweb new ~/go/src/github.com/john/awesomeapp",
	PreRunE: newCmdPreRun,
	Run: func(cmd *cobra.Command, args []string) {
		// use walk to loop over every file and folder in templates folder
		// if the path matches any of the ignore regexps, skip this file/folder
		//
		// if file ends in .tmpl, process the file as a template, and then
		// output it to the users app_path with the .tmpl extension removed if
		// the processing result is not empty.
		//
		// if its a regular file (no tmpl extension), copy the file to the target location
		// if its a folder, attempt to mkdir the folder in app_path (do nothing if it already exists)
		//
		//
		fmt.Println("new called")
	},
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
	newCmd.Flags().BoolP("force-override", "", false, "Force override of existing files in app_path")

	RootCmd.AddCommand(newCmd)
	viper.BindPFlags(newCmd.Flags())
}

func newCmdPreRun(cmd *cobra.Command, args []string) error {
	var err error

	newCmdConfig = newConfig{
		noGitIgnore:   viper.GetBool("no-gitignore"),
		noBootstrap:   viper.GetBool("no-twitter-bootstrap"),
		noFontAwesome: viper.GetBool("no-font-awesome"),
		noLiveReload:  viper.GetBool("no-livereload"),
		noSSHKeys:     viper.GetBool("no-ssh-keys"),
		noReadme:      viper.GetBool("no-readme"),
		noConfig:      viper.GetBool("no-config"),
		forceOverride: viper.GetBool("force-override"),
		prodStorer:    viper.GetString("sessions-prod-storer"),
		devStorer:     viper.GetString("sessions-dev-storer"),
	}

	newCmdConfig.appPath, newCmdConfig.appName, err = getAppPath(args)
	return err
}

func getAppPath(args []string) (string, string, error) {
	if len(args) == 0 {
		return "", "", errors.New("must provide an app path")
	}

	appPath := filepath.Clean(args[0])
	appName := filepath.Base(appPath)

	if appName == "." || appName == "/" {
		return appPath, "", errors.New("app path must contain an output folder name")
	}

	return appPath, appName, nil
}
