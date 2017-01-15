package cmd

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/kat-co/vala"
	"github.com/nullbio/abcweb/config"
	"github.com/nullbio/abcweb/strmangle"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vattle/sqlboiler/bdb/drivers"
	"github.com/vattle/sqlboiler/boilingcore"
)

var modelsCmdConfig boilingcore.Config
var modelsCmdState *boilingcore.State

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
This tool pipes out to SQLBoiler: https://github.com/vattle/sqlboiler
See README.md at sqlboiler repo for API guidance.`,
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
	envViper := viper.Sub(config.ActiveEnv)
	envViper.AddConfigPath(filepath.Join(config.AppPath, config.DBTomlFilename))
	envViper.SetEnvPrefix(strmangle.EnvAppName(config.GetAppName(config.AppPath)))
	envViper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// models flags
	modelsCmd.Flags().StringP("env", "e", "", `database.toml environment to load, obtained from config.toml default_env or $YOURPROJECTNAME_ENV`)
	modelsCmd.Flags().StringP("db", "b", "", `Valid options: (postgres|mysql) (default "database.toml db field")`)
	modelsCmd.Flags().StringP("schema", "s", "", `The name of your database schema, for databases that support real schemas (default "public"`)
	modelsCmd.Flags().StringP("basedir", "", "", "The base directory has the templates and templates_test folders")
	modelsCmd.Flags().BoolP("debug", "d", false, "Debug mode prints stack traces on error")
	modelsCmd.Flags().BoolP("no-tests", "", false, "Disable generated go test files")
	modelsCmd.Flags().BoolP("no-hooks", "", false, "Disable hooks feature for your models")
	modelsCmd.Flags().BoolP("no-auto-timestamps", "", false, "Disable automatic timestamps for created_at/updated_at")
	modelsCmd.Flags().BoolP("tinyint-not-bool", "", false, "Map MySQL tinyint(1) in Go to int8 instead of bool")
	modelsCmd.Flags().StringSliceP("blacklist", "b", nil, "Do not include these tables in your generated package")
	modelsCmd.Flags().StringSliceP("whitelist", "w", nil, "Only include these tables in your generated package")
	modelsCmd.Flags().StringSliceP("tag", "t", nil, "Struct tags to be included on your models in addition to json, yaml, toml")

	// migration flags
	migrationCmd.Flags().BoolP("sql", "s", false, "Generate an .sql migration instead of a .go migration")
	migrationCmd.Flags().StringP("dir", "d", migrationsDirectory, "Directory with migration files")

	RootCmd.AddCommand(generateCmd)

	// Add generate subcommands
	generateCmd.AddCommand(modelsCmd)
	generateCmd.AddCommand(migrationCmd)

	viper.BindPFlags(generateCmd.Flags())
}

// modelsCmdPreRun sets up the modelsCmdState and modelsCmdConfig objects
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

	// Override string fields if they exist as cmd line args
	dbStr := viper.GetString("db")
	if len(dbStr) > 0 {
		dbConfig.DB = dbStr
	}
	// error if no db is provided
	if len(dbConfig.DB) == 0 {
		return errors.New("must provide a driver name")
	}
	schemaStr := viper.GetString("schema")
	if len(schemaStr) > 0 {
		dbConfig.Schema = schemaStr
	}
	basedirStr := viper.GetString("basedir")
	if len(basedirStr) > 0 {
		dbConfig.BaseDir = basedirStr
	}

	// Override string slice fields if they exist as cmd line args
	blacklistStrs := viper.GetStringSlice("blacklist")
	if len(blacklistStrs) == 0 {
		dbConfig.Blacklist = blacklistStrs
	}
	whitelistStrs := viper.GetStringSlice("whitelist")
	if len(whitelistStrs) == 0 {
		dbConfig.Whitelist = whitelistStrs
	}
	tagStrs := viper.GetStringSlice("tag")
	if len(tagStrs) == 0 {
		dbConfig.Tag = tagStrs
	}

	// Some database specific defaults
	if dbConfig.DB == "postgres" {
		if len(dbConfig.SSLMode) == 0 {
			dbConfig.SSLMode = "require"
		}
		if dbConfig.Port == 0 {
			dbConfig.Port = 5432
		}
	} else if dbConfig.DB == "mysql" {
		if len(dbConfig.SSLMode) == 0 {
			dbConfig.SSLMode = "true"
		}
		if dbConfig.Port == 0 {
			dbConfig.Port = 3306
		}
	}

	// If bools aren't default value of "false" then set them as true
	if viper.GetBool("debug") != false {
		dbConfig.Debug = true
	}
	if viper.GetBool("no-tests") != false {
		dbConfig.NoTests = true
	}
	if viper.GetBool("no-hooks") != false {
		dbConfig.NoHooks = true
	}
	if viper.GetBool("no-auto-timestamps") != false {
		dbConfig.NoAutoTimestamps = true
	}
	if viper.GetBool("tinyint-not-bool") != false {
		dbConfig.TinyintNotBool = true
	}

	modelsCmdConfig = &boilingcore.Config{
		DriverName:       dbConfig.DB,
		OutFolder:        dbConfig.Output,
		Schema:           dbConfig.Schema,
		PkgName:          dbConfig.PkgName,
		BaseDir:          dbConfig.BaseDir,
		Debug:            dbConfig.Debug,
		NoTests:          dbConfig.NoTests,
		NoHooks:          dbConfig.NoHooks,
		NoAutoTimestamps: dbConfig.NoAutoTimestamps,
	}

	// BUG: https://github.com/spf13/viper/issues/200
	// Look up the value of blacklist, whitelist & tags directly from PFlags in Cobra if we
	// detect a malformed value coming out of viper.
	modelsCmdConfig.BlacklistTables = viper.GetStringSlice("blacklist")
	if len(modelsCmdConfig.BlacklistTables) == 1 && strings.ContainsRune(modelsCmdConfig.BlacklistTables[0], ',') {
		modelsCmdConfig.BlacklistTables, err = cmd.PersistentFlags().GetStringSlice("blacklist")
		if err != nil {
			return err
		}
	}

	modelsCmdConfig.WhitelistTables = viper.GetStringSlice("whitelist")
	if len(modelsCmdConfig.WhitelistTables) == 1 && strings.ContainsRune(modelsCmdConfig.WhitelistTables[0], ',') {
		modelsCmdConfig.WhitelistTables, err = cmd.PersistentFlags().GetStringSlice("whitelist")
		if err != nil {
			return err
		}
	}

	modelsCmdConfig.Tags = viper.GetStringSlice("tag")
	if len(modelsCmdConfig.Tags) == 1 && strings.ContainsRune(modelsCmdConfig.Tags[0], ',') {
		modelsCmdConfig.Tags, err = cmd.PersistentFlags().GetStringSlice("tag")
		if err != nil {
			return err
		}
	}

	if driverName == "postgres" {
		modelsCmdConfig.Postgres = boilingcore.PostgresConfig{
			User:    viper.GetString("postgres.user"),
			Pass:    viper.GetString("postgres.pass"),
			Host:    viper.GetString("postgres.host"),
			Port:    viper.GetInt("postgres.port"),
			DBName:  viper.GetString("postgres.dbname"),
			SSLMode: viper.GetString("postgres.sslmode"),
		}

		// BUG: https://github.com/spf13/viper/issues/71
		// Despite setting defaults, nested values don't get defaults
		// Set them manually
		if modelsCmdConfig.Postgres.SSLMode == "" {
			modelsCmdConfig.Postgres.SSLMode = "require"
			viper.Set("postgres.sslmode", modelsCmdConfig.Postgres.SSLMode)
		}

		if modelsCmdConfig.Postgres.Port == 0 {
			modelsCmdConfig.Postgres.Port = 5432
			viper.Set("postgres.port", modelsCmdConfig.Postgres.Port)
		}

		err = vala.BeginValidation().Validate(
			vala.StringNotEmpty(modelsCmdConfig.Postgres.User, "postgres.user"),
			vala.StringNotEmpty(modelsCmdConfig.Postgres.Host, "postgres.host"),
			vala.Not(vala.Equals(modelsCmdConfig.Postgres.Port, 0, "postgres.port")),
			vala.StringNotEmpty(modelsCmdConfig.Postgres.DBName, "postgres.dbname"),
			vala.StringNotEmpty(modelsCmdConfig.Postgres.SSLMode, "postgres.sslmode"),
		).Check()

		if err != nil {
			return commandFailure(err.Error())
		}
	}

	if driverName == "mysql" {
		modelsCmdConfig.MySQL = boilingcore.MySQLConfig{
			User:    viper.GetString("mysql.user"),
			Pass:    viper.GetString("mysql.pass"),
			Host:    viper.GetString("mysql.host"),
			Port:    viper.GetInt("mysql.port"),
			DBName:  viper.GetString("mysql.dbname"),
			SSLMode: viper.GetString("mysql.sslmode"),
		}

		// Set MySQL TinyintAsBool global var. This flag only applies to MySQL.
		drivers.TinyintAsBool = viper.GetBool("tinyint-as-bool")

		// MySQL doesn't have schemas, just databases
		modelsCmdConfig.Schema = modelsCmdConfig.MySQL.DBName

		// BUG: https://github.com/spf13/viper/issues/71
		// Despite setting defaults, nested values don't get defaults
		// Set them manually
		if modelsCmdConfig.MySQL.SSLMode == "" {
			modelsCmdConfig.MySQL.SSLMode = "true"
			viper.Set("mysql.sslmode", modelsCmdConfig.MySQL.SSLMode)
		}

		if modelsCmdConfig.MySQL.Port == 0 {
			modelsCmdConfig.MySQL.Port = 3306
			viper.Set("mysql.port", modelsCmdConfig.MySQL.Port)
		}

		err = vala.BeginValidation().Validate(
			vala.StringNotEmpty(modelsCmdConfig.MySQL.User, "mysql.user"),
			vala.StringNotEmpty(modelsCmdConfig.MySQL.Host, "mysql.host"),
			vala.Not(vala.Equals(modelsCmdConfig.MySQL.Port, 0, "mysql.port")),
			vala.StringNotEmpty(modelsCmdConfig.MySQL.DBName, "mysql.dbname"),
			vala.StringNotEmpty(modelsCmdConfig.MySQL.SSLMode, "mysql.sslmode"),
		).Check()

		if err != nil {
			return err
		}
	}

	modelsCmdState, err = boilingcore.New(modelsCmdConfig)

	return err
}

func modelsCmdRun(cmd *cobra.Command, args []string) error {
	var boilArgs []string

	// First arg to sqlboiler is database
	boilArgs = append(boilArgs, modelsCmdConfig.DB)

	// Append flags if they're not empty
	if len(modelsCmdConfig.Schema) > 0 {
		boilArgs = append(boilArgs, "--schema", modelsCmdConfig.Schema)
	}
	if len(modelsCmdConfig.BaseDir) > 0 {
		boilArgs = append(boilArgs, "--basedir", modelsCmdConfig.BaseDir)
	}
	if len(modelsCmdConfig.Output) > 0 {
		boilArgs = append(boilArgs, "--output", modelsCmdConfig.Output)
	}
	if len(modelsCmdConfig.Schema) > 0 {
		boilArgs = append(boilArgs, "--schema", modelsCmdConfig.Schema)
	}

	// TODO
	// strings join array types using comma to pass in blacklist, whitelist etc
	// will have to make sure sqlboiler can parse this first

	// Append bool flags
	if modelsCmdConfig.Debug {
		boilArgs = append(boilArgs, "--debug")
	}
	if modelsCmdConfig.NoAutoTimestamps {
		boilArgs = append(boilArgs, "--no-auto-timestamps")
	}
	if modelsCmdConfig.NoHooks {
		boilArgs = append(boilArgs, "--no-hooks")
	}
	if modelsCmdConfig.NoTests {
		boilArgs = append(boilArgs, "--no-tests")
	}
	// sqlboilers field is "as" bool instead of "not" bool, so invert value
	if !modelsCmdConfig.TinyintNotBool {
		boilArgs = append(boilArgs, "--tinyint-as-bool")
	}

	// going to have to set env vars for db connection flags

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
