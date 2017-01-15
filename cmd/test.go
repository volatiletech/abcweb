package cmd

import "github.com/spf13/cobra"

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Runs the tests for your ABCWeb app",
	RunE:  testCmdRun,
}

// TestInit initializes the build commands and flags
func TestInit() {
	RootCmd.AddCommand(testCmd)
}

func testCmdRun(cmd *cobra.Command, args []string) error {
	return nil
}
