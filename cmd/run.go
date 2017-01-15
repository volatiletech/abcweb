package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// runCmd represents the "run" command
var runCmd = &cobra.Command{
	Use:     "run",
	Short:   "Runs your ABCWeb app",
	Long:    "Runs your ABCWeb app, watches files, etc.....",
	Example: `abcweb run`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Work your own magic here
		fmt.Println("run called")
	},
}

// RunInit initializes the build commands and flags
func RunInit() {
	RootCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// watchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// watchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
