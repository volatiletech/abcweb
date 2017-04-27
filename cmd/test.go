package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:     "test",
	Short:   "Runs the tests for your abcweb app",
	Long:    `Runs all Go tests for your abcweb app and skips your vendor directory.`,
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
	checkDep("go")

	var exc *exec.Cmd

	exc = exec.Command("go", "list", "./...")
	exc.Dir = cnf.AppPath
	out, err := exc.CombinedOutput()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if cnf.ModeViper.GetBool("verbose") {
		exc = exec.Command("go", "test", "-v", "-p", "1")
	} else {
		exc = exec.Command("go", "test", "-p", "1")
	}

	rgx, err := regexp.Compile(cnf.AppName + "/vendor")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	pkgs := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, pkg := range pkgs {
		if !rgx.MatchString(pkg) {
			exc.Args = append(exc.Args, pkg)
		}
	}

	exc.Dir = cnf.AppPath
	out, err = exc.CombinedOutput()
	fmt.Print(string(out))

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
