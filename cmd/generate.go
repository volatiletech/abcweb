package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/kat-co/vala"
	"github.com/spf13/cobra"
	"github.com/volatiletech/sqlboiler/bdb/drivers"
	"github.com/volatiletech/sqlboiler/boilingcore"
	"github.com/volatiletech/abcweb/config"
)

// generateCmd represents the "generate" command
var generateCmd = &cobra.Command{
	Use:     "gen",
	Short:   "Generate your database models and migration files",
	Example: "abcweb gen models\nabcweb gen migration add_users",
}

// modelsCmd represents the "generate models" command
var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "Generate your database models",
	Long: `Generate models will connect to your database and generate your models from your existing database structure.
Don't forget to run your migrations.

This tool pipes out to SQLBoiler: https://github.com/volatiletech/sqlboiler -- See README.md at sqlboiler repo for API guidance.`,
	Example: "abcweb gen models",
	RunE:    modelsCmdRun,
}

// migrationCmd represents the "generate migration" command
var migrationCmd = &cobra.Command{
	Use:   "migration <name> [flags]",
	Short: "Generate a migration file",
	Long: `Generate migration will generate a .sql migration file in your db/migrations directory.
This tool pipes out to mig: https://github.com/volatiletech/mig`,
	Example: "abcweb gen migration add_users",
	RunE:    migrationCmdRun,
}

var configCmd = &cobra.Command{
	Use:     "config",
	Short:   "Generate fresh config files",
	Long:    "Generate fresh config files",
	Example: "abcweb gen config",
	RunE:    configCmdRun,
}

var certsCmd = &cobra.Command{
	Use:     "certs",
	Short:   "Generate fresh TLS certificates",
	Long:    "Generate fresh TLS certificates",
	Example: "abcweb gen certs",
	RunE:    certsCmdRun,
}

// The custom SQLBoiler template file replacements
var replaceFiles = [][]string{
	{"templates_test/main_test/mysql_main.tpl", "sqlboiler/mysql_main.tmpl"},
	{"templates_test/main_test/postgres_main.tpl", "sqlboiler/postgres_main.tmpl"},
	{"templates_test/singleton/boil_main_test.tpl", "sqlboiler/boil_main_test.tmpl"},
}

func init() {
	basepath, err := config.GetBasePath()
	if err != nil {
		panic(fmt.Sprintf("unable to get abcweb base path: %s", err))
	}

	replaceArgs := make([]string, len(replaceFiles))

	// Prefix the replaceWith file with the basepath
	for i := 0; i < len(replaceFiles); i++ {
		replaceArgs[i] = fmt.Sprintf("%s:%s", replaceFiles[i][0], filepath.Join(basepath, replaceFiles[i][1]))
	}

	// models flags
	modelsCmd.Flags().StringP("db", "", "", `Valid options: postgres|mysql (default "config.toml db field")`)
	modelsCmd.Flags().StringP("schema", "s", "public", "The name of your database schema, for databases that support real schemas")
	modelsCmd.Flags().StringP("pkgname", "p", "models", "The name you wish to assign to your generated package")
	modelsCmd.Flags().StringP("output", "o", filepath.FromSlash("db/models"), "The name of the folder to output to. Automatically created relative to webapp root dir")
	modelsCmd.Flags().StringP("basedir", "", "", "The base directory has the templates and templates_test folders")
	modelsCmd.Flags().StringSliceP("blacklist", "b", nil, "Do not include these tables in your generated package")
	modelsCmd.Flags().StringSliceP("whitelist", "w", nil, "Only include these tables in your generated package")
	modelsCmd.Flags().StringSliceP("tag", "t", nil, "Struct tags to be included on your models in addition to json, yaml, toml")
	modelsCmd.Flags().StringSliceP("replace", "", replaceArgs, "Replace templates by directory: relpath/to_file.tpl:relpath/to_replacement.tpl")
	modelsCmd.Flags().BoolP("debug", "d", false, "Debug mode prints stack traces on error")
	modelsCmd.Flags().BoolP("no-tests", "", false, "Disable generated go test files")
	modelsCmd.Flags().BoolP("no-hooks", "", false, "Disable hooks feature for your models")
	modelsCmd.Flags().BoolP("no-auto-timestamps", "", false, "Disable automatic timestamps for created_at/updated_at")
	modelsCmd.Flags().BoolP("tinyint-not-bool", "", false, "Map MySQL tinyint(1) in Go to int8 instead of bool")
	modelsCmd.Flags().BoolP("wipe", "", true, "Delete the output folder (rm -rf) before generation to ensure sanity")
	modelsCmd.Flags().StringP("env", "e", "dev", "The config.toml file environment to load")

	// hide flags not recommended for use
	modelsCmd.Flags().MarkHidden("replace")

	// config flags
	configCmd.Flags().BoolP("force", "f", false, "Overwrite files if they already exist")

	// certs flags
	certsCmd.Flags().BoolP("force", "f", false, "Overwrite files if they already exist")

	RootCmd.AddCommand(generateCmd)

	// hook up pre-run hooks, this avoids initialization loops
	modelsCmd.PreRunE = modelsCmdPreRun
	migrationCmd.PreRun = migrationCmdPreRun
	configCmd.PreRun = configCmdPreRun
	certsCmd.PreRun = certsCmdPreRun

	// Add generate subcommands
	generateCmd.AddCommand(modelsCmd)
	generateCmd.AddCommand(migrationCmd)
	generateCmd.AddCommand(configCmd)
	generateCmd.AddCommand(certsCmd)
}

