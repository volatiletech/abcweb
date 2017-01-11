package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var buildCmdConfig buildConfig

// buildCmd represents the new command
var buildCmd = &cobra.Command{
	Use:   "new <import_path> [flags]",
	Short: "Generate a new ABCWeb application.",
	Long: `The 'abcweb new' command generates a new ABCWeb application with a 
default directory structure and configuration at the Go src import path you specify.

The app will generate in $GOPATH/src/<import_path>.
`,
	Example: "abcweb new github.com/yourusername/myapp",
	PreRunE: buildCmdPreRun,
	RunE:    buildCmdRun,
}

func init() {
	buildCmd.Flags().StringP("sessions-prod-storer", "p", "disk", "Session storer to use in production mode")
	buildCmd.Flags().BoolP("silent", "", false, "Disable console output")

	RootCmd.AddCommand(buildCmd)
	viper.BindPFlags(buildCmd.Flags())
}

func buildCmdPreRun(cmd *cobra.Command, args []string) error {
	var err error

	//	buildCmdConfig = newConfig{
	//		NoGitIgnore: viper.GetBool("no-gitignore"),
	//		DefaultEnv:  viper.GetString("default-env"),
	//		Bootstrap:   strings.ToLower(viper.GetString("bootstrap")),
	//	}

	return err
}

func buildCmdRun(cmd *cobra.Command, args []string) error {
	return nil
}
