package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
	"github.com/volatiletech/refresh/refresh"
)

// devCmd represents the "dev" command
var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Runs your abcweb app for development",
	Long: `Runs your abcweb app for development. The dev command has a watch component
that watches for any changes to Go files and re-builds/re-runs your app on change.
It will also watch for any change to assets and templates and execute the corresponding
gulp tasks for the changed asset file(s). The dev command also hosts a live-reload server
and will notify any connected live-reload clients and/or browsers to reload
the page once the corresponding watcher task is finished.`,
	Example: `abcweb dev`,
	Run:     devCmdRun,
}

func init() {
	devCmd.Flags().BoolP("go-only", "g", false, "Only watch and rebuild the go app")
	devCmd.Flags().StringP("env", "e", "dev", "The config files development environment to load")

	RootCmd.AddCommand(devCmd)
}

func devCmdRun(cmd *cobra.Command, args []string) {
	cnf.ModeViper.BindPFlags(cmd.Flags())

	// If the env command flag is set manually or ActiveEnv
	// is set to "", then set it to the viper value.
	// "dev" is a much more sane default for abcweb dev, and Windows
	// doesn't work well with temporary environment variables so setting
	// the env with APP_ENV is a lot less feasible.
	if cnf.ActiveEnv == "" || cmd.Flag("env").Changed {
		cnf.ActiveEnv = cnf.ModeViper.GetString("env")
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	if cnf.ModeViper.GetBool("go-only") {
		go func() {
			r, err := startRefresh("", ctx)
			if err != nil {
				cancel()
				fmt.Println(err)
				os.Exit(1)
			}

			if err := r.Start(); err != nil {
				cancel()
				fmt.Println(err)
				os.Exit(1)
			}
		}()
	} else {
		publicPath := filepath.Join(os.TempDir(), cnf.AppName, "public")
		err := os.MkdirAll(publicPath, 0755)
		if err != nil {
			cancel()
			fmt.Println(fmt.Sprintf("Cannot make temp folder public directory using app name %q: %s", cnf.AppName, err))
			os.Exit(1)
		}

		publicPathEnv := fmt.Sprintf("%s_SERVER_PUBLIC_PATH=%s", cnf.AppEnvName, publicPath)

		go func() {
			err := startGulp(publicPathEnv, ctx)
			if err != nil {
				cancel()
				fmt.Println(err)
				os.Exit(1)
			}
		}()
		// Delay build so terminal output is grouped together for first run.
		time.Sleep(1500 * time.Millisecond)
		go func() {
			r, err := startRefresh(publicPathEnv, ctx)
			if err != nil {
				cancel()
				fmt.Println(err)
				os.Exit(1)
			}

			// Start "hit enter key to rebuild" go routine
			go func() {
				var input string
				for {
					fmt.Scanln(&input)
					r.Restart <- true
				}
			}()

			if err := r.Start(); err != nil {
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
func startGulp(publicPathEnv string, ctx context.Context) error {
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
	cmd.Env = append([]string{publicPathEnv}, os.Environ()...)

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
	CommandEnv         []string      `toml:"command_env" yaml:"command_env"`
	EnableColors       bool          `toml:"enable_colors" yaml:"enable_colors"`
	LogName            string        `toml:"log_name" yaml:"log_name"`
}

// startRefresh starts the refresh server to watch go files for recompilation.
func startRefresh(publicPathEnv string, ctx context.Context) (*refresh.Manager, error) {
	cfgFile := filepath.Join(cnf.AppPath, "watch.toml")

	_, err := os.Stat(cfgFile)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	c := &refreshConfig{
		AppRoot:            ".",
		IgnoredFolders:     []string{"vendor", "log", "logs", "tmp", "node_modules", "bin", "templates"},
		IncludedExtensions: []string{".go", ".toml"},
		BuildPath:          os.TempDir(),
		BuildDelay:         200,
		BinaryName:         "watch-build",
		CommandFlags:       []string{},
		CommandEnv: []string{
			fmt.Sprintf("%s_ENV=%s", cnf.AppEnvName, cnf.ActiveEnv),
		},
		EnableColors: true,
	}

	// Only read the config file if it exists, otherwise use the defaults above.
	if !os.IsNotExist(err) {
		_, err = toml.DecodeFile(cfgFile, c)
		if err != nil {
			return nil, err
		}
	}

	// Append the APPNAME_SERVER_PUBLIC_PATH environment var to refresh libs CommandEnv.
	c.CommandEnv = append(c.CommandEnv, publicPathEnv)

	rc := &refresh.Configuration{
		AppRoot:            c.AppRoot,
		IgnoredFolders:     c.IgnoredFolders,
		IncludedExtensions: c.IncludedExtensions,
		BuildPath:          c.BuildPath,
		BuildDelay:         c.BuildDelay,
		BinaryName:         c.BinaryName,
		CommandFlags:       c.CommandFlags,
		CommandEnv:         c.CommandEnv,
		EnableColors:       c.EnableColors,
		LogName:            c.LogName,
	}

	r := refresh.NewWithContext(rc, ctx)

	return r, nil
}
