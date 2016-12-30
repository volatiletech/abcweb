package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use: "version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("ABCWeb version 1.0.0")
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
