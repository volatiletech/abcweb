package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// watchCmd represents the watch command
var watchCmd = &cobra.Command{
	Use:     "watch [flags]",
	Short:   "short desc.",
	Long:    `The 'abcweb watch' command does............`,
	Example: `abcweb watch`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Work your own magic here
		fmt.Println("watch called")
	},
}

func init() {
	RootCmd.AddCommand(watchCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// watchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// watchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
