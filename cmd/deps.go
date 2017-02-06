package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// depsCmd represents the deps command
var depsCmd = &cobra.Command{
	Use:   "deps",
	Short: "Download and optionally update all abcweb dependencies",
	Long: `Download and optionally update all abcweb dependencies used by
your generated app or the abcweb tool by executing "go get" commands`,
	Example: "abcweb deps -u",
	RunE:    depsCmdRun,
}

func init() {
	depsCmd.Flags().BoolP("update", "u", false, "Also update already installed dependencies")

	RootCmd.AddCommand(depsCmd)
	viper.BindPFlags(depsCmd.Flags())
}

func depsCmdRun(cmd *cobra.Command, args []string) error {
	var err error

	goGetArgs := [][]string{
		{"-t", "github.com/vattle/sqlboiler"},
		{"github.com/pressly/goose"},
		{"github.com/satori/go.uuid"},
		{"github.com/pkg/errors"},
		{"github.com/lib/pq"},
		{"github.com/go-sql-driver/mysql"},
		{"github.com/djherbis/times"},
		{"github.com/uber-go/zap"},
		{"github.com/spf13/cobra"},
		{"github.com/spf13/viper"},
		{"github.com/pressly/chi"},
		{"github.com/kat-co/vala"},
		{"github.com/goware/cors"},
		{"github.com/unrolled/render"},
	}

	npmInstallArgs := [][]string{
		{"gulp-cli"},
	}

	// Prefix Go args with "get" and optionally "-u"
	if viper.GetBool("update") {
		for i := 0; i < len(goGetArgs); i++ {
			goGetArgs[i] = append([]string{"get", "-u"}, goGetArgs[i]...)
		}
	} else {
		for i := 0; i < len(goGetArgs); i++ {
			goGetArgs[i] = append([]string{"get"}, goGetArgs[i]...)
		}
	}

	fmt.Printf("Retrieving all Go dependencies using \"go get\":\n\n")

	for _, goGetArg := range goGetArgs {
		fmt.Printf("%s ... ", goGetArg[len(goGetArg)-1])

		exc := exec.Command("go", goGetArg...)
		err := exc.Run()

		if err != nil {
			fmt.Printf("ERROR\n\n")
			fmt.Println(err)
			os.Exit(-1)
		}

		fmt.Printf("SUCCESS\n")
	}

	// Prefix NPM args with "install --global"
	for i := 0; i < len(npmInstallArgs); i++ {
		npmInstallArgs[i] = append([]string{"install", "--global"}, npmInstallArgs[i]...)
	}

	fmt.Printf("\nRetrieving all Nodejs dependencies using \"npm install --global\":\n\n")

	for _, npmInstallArg := range npmInstallArgs {
		fmt.Printf("%s ... ", npmInstallArg[len(npmInstallArg)-1])

		var out bytes.Buffer
		exc := exec.Command("npm", npmInstallArg...)
		exc.Stdout = &out
		err := exc.Run()

		if err != nil {
			fmt.Printf("ERROR\n\n")
		} else {
			fmt.Printf("SUCCESS\n\n")
		}

		fmt.Println(out.String())

		if err != nil {
			fmt.Printf("%s\n\n", err)
			fmt.Printf(`Note: If you have not already installed nodejs and npm you must do so before proceeding.
Please follow the instructions at: https://docs.npmjs.com/getting-started/installing-node

If you are receiving permission related errors, please apply the following fix: 
https://docs.npmjs.com/getting-started/fixing-npm-permissions
`)
			os.Exit(-1)
		}
	}

	fmt.Printf("All dependencies successfully installed.\n\n")

	return err
}
