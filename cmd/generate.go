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
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
)

// generateCmd represents the "generate" command
var generateCmd = &cobra.Command{
	Use:     "gen",
	Short:   "Generate your migrations, certs and config files",
	Example: "abcweb gen migration add_users",
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

func init() {
	// config flags
	configCmd.Flags().BoolP("force", "f", false, "Overwrite files if they already exist")

	// certs flags
	certsCmd.Flags().BoolP("force", "f", false, "Overwrite files if they already exist")

	RootCmd.AddCommand(generateCmd)

	// hook up pre-run hooks, this avoids initialization loops
	migrationCmd.PreRun = migrationCmdPreRun
	configCmd.PreRun = configCmdPreRun
	certsCmd.PreRun = certsCmdPreRun

	// Add generate subcommands
	generateCmd.AddCommand(migrationCmd)
	generateCmd.AddCommand(configCmd)
	generateCmd.AddCommand(certsCmd)
}

// migrationCmdPreRun sets up the flag bindings
func migrationCmdPreRun(*cobra.Command, []string) {
	cnf.ModeViper.BindPFlags(migrationCmd.Flags())
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
