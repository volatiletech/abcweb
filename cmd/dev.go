package cmd

import "github.com/spf13/cobra"

// devCmd represents the "dev" command
var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Runs your ABCWeb app for development",
	Long: `Runs your ABCWeb app for development. The dev command has a watch component
that watches for any changes to Go files and re-builds/re-runs your app on change. 
It will also watch for any change to assets and execute the corresponding gulp tasks 
for the changed asset file(s). The dev command also hosts a live-reload server
and will notify any connected live-reload clients and/or browsers to reload
the page once the corresponding watcher task is finished.`,
	Example: `abcweb dev`,
	RunE:    devCmdRun,
}

func init() {
	RootCmd.AddCommand(devCmd)
}

func devCmdRun(cmd *cobra.Command, args []string) error {
	return nil
}
