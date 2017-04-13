package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var buildCmdConfig buildConfig

// buildCmd represents the new command
var buildCmd = &cobra.Command{
	Use:     "build",
	Short:   "Builds your abcweb binary and executes the gulp build task",
	Long:    "Builds your abcweb binary and executes the gulp build task",
	Example: "abcweb build",
	RunE:    buildCmdRun,
}

func init() {
	buildCmd.Flags().BoolP("go-only", "g", false, "Only watch and rebuild the go app by piping to refresh app")
	buildCmd.Flags().BoolP("assets-only", "a", false, "Only watch and build the assets by piping to gulp watch")

	RootCmd.AddCommand(buildCmd)
}

func buildCmdRun(cmd *cobra.Command, args []string) error {
	cnf.ModeViper.BindPFlags(cmd.Flags())

	if !cnf.ModeViper.GetBool("go-only") {
		fmt.Println("Building assets...")
		if err := buildAssets(); err != nil {
			return err
		}
	}

	if !cnf.ModeViper.GetBool("assets-only") {
		fmt.Println("Building Go app...")
		if err := buildApp(); err != nil {
			return err
		}
	}

	return nil
}

func buildApp() error {
	_, err := exec.LookPath("go")
	if err != nil {
		fmt.Println("Cannot Go in PATH.")
		return nil
	}

	cmd := exec.Command("go", "build")
	cmd.Dir = cnf.AppPath

	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err = cmd.Run(); err != nil {
		return err
	}

	fmt.Printf("Build app complete.\n\n")
	return nil
}

func buildAssets() error {
	_, err := os.Stat(filepath.Join(cnf.AppPath, "gulpfile.js"))
	if os.IsNotExist(err) {
		fmt.Println("No gulpfile.js present, skipping gulp build")
		return nil
	} else if err != nil {
		return err
	}

	_, err = exec.LookPath("gulp")
	if err != nil {
		fmt.Println("Cannot find gulp in PATH. Is it installed?")
		return nil
	}

	cmd := exec.Command("gulp", "build")

	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err = cmd.Run(); err != nil {
		return err
	}

	fmt.Printf("Build assets complete.\n\n")
	return nil
}
