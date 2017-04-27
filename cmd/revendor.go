package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
)

var revendorCmdConfig revendorConfig

// revendorCmd represents the revendor command
var revendorCmd = &cobra.Command{
	Use:   "revendor",
	Short: "Revendor updates your vendor.json file and resyncs your vendor folder",
	Long: `Revendor updates your vendor/vendor.json file and resyncs your vendor folder.
	Revendor will only update the packages that abcweb has provided and skip the rest.
	Make sure you have run "abcweb deps -u" before running this command.`,
	Example: "abcweb revendor",
	RunE:    revendorCmdRun,
}

func init() {
	RootCmd.AddCommand(revendorCmd)
}

func revendorCmdRun(cmd *cobra.Command, args []string) error {
	cnf.ModeViper.BindPFlags(cmd.Flags())

	// Bare minimum requires git, go and govendor dependencies
	checkDep("git", "go", "govendor")

	_, err := os.Stat(filepath.Join(cnf.AppPath, "vendor", "vendor.json"))
	if err == nil {
		fmt.Println("Revendoring existing vendor.json file...")
		if err := revendor(); err != nil {
			return err
		}
	} else {
		fmt.Println("Generating new vendor.json file...")
		if err := newVendor(); err != nil {
			return err
		}
	}
	fmt.Printf("SUCCESS.\n\n")

	fmt.Println("Running govendor sync...")
	// Execute govendor sync
	err = vendorSync(newConfig{Silent: true})
	if err == nil {
		fmt.Println("SUCCESS.")
	}

	return err
}

func newVendor() error {
	cfg := newConfig{}
	_, err := toml.DecodeFile(filepath.Join(cnf.AppPath, ".abcweb.toml"), &cfg)

	if err != nil || len(cfg.ImportPath) == 0 {
		appPath := strings.Split(cnf.AppPath, string(filepath.Separator))
		if len(appPath) < 3 {
			return errors.New("unable to find .abcweb.toml file and could not guess import path")
		}

		// Last 3 elements hopefully something like "github.com/user/app"
		importPath := strings.Join(appPath[len(appPath)-3:], "/")
		fmt.Println("Warning: Unable to find .abcweb.toml, so your vendor.json import path may be wrong.")
		fmt.Printf("Guessing import path of: %s\n", importPath)
		cfg.ImportPath = importPath
	}

	cfg.AppName = cnf.AppName
	cfg.AppEnvName = cnf.AppEnvName
	cfg.AppPath = cnf.AppPath
	cfg.Silent = true

	p, _ := build.Default.Import(basePackage, "", build.FindOnly)
	if p == nil || len(p.Dir) == 0 {
		return errors.New("cannot locate base path containing templates folder")
	}

	srcContents, err := ioutil.ReadFile(filepath.Join(p.Dir, "templates", "vendor", "vendor.json.tmpl"))

	t, err := template.New("").Funcs(templateFuncs).Parse(string(srcContents))
	if err != nil {
		return err
	}

	fileContents := &bytes.Buffer{}
	err = t.Execute(fileContents, cfg)

	// Make directory if it doesn't already exist
	os.Mkdir(filepath.Join(cnf.AppPath, "vendor"), 0755)

	return ioutil.WriteFile(filepath.Join(cnf.AppPath, "vendor", "vendor.json"), fileContents.Bytes(), 0644)
}

func revendor() error {
	p, _ := build.Default.Import(basePackage, "", build.FindOnly)
	if p == nil || len(p.Dir) == 0 {
		return errors.New("cannot locate base path containing templates folder")
	}

	abcContents, err := ioutil.ReadFile(filepath.Join(p.Dir, "templates", "vendor", "vendor.json.tmpl"))
	if err != nil {
		return err
	}

	abcPkgs := map[string]interface{}{}
	err = json.Unmarshal(abcContents, &abcPkgs)
	if err != nil {
		return err
	}

	appContents, err := ioutil.ReadFile(filepath.Join(cnf.AppPath, "vendor", "vendor.json"))
	if err != nil {
		return err
	}

	appPkgs := map[string]interface{}{}
	err = json.Unmarshal(appContents, &appPkgs)
	if err != nil {
		return err
	}

	for _, appPkg := range appPkgs["package"].([]interface{}) {
		a, ok := appPkg.(map[string]interface{})
		if !ok {
			return errors.New("unable to convert package element into map string interface")
		}
		for _, abcPkg := range abcPkgs["package"].([]interface{}) {
			b, ok := abcPkg.(map[string]interface{})
			if !ok {
				return errors.New("unable to convert abcwebs vendor.json package element into map string interface")
			}
			// Find the package sections with matching github paths
			if a["path"].(string) == b["path"].(string) {
				// Update all of their relevant fields
				a["checksumSHA1"] = b["checksumSHA1"]
				a["revision"] = b["revision"]
				a["revisionTime"] = b["revisionTime"]
			}
		}
	}

	newContents, err := json.MarshalIndent(appPkgs, "", "\t")
	if err != nil {
		return err
	}

	// Overwrite the old vendor.json
	return ioutil.WriteFile(filepath.Join(cnf.AppPath, "vendor", "vendor.json"), newContents, 0644)
}
