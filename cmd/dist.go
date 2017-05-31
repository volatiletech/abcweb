package cmd

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/pierrre/archivefile/zip"
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

	if cnf.ModeViper.GetBool("config") && cnf.ModeViper.GetBool("copy-config") {
		return errors.New("config and copy-config cannot both be set at the same time")
	}

	// Delete the dist dir and recreate it so it's fresh
	if err := os.RemoveAll(filepath.Join(cnf.AppPath, "dist")); err != nil {
		return err
	}
	if err := os.Mkdir(filepath.Join(cnf.AppPath, "dist"), 0755); err != nil {
		return err
	}

	fmt.Println("Building assets...")
	if err := buildAssets(); err != nil {
		return err
	}

	fmt.Println("Building Go app...")
	if err := buildApp(); err != nil {
		return err
	}

	fmt.Println("Copying binary to dist...")
	if err := copyBinary(); err != nil {
		return err
	}

	fmt.Println("Copying folders to dist...")
	if err := copyFolders(); err != nil {
		return err
	}

	if cnf.ModeViper.GetBool("config") {
		fmt.Println("Generating fresh config files in dist...")
		if err := freshConfig(); err != nil {
			return err
		}
	}

	if cnf.ModeViper.GetBool("copy-config") {
		fmt.Println("Copying all .toml files from app root into dist...")
		if err := copyConfig(); err != nil {
			return err
		}
	}

	if cnf.ModeViper.GetBool("zip") {
		fmt.Println("Zipping dist folder into dist.zip...")
		if err := zipDist(); err != nil {
			return err
		}
	}

	return nil
}

func zipDist() error {
	// remove zip if it already exists
	os.Remove(filepath.Join(cnf.AppPath, cnf.AppName+".zip"))

	file, err := os.Create(filepath.Join(cnf.AppPath, cnf.AppName+".zip"))
	defer file.Close()
	if err != nil {
		return err
	}

	distSrc := filepath.Join(cnf.AppPath, "dist") + string(filepath.Separator)
	zip.Archive(distSrc, file, nil)
	if err != nil {
		return err
	}

	fmt.Printf("SUCCESS.\n\n")
	return nil
}

func copyConfig() error {
	rootDir, err := os.Open(filepath.Join(cnf.AppPath))
	if err != nil {
		return err
	}

	files, err := rootDir.Readdir(0)
	if err != nil {
		return err
	}

	for _, file := range files {
		name := file.Name()
		if strings.HasSuffix(name, ".toml") &&
			name != "Gopkg.toml" && name != ".abcweb.toml" && name != "watch.toml" {
			err = copyFile(
				filepath.Join(cnf.AppPath, name),
				filepath.Join(cnf.AppPath, "dist", name),
				file.Mode(),
			)
			if err != nil {
				return err
			}
		}
	}

	fmt.Printf("SUCCESS.\n\n")
	return nil
}

func freshConfig() error {
	cfg := &newConfig{}

	_, err := toml.DecodeFile(filepath.Join(cnf.AppPath, ".abcweb.toml"), cfg)
	if err != nil {
		fmt.Println("warning, unable to find .abcweb.toml, so your config may need tweaking")
	}

	// Overwrite default env to prod
	cfg.DefaultEnv = "prod"
	cfg.AppName = cnf.AppName
	cfg.AppEnvName = cnf.AppEnvName
	cfg.AppPath = cnf.AppPath

	err = genConfigFiles(filepath.Join(cnf.AppPath, "dist"), cfg, true, true)
	if err != nil {
		return err
	}

	fmt.Printf("SUCCESS.\n\n")
	return nil
}

func copyBinary() error {
	var appName string
	if runtime.GOOS == "windows" {
		appName = cnf.AppName + ".exe"
	} else {
		appName = cnf.AppName
	}

	f, err := os.Stat(filepath.Join(cnf.AppPath, appName))
	if err != nil {
		return err
	}

	err = copyFile(
		filepath.Join(cnf.AppPath, appName),
		filepath.Join(cnf.AppPath, "dist", appName),
		f.Mode(),
	)
	if err != nil {
		return err
	}

	fmt.Printf("SUCCESS.\n\n")
	return nil
}

func copyFolders() error {
	var err error
	if !cnf.ModeViper.GetBool("no-assets") {
		// copy public folder
		err = os.MkdirAll(filepath.Join(cnf.AppPath, "dist", "public"), 0755)
		if err != nil {
			return err
		}
		err = copyFolder(
			filepath.Join(cnf.AppPath, "public"),
			filepath.Join(cnf.AppPath, "dist", "public"),
		)
		if err != nil {
			return err
		}
	}

	if !cnf.ModeViper.GetBool("no-templates") {
		// copy templates folder
		err = os.Mkdir(filepath.Join(cnf.AppPath, "dist", "templates"), 0755)
		if err != nil {
			return err
		}
		err = copyFolder(
			filepath.Join(cnf.AppPath, "templates"),
			filepath.Join(cnf.AppPath, "dist", "templates"),
		)
		if err != nil {
			return err
		}
	}

	if !cnf.ModeViper.GetBool("no-migrations") {
		// copy db/migrations folder if it exists
		f, err := os.Stat(filepath.Join(cnf.AppPath, "db", "migrations"))
		if err == nil && f.Size() > 0 {
			migDir, err := os.Open(filepath.Join(cnf.AppPath, "db", "migrations"))
			defer migDir.Close()
			if err != nil {
				return err
			}

			_, err = migDir.Readdirnames(1)
			if err != nil && err != io.EOF {
				return err
			}

			// Only copy db/migrations if folder has at least 1 file
			if err != io.EOF {
				err := os.MkdirAll(filepath.Join(cnf.AppPath, "dist", "db", "migrations"), 0755)
				if err != nil {
					return err
				}
				err = copyFolder(
					filepath.Join(cnf.AppPath, "db", "migrations"),
					filepath.Join(cnf.AppPath, "dist", "db", "migrations"),
				)
				if err != nil {
					return err
				}
			}
		}
	}

	fmt.Printf("SUCCESS.\n\n")

	return nil
}

func copyFile(src, dst string, perm os.FileMode) error {
	contents, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(dst, contents, perm)
}

func copyFolder(src, dst string) error {
	return filepath.Walk(src, filepath.WalkFunc(func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, strings.TrimPrefix(path, src))
		if info.IsDir() {
			err := os.MkdirAll(dstPath, info.Mode())
			if err != nil {
				return err
			}
		} else {
			contents, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			if err := ioutil.WriteFile(dstPath, contents, info.Mode()); err != nil {
				return err
			}
		}

		return nil
	}))
}
