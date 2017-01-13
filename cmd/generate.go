package cmd

import (
	"fmt"

	"github.com/nullbio/abcweb/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var modelsCmdConfig modelsConfig
var migrationCmdConfig migrateConfig

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
your models from your existing database structure. 
Make sure you run your migrations first.
This tool pipes out to SQLBoiler: https://github.com/vattle/sqlboiler`,
	Example: "abcweb generate models",
	PreRunE: modelsCmdPreRun,
	RunE:    modelsCmdRun,
}

// migrationCmd represents the "generate migration" command
var migrationCmd = &cobra.Command{
	Use:   "migration <name> [flags]",
	Short: "Generate a migration file",
	Long: `Generate migration will generate a .go or .sql migration file in your migrations directory.
This tool pipes out to Goose: https://github.com/pressly/goose`,
	Example: "abcweb generate migration add_users",
	PreRunE: migrationCmdPreRun,
	RunE:    migrationCmdRun,
}

func init() {
	// models flags
	modelsCmd.Flags().StringP("env", "e", "", `database.toml environment to load, obtained from config.toml default_env or $YOURPROJECTNAME_ENV`)
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
	env := config.ActiveEnv

	// Override env string if it exists as a cmd line arg
	envStr := viper.GetString("env")
	if len(envStr) > 0 {
		env = envStr
	}

	// If no env mode is found in config.toml, $APPNAME_ENV OR on command line
	// then fall back to a default value of "dev"
	if len(env) == 0 {
		fmt.Printf(`No environment mode could be found, attempting fallback of "dev"`)
		env = "dev"
	}

	dbConfig := config.LoadDBConfig(config.AppPath, env)

	// Override DB field if it exists as a cmd line arg
	dbStr := viper.GetString("db")
	if len(dbStr) > 0 {
		dbConfig.DB = dbStr
	}

	modelsCmdConfig = modelsConfig{
		DBConfig: dbConfig,
	}

	return err
}

func modelsCmdRun(cmd *cobra.Command, args []string) error {
	return nil
}

func migrationCmdPreRun(cmd *cobra.Command, args []string) error {
	var err error

	env := config.ActiveEnv

	// Override env string if it exists as a cmd line arg
	envStr := viper.GetString("env")
	if len(envStr) > 0 {
		env = envStr
	}

	// If no env mode is found in config.toml, $APPNAME_ENV OR on command line
	// then fall back to a default value of "dev"
	if len(env) == 0 {
		fmt.Printf(`No environment mode could be found, attempting fallback of "dev"`)
		env = "dev"
	}

	dbConfig := config.LoadDBConfig(config.AppPath, env)

	// Override DB field if it exists as a cmd line arg
	dbStr := viper.GetString("db")
	if len(dbStr) > 0 {
		dbConfig.DB = dbStr
	}

	// Override MigrationsDir field if it exists as a cmd line arg
	dirStr := viper.GetString("dir")
	if len(dirStr) > 0 {
		dbConfig.MigrationsDir = dirStr
	}

	migrationCmdConfig = migrateConfig{
		DBConfig: dbConfig,
	}

	return err
}

func migrationCmdRun(cmd *cobra.Command, args []string) error {
	return nil
}
