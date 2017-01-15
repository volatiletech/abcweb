package main

import (
	"fmt"
	"os"

	"github.com/nullbio/abcweb/cmd"
	"github.com/nullbio/abcweb/config"
)

const abcwebVersion = "1.0.0"

func main() {
	// Too much happens between here and cobra's argument handling, for
	// something so simple. Just do it immediately.
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println("ABCWeb v" + abcwebVersion)
		return
	}

	// initialize ModeViper to be used in the initialization funcs below
	config.ModeViper = config.NewModeViper(config.AppPath, config.ActiveEnv)

	// initialize all the flags and commands
	cmd.RootInit()
	cmd.BuildInit()
	cmd.GenerateInit()
	cmd.MigrateInit()
	cmd.NewInit()
	cmd.RunInit()
	cmd.TestInit()

	cmd.Execute()
}
