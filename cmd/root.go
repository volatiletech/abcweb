package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/volatiletech/abcweb/config"
)

var (
	cnf *config.Configuration

	appFS = afero.NewOsFs()
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:  "abcweb",
	Long: `ABCWeb is a tool to help you scaffold and develop Go web applications.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		cnf, err = config.Initialize(nil)
		if err != nil {
			return err
		}
		return nil
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func checkDep(names ...string) {
	for _, name := range names {
		_, err := exec.LookPath(name)
		if err != nil {
			fmt.Printf("Error: could not find %q dependency in $PATH. Please run \"abcweb deps\" to install all missing dependencies.", name)
			os.Exit(1)
		}
	}
}

func init() {
	RootCmd.Flags().BoolP("version", "", false, "Print the abcweb version")
	viper.BindPFlags(RootCmd.Flags())
}
