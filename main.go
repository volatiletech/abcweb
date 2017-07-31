package main

import (
	"fmt"
	"os"

	"github.com/volatiletech/abcweb/cmd"
)

const abcwebVersion = "3.0.3"

func main() {
	// Too much happens between here and cobra's argument handling, for
	// something so simple. Just do it immediately.
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println("ABCWeb v" + abcwebVersion)
		return
	}

	cmd.Execute()
}