// migrationCmdPreRun sets up the flag bindings
func migrationCmdPreRun(*cobra.Command, []string) {
	cnf.ModeViper.BindPFlags(migrationCmd.Flags())
}

// modelsCmdPreRun sets up the modelsCmdState and modelsCmdConfig objects
func modelsCmdPreRun(cmd *cobra.Command, args []string) error {
	var err error
	cnf, err = config.Initialize(cmd.Flags().Lookup("env"))
	if err != nil {
		return err
	}

	cnf.ModeViper.BindPFlags(modelsCmd.Flags())

	err = cnf.CheckEnv()
	if err != nil {
		return err
	}

	return modelsCmdSetup(cmd, args)
}

func modelsCmdSetup(cmd *cobra.Command, args []string) error {
	var err error
	modelsCmdConfig = &boilingcore.Config{
		DriverName:       cnf.ModeViper.GetString("db"),
		OutFolder:        filepath.Join(cnf.AppPath, cnf.ModeViper.GetString("output")),
		Schema:           cnf.ModeViper.GetString("schema"),
		PkgName:          cnf.ModeViper.GetString("pkgname"),
		BaseDir:          cnf.ModeViper.GetString("basedir"),
		Debug:            cnf.ModeViper.GetBool("debug"),
		NoTests:          cnf.ModeViper.GetBool("no-tests"),
		NoHooks:          cnf.ModeViper.GetBool("no-hooks"),
		NoAutoTimestamps: cnf.ModeViper.GetBool("no-auto-timestamps"),
		Wipe:             cnf.ModeViper.GetBool("wipe"),
	}

	// BUG: https://github.com/spf13/viper/pull/296
	// Look up the value of blacklist, whitelist & tags directly from PFlags in Cobra if we
	// detect a malformed value coming out of viper.
	modelsCmdConfig.BlacklistTables = cnf.ModeViper.GetStringSlice("blacklist")
	if len(modelsCmdConfig.BlacklistTables) == 1 && strings.ContainsRune(modelsCmdConfig.BlacklistTables[0], ',') {
		modelsCmdConfig.BlacklistTables, err = cmd.Flags().GetStringSlice("blacklist")
		if err != nil {
			return err
		}
	}

	modelsCmdConfig.WhitelistTables = cnf.ModeViper.GetStringSlice("whitelist")
	if len(modelsCmdConfig.WhitelistTables) == 1 && strings.ContainsRune(modelsCmdConfig.WhitelistTables[0], ',') {
		modelsCmdConfig.WhitelistTables, err = cmd.Flags().GetStringSlice("whitelist")
		if err != nil {
			return err
		}
	}

	modelsCmdConfig.Tags = cnf.ModeViper.GetStringSlice("tag")
	if len(modelsCmdConfig.Tags) == 1 && strings.ContainsRune(modelsCmdConfig.Tags[0], ',') {
		modelsCmdConfig.Tags, err = cmd.Flags().GetStringSlice("tag")
		if err != nil {
			return err
		}
	}

	modelsCmdConfig.Replacements = cnf.ModeViper.GetStringSlice("replace")
	if len(modelsCmdConfig.Replacements) == 1 && strings.ContainsRune(modelsCmdConfig.Replacements[0], ',') {
		modelsCmdConfig.Replacements, err = cmd.Flags().GetStringSlice("replace")
		if err != nil {
			return err
		}
	}

	if modelsCmdConfig.DriverName == "postgres" {
		modelsCmdConfig.Postgres = boilingcore.PostgresConfig{
			User:    cnf.ModeViper.GetString("user"),
			Pass:    cnf.ModeViper.GetString("pass"),
			Host:    cnf.ModeViper.GetString("host"),
			Port:    cnf.ModeViper.GetInt("port"),
			DBName:  cnf.ModeViper.GetString("dbname"),
			SSLMode: cnf.ModeViper.GetString("sslmode"),
		}

		if modelsCmdConfig.Postgres.SSLMode == "" {
			modelsCmdConfig.Postgres.SSLMode = "require"
			cnf.ModeViper.Set("sslmode", modelsCmdConfig.Postgres.SSLMode)
		}

		if modelsCmdConfig.Postgres.Port == 0 {
			modelsCmdConfig.Postgres.Port = 5432
			cnf.ModeViper.Set("port", modelsCmdConfig.Postgres.Port)
		}

		err = vala.BeginValidation().Validate(
			vala.StringNotEmpty(modelsCmdConfig.Postgres.User, "user"),
			vala.StringNotEmpty(modelsCmdConfig.Postgres.Host, "host"),
			vala.Not(vala.Equals(modelsCmdConfig.Postgres.Port, 0, "port")),
			vala.StringNotEmpty(modelsCmdConfig.Postgres.DBName, "dbname"),
			vala.StringNotEmpty(modelsCmdConfig.Postgres.SSLMode, "sslmode"),
		).Check()

		if err != nil {
			return err
		}
	}

	if modelsCmdConfig.DriverName == "mysql" {
		modelsCmdConfig.MySQL = boilingcore.MySQLConfig{
			User:    cnf.ModeViper.GetString("user"),
			Pass:    cnf.ModeViper.GetString("pass"),
			Host:    cnf.ModeViper.GetString("host"),
			Port:    cnf.ModeViper.GetInt("port"),
			DBName:  cnf.ModeViper.GetString("dbname"),
			SSLMode: cnf.ModeViper.GetString("sslmode"),
		}

		// Set MySQL TinyintAsBool global var. This flag only applies to MySQL.
		// Invert the value since ABCWeb takes it as "not" bool instead of "as" bool.
		drivers.TinyintAsBool = !cnf.ModeViper.GetBool("tinyint-not-bool")

		// MySQL doesn't have schemas, just databases
		modelsCmdConfig.Schema = modelsCmdConfig.MySQL.DBName

		if modelsCmdConfig.MySQL.SSLMode == "" {
			modelsCmdConfig.MySQL.SSLMode = "true"
			cnf.ModeViper.Set("sslmode", modelsCmdConfig.MySQL.SSLMode)
		}

		if modelsCmdConfig.MySQL.Port == 0 {
			modelsCmdConfig.MySQL.Port = 3306
			cnf.ModeViper.Set("port", modelsCmdConfig.MySQL.Port)
		}

		err = vala.BeginValidation().Validate(
			vala.StringNotEmpty(modelsCmdConfig.MySQL.User, "user"),
			vala.StringNotEmpty(modelsCmdConfig.MySQL.Host, "host"),
			vala.Not(vala.Equals(modelsCmdConfig.MySQL.Port, 0, "port")),
			vala.StringNotEmpty(modelsCmdConfig.MySQL.DBName, "dbname"),
			vala.StringNotEmpty(modelsCmdConfig.MySQL.SSLMode, "sslmode"),
		).Check()

		if err != nil {
			return err
		}
	}

	modelsCmdState, err = boilingcore.New(modelsCmdConfig)
	if err != nil {
		return err
	}

	// fix imports
	modelsCmdState.Importer.TestMain.Add("postgres", `"github.com/volatiletech/abcweb/abcconfig"`, true)
	modelsCmdState.Importer.TestMain.Add("postgres", `"github.com/volatiletech/helpers/git"`, true)
	modelsCmdState.Importer.TestMain.Add("postgres", `"github.com/volatiletech/abcweb/abcconfig"`, true)
	modelsCmdState.Importer.TestMain.Add("postgres", `"path/filepath"`, true)
	modelsCmdState.Importer.TestMain.Add("mysql", `"github.com/volatiletech/abcweb/abcconfig"`, true)
	modelsCmdState.Importer.TestMain.Add("mysql", `"github.com/volatiletech/helpers/git"`, true)
	modelsCmdState.Importer.TestMain.Add("mysql", `"github.com/volatiletech/abcweb/abcconfig"`, true)
	modelsCmdState.Importer.TestMain.Add("mysql", `"path/filepath"`, true)
	modelsCmdState.Importer.TestSingleton.Remove("boil_main_test", `"path/filepath"`)
	modelsCmdState.Importer.TestSingleton.Remove("boil_main_test", `"github.com/kat-co/vala"`)
	modelsCmdState.Importer.TestSingleton.Remove("boil_main_test", `"github.com/pkg/errors"`)
	modelsCmdState.Importer.TestSingleton.Remove("boil_main_test", `"github.com/spf13/viper"`)
	modelsCmdState.Importer.TestMain.Remove("postgres", `"github.com/spf13/viper"`)
	modelsCmdState.Importer.TestMain.Remove("mysql", `"github.com/spf13/viper"`)

	return err
}

