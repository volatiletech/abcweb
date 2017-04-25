package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var buildCmdConfig buildConfig

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:     "build",
	Short:   "Builds your abcweb binary and executes the gulp build task",
	Long:    "Builds your abcweb binary and executes the gulp build task",
	Example: "abcweb build",
	RunE:    buildCmdRun,
}

func init() {
	buildCmd.Flags().BoolP("go-only", "g", false, "Only build the go binary")
	buildCmd.Flags().BoolP("assets-only", "a", false, "Only build the assets")
	buildCmd.Flags().BoolP("verbose", "v", false, "Verbose output")

	RootCmd.AddCommand(buildCmd)
}

func buildCmdRun(cmd *cobra.Command, args []string) error {
	cnf.ModeViper.BindPFlags(cmd.Flags())

	// Bare minimum requires git and go dependencies
	checkDep("git", "go")

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
	args := []string{"build"}

	if cnf.ModeViper.GetBool("verbose") {
		args = append(args, "-v")
	}

	var version string
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	out := &bytes.Buffer{}
	cmd.Stdout = out
	_ = cmd.Run()

	version = strings.TrimSpace(out.String())
	if len(version) == 0 {
		version = "0"
	}

	buildTime := time.Now().Format(time.RFC3339)

	args = append(args, "-ldflags", fmt.Sprintf("-X main.version=%s -X main.buildTime=%s", version, buildTime))

	cmd = exec.Command("go", args...)
	cmd.Dir = cnf.AppPath
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		return err
	}

	fmt.Printf("SUCCESS.\n\n")
	return nil
}

func buildAssets() error {
	_, err := os.Stat(filepath.Join(cnf.AppPath, "gulpfile.js"))
	if os.IsNotExist(err) {
		fmt.Println("No gulpfile.js present, skipping gulp build.")
		return nil
	} else if err != nil {
		return err
	}

	_, err = exec.LookPath("gulp")
	if err != nil {
		fmt.Println("Cannot find gulp in PATH. Is it installed? Skipping gulp build.")
		return nil
	}

	cmd := exec.Command("gulp", "build")

	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err = cmd.Run(); err != nil {
		return err
	}

	fmt.Printf("SUCCESS.\n\n")
	return nil
}
