package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var buildCmdConfig buildConfig

// buildCmd represents the new command
var buildCmd = &cobra.Command{
	Use:     "build",
	Short:   "Builds your ABCWeb app binary and.... does stuff to assets?",
	Long:    "describe build command here",
	Example: "build example here",
	PreRunE: buildCmdPreRun,
	RunE:    buildCmdRun,
}

func init() {
	buildCmd.Flags().StringP("sessions-prod-storer", "p", "disk", "Session storer to use in production mode")
	buildCmd.Flags().BoolP("silent", "", false, "Disable console output")

	RootCmd.AddCommand(buildCmd)
}

func buildCmdPreRun(cmd *cobra.Command, args []string) error {
	var err error

	viper.BindPFlags(cmd.Flags())

	//	buildCmdConfig = newConfig{
	//		NoGitIgnore: viper.GetBool("no-gitignore"),
	//		DefaultEnv:  viper.GetString("env"),
	//		Bootstrap:   strings.ToLower(viper.GetString("bootstrap")),
	//	}

	return err
}

func buildCmdRun(cmd *cobra.Command, args []string) error {
	return nil
}
