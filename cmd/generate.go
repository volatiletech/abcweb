package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var modelsCmdConfig modelsConfig

// generateCmd represents the "generate" command
var generateCmd = &cobra.Command{
	Use:     "generate",
	Short:   "Generate your database models and migration files",
	Example: "abcweb generate models\nabcweb generate migration add_users",
}

// modelsCmd represents the "generate models" command
var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "Generate your database models",
	Long: `Generate models will connect to your database and generate 
your models from your existing database structure. Make sure you run your
migrations first.`,
	Example: "abcweb generate models",
	PreRunE: modelsCmdPreRun,
	RunE:    modelsCmdRun,
}

// migrationCmd represents the "generate migration" command
var migrationCmd = &cobra.Command{
	Use:   "migration <name> [flags]",
	Short: "Generate a migration file",
	Long: `Generate migration will generate a .go or .sql migration file in 
your migrations directory.`,
	Example: "abcweb generate migration add_users",
	RunE:    migrationCmdRun,
}

func init() {
	// models flags
	modelsCmd.Flags().StringP("env", "e", "dev", `database.toml environment to load, obtained from config.toml default_env or $YOURPROJECTNAME_ENV`)
	modelsCmd.Flags().StringP("db", "b", "", `Valid options: (postgres|mysql) (default "database.toml db field")`)

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
