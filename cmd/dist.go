package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var distCmdConfig distConfig

// distCmd represents the dist command
var distCmd = &cobra.Command{
	Use:   "dist",
	Short: "Dist creates a distribution bundle for easy deployment",
	Long: `Dist compiles your binary, builds your assets, and copies
the binary, assets, templates and migrations into a dist folder
for easy deployment to your production server. It can also
optionally zip your dist bundle for a single file deploy, and copy
over or generate new config files.`,
	Example: "abcweb dist",
	RunE:    distCmdRun,
}

func init() {
	distCmd.Flags().BoolP("config", "c", false, "Generate fresh config files in dist package")
	distCmd.Flags().BoolP("copy-config", "", false, "Copy all .toml files from app root to dist package")
	distCmd.Flags().BoolP("zip", "z", false, "Zip dist package once completed")
	distCmd.Flags().BoolP("no-migrations", "", false, "Skip inclusion of migrations folder")
	distCmd.Flags().BoolP("no-assets", "", false, "Skip inclusion of public assets folder")
	distCmd.Flags().BoolP("no-templates", "", false, "Skip inclusion of templates folder")

	RootCmd.AddCommand(distCmd)
}

func distCmdRun(cmd *cobra.Command, args []string) error {
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
