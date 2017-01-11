package cmd

import (
	"os"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// AppFS is a handle to the filesystem in use
var fs = afero.NewOsFs()

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:  "abcweb",
	Long: `ABCWeb is a tool to help you scaffold and develop Go web applications.`,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}

func init() {
	RootCmd.Flags().BoolP("version", "", false, "Print the ABCWeb version")
	viper.BindPFlags(RootCmd.Flags())
}
