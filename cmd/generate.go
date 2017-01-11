package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var modelsCmdConfig modelsConfig

// generateCmd represents the "generate" command
var generateCmd = &cobra.Command{
	Use:     "generate",
	Short:   "Generate your models and migration files",
	Long:    "stuff here",
	Example: "stuff here",
}

// modelsCmd represents the "generate models" command
var modelsCmd = &cobra.Command{
	Use:     "models",
	Short:   "stuff here",
	Long:    "models cmd description",
	Example: "whatever",
	PreRunE: modelsCmdPreRun,
	RunE:    modelsCmdRun,
}

// migrationCmd represents the "generate migration" command
var migrationCmd = &cobra.Command{
	Use:     "migration <name> [flags]",
	Short:   "stuff here",
	Long:    "Generate a migration file.",
	Example: "abcweb generate migration add_users",
	RunE:    migrationCmdRun,
}

func init() {
	// models flags
	modelsCmd.Flags().StringP("db", "b", "", `Valid options: (postgres|mysql) (default: config.toml "db" field value)`)
	modelsCmd.Flags().StringP("env", "e", "dev", `config.toml environment to load (default: will only use "dev" default if cannot find in $PROJECTNAME_ENV)`)

	// migration flags
	migrationCmd.Flags().BoolP("sql", "s", false, "Generate an .sql migration instead of a .go migration")
	migrationCmd.Flags().StringP("dir", "d", migrationsDirectory, "Directory with migration files")

	RootCmd.AddCommand(generateCmd)

	// Add generate subcommands
	generateCmd.AddCommand(modelsCmd)
	generateCmd.AddCommand(migrationCmd)

	viper.BindPFlags(generateCmd.Flags())
}

func modelsCmdPreRun(cmd *cobra.Command, args []string) error {
	var err error

	modelsCmdConfig = modelsConfig{}

	return err
}

func modelsCmdRun(cmd *cobra.Command, args []string) error {
	return nil
}

func migrationCmdRun(cmd *cobra.Command, args []string) error {
	return nil
}
