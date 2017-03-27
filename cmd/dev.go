package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/markbates/refresh/refresh"
	"github.com/spf13/cobra"
)

// devCmd represents the "dev" command
var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Runs your ABCWeb app for development",
	Long: `Runs your ABCWeb app for development. The dev command has a watch component
that watches for any changes to Go files and re-builds/re-runs your app on change. 
It will also watch for any change to assets and execute the corresponding gulp tasks 
for the changed asset file(s). The dev command also hosts a live-reload server
and will notify any connected live-reload clients and/or browsers to reload
the page once the corresponding watcher task is finished.`,
	Example: `abcweb dev`,
	Run:     devCmdRun,
}

func init() {
	devCmd.Flags().BoolP("go-only", "g", false, "Only watch and rebuild the go app by piping to refresh app")
	devCmd.Flags().BoolP("assets-only", "a", false, "Only watch and build the assets by piping to gulp watch")

	RootCmd.AddCommand(devCmd)
}

func devCmdRun(cmd *cobra.Command, args []string) {
	cnf.ModeViper.BindPFlags(cmd.Flags())

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	if cnf.ModeViper.GetBool("go-only") {
		go func() {
			err := startRefresh(ctx)
			if err != nil {
				cancel()
				fmt.Println(err)
				os.Exit(1)
			}
		}()
	} else if cnf.ModeViper.GetBool("assets-only") {
		go func() {
			err := startGulp(ctx)
			if err != nil {
				cancel()
				fmt.Println(err)
				os.Exit(1)
			}
		}()
	} else {
		go func() {
			err := startGulp(ctx)
			if err != nil {
				cancel()
				fmt.Println(err)
				os.Exit(1)
			}
		}()
		go func() {
			err := startRefresh(ctx)
			if err != nil {
				cancel()
				fmt.Println(err)
				os.Exit(1)
			}
		}()
	}

	// Wait for ctx to finish
	<-ctx.Done()
}

// startGulp starts "gulp watch" if it can find gulpfile.js and the gulp command.
func startGulp(ctx context.Context) error {
	_, err := os.Stat(filepath.Join(cnf.AppPath, "gulpfile.js"))
	if os.IsNotExist(err) {
		fmt.Println("No gulpfile.js present, skipping gulp watch")
		return nil
	} else if err != nil {
		return err
	}

	_, err = exec.LookPath("gulp")
	if err != nil {
		fmt.Println("Cannot find gulp in PATH. Is it installed?")
		return nil
	}

	cmd := exec.Command("gulp", "watch")

	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	return cmd.Run()
}

// refresh library uses toml configuration file. This struct allows us
// to read in a toml file instead.
type refreshConfig struct {
	AppRoot            string        `toml:"app_root" yaml:"app_root"`
	IgnoredFolders     []string      `toml:"ignored_folders" yaml:"ignored_folders"`
	IncludedExtensions []string      `toml:"included_extensions" yaml:"included_extensions"`
	BuildPath          string        `toml:"build_path" yaml:"build_path"`
	BuildDelay         time.Duration `toml:"build_delay" yaml:"build_delay"`
	BinaryName         string        `toml:"binary_name" yaml:"binary_name"`
	CommandFlags       []string      `toml:"command_flags" yaml:"command_flags"`
	EnableColors       bool          `toml:"enable_colors" yaml:"enable_colors"`
	LogName            string        `toml:"log_name" yaml:"log_name"`
}

// startRefresh starts the refresh server to watch go files for recompilation.
func startRefresh(ctx context.Context) error {
	cfgFile := filepath.Join(cnf.AppPath, "watch.toml")

	_, err := os.Stat(cfgFile)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	c := &refreshConfig{
		AppRoot:            ".",
		IgnoredFolders:     []string{"vendor", "log", "logs", "tmp", "node_modules", "bin", "templates"},
		IncludedExtensions: []string{".go", ".toml"},
		BuildPath:          os.TempDir(),
		BuildDelay:         200,
		BinaryName:         "watch-build",
		CommandFlags:       []string{},
		EnableColors:       true,
	}

	// Only read the config file if it exists, otherwise use the defaults above.
	if !os.IsNotExist(err) {
		_, err = toml.DecodeFile(cfgFile, c)
		if err != nil {
			return err
		}
	}

	rc := &refresh.Configuration{
		AppRoot:            c.AppRoot,
		IgnoredFolders:     c.IgnoredFolders,
		IncludedExtensions: c.IncludedExtensions,
		BuildPath:          c.BuildPath,
		BuildDelay:         c.BuildDelay,
		BinaryName:         c.BinaryName,
		CommandFlags:       c.CommandFlags,
		EnableColors:       c.EnableColors,
		LogName:            c.LogName,
	}

	r := refresh.NewWithContext(rc, ctx)
	return r.Start()
}