func modelsCmdRun(cmd *cobra.Command, args []string) error {
	err := modelsCmdState.Run(true)
	if err != nil {
		return err
	}

	fmt.Println("Success   Generating models")
	return nil
}

func migrationCmdRun(cmd *cobra.Command, args []string) error {
	checkDep("mig")

	if len(args) == 0 || len(args[0]) == 0 {
		fmt.Println(`command requires a migration name argument`)
		os.Exit(1)
	}

	exc := exec.Command("mig", "create", args[0])
	exc.Dir = filepath.Join(cnf.AppPath, migrationsDirectory)

	out, err := exc.CombinedOutput()

	fmt.Print(string(out))

	return err
}

// configCmdPreRun sets up the flag bindings
func configCmdPreRun(*cobra.Command, []string) {
	cnf.ModeViper.BindPFlags(configCmd.Flags())
}

func configCmdRun(cmd *cobra.Command, args []string) error {
	fmt.Println("Generating fresh config files...")
	cfg := &newConfig{}
	_, err := toml.DecodeFile(filepath.Join(cnf.AppPath, ".abcweb.toml"), cfg)
	if err == os.ErrNotExist {
		fmt.Println("warning: unable to find .abcweb.toml, so your config may need tweaking")
		cfg.DefaultEnv = "prod"
	} else if err != nil {
		return err
	}

	cfg.AppName = cnf.AppName
	cfg.AppEnvName = cnf.AppEnvName
	cfg.AppPath = cnf.AppPath

	err = genConfigFiles(cnf.AppPath, cfg, false, cnf.ModeViper.GetBool("force"))
	if err != nil {
		return err
	}

	fmt.Println("SUCCESS.")
	return nil
}

