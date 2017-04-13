package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Runs the tests for your abcweb app",
	Long: `Runs all Go tests for your abcweb app by executing "go test ./..."
from the root directory of your app.`,
	Example: "abcweb test -v",
	Run:     testCmdRun,
}

func init() {
	testCmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")

	RootCmd.AddCommand(testCmd)

	testCmd.PreRun = func(*cobra.Command, []string) {
		cnf.ModeViper.BindPFlags(testCmd.Flags())
	}
}

func testCmdRun(cmd *cobra.Command, args []string) {
	var exc *exec.Cmd

	// Set the APPNAME_ENV flag to the env value in case it was obtained
	// through the config file instead of through an environment variable.
	// This is required so that the models tests know what environment to load.
	// os.Setenv(strmangle.EnvAppName(cnf.GetAppName(cnf.AppPath))+"_ENV", cnf.ActiveEnv)

	if cnf.ModeViper.GetBool("verbose") {
		exc = exec.Command("go", "test", "./...", "-v")
	} else {
		exc = exec.Command("go", "test", "./...")
	}

	exc.Dir = cnf.AppPath
	out, err := exc.CombinedOutput()
	fmt.Print(string(out))

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
