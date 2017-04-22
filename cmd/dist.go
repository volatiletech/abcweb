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
	Use:   "dist",
	Short: "Dist creates a distribution bundle for deployment",
	Long: `Dist compiles your binary, builds your assets, and copies
the binary, assets, templates and migrations into a dist folder
for easy deployment to your production server. It can also
optionally zip your dist bundle for a single file deploy.`,
	Example: "abcweb build",
	RunE:    buildCmdRun,
}

func init() {
	buildCmd.Flags().BoolP("config", "c", false, "Generate fresh config files in dist package")

	RootCmd.AddCommand(buildCmd)
}

func buildCmdRun(cmd *cobra.Command, args []string) error {
	cnf.ModeViper.BindPFlags(cmd.Flags())

	// Bare minimum requires git and go dependencies
	checkDep("git", "go")

	fmt.Println("Building assets...")
	if err := buildAssets(); err != nil {
		return err
	}

	fmt.Println("Building Go app...")
	if err := buildApp(); err != nil {
		return err
	}

	return nil
}

func buildApp() error {
	cmd := exec.Command("go", "build")
	cmd.Dir = cnf.AppPath

	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		return err
	}

	fmt.Printf("Build app complete.\n\n")
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

	fmt.Printf("Build assets complete.\n\n")
	return nil
}
