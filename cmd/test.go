package cmd

import "github.com/spf13/cobra"

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Runs the tests for your ABCWeb app",
	RunE:  testCmdRun,
}

func init() {
	RootCmd.AddCommand(testCmd)
}

func testCmdRun(cmd *cobra.Command, args []string) error {
	return nil
}