// certsCmdPreRun sets up the flag bindings
func certsCmdPreRun(*cobra.Command, []string) {
	cnf.ModeViper.BindPFlags(certsCmd.Flags())
}

func certsCmdRun(cmd *cobra.Command, args []string) error {
	fmt.Println("Generating TLS certificates...")
	cfg := newConfig{}
	_, err := toml.DecodeFile(filepath.Join(cnf.AppPath, ".abcweb.toml"), &cfg)
	if err != nil {
		fmt.Println("warning: unable to find .abcweb.toml, so your cert configuration may be invalid")
		cfg.DefaultEnv = "prod"
		cfg.TLSCommonName = "localhost"
	}

	cfg.AppName = cnf.AppName
	cfg.AppEnvName = cnf.AppEnvName
	cfg.AppPath = cnf.AppPath
	cfg.Silent = true

	if cnf.ModeViper.GetBool("force") {
		os.Remove(filepath.Join(cnf.AppPath, "cert.pem"))
		os.Remove(filepath.Join(cnf.AppPath, "private.key"))
	}

	if err := generateTLSCerts(cfg); err != nil {
		return err
	}

	fmt.Println("SUCCESS.")
	return nil
}

// genConfigFiles will generate fresh config files into dstFolder.
// If skipNonProd is set to true it will skip config files that are not
// required in production (such as watch.toml)
func genConfigFiles(dstFolder string, cfg *newConfig, skipNonProd bool, force bool) error {
	// Get base path containing templates folder and source files
	p, _ := build.Default.Import(basePackage, "", build.FindOnly)
	if p == nil || len(p.Dir) == 0 {
		return errors.New("cannot locate base path containing templates folder")
	}

	configFiles := map[string]string{
		filepath.Join(p.Dir, "templates", "config.toml.tmpl"): "config.toml",
	}

	if !skipNonProd {
		configFiles[filepath.Join(p.Dir, "templates", "watch.toml.tmpl")] = "watch.toml"
	}

	for src, fname := range configFiles {
		dst := filepath.Join(dstFolder, fname)

		var perm os.FileMode
		f, err := os.Stat(dst)
		if err == nil {
			// if force set and file exists delete file for recreation
			if force {
				perm = f.Mode()
				if err := os.Remove(dst); err != nil {
					return err
				}
			} else { // if force not set and file exists then continue to next file
				continue
			}
		} else { // if file doesnt exist default to 0644 perms
			perm = 0644
		}

		srcContents, err := ioutil.ReadFile(src)
		if err != nil {
			return err
		}

		t, err := template.New("").Funcs(templateFuncs).Parse(string(srcContents))
		if err != nil {
			return err
		}

		fileContents := &bytes.Buffer{}
		err = t.Execute(fileContents, cfg)

		if err := ioutil.WriteFile(dst, fileContents.Bytes(), perm); err != nil {
			return err
		}
	}

	return nil
}
