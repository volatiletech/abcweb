package cmd

import (
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
	depsCmd.Flags().BoolP("verbose", "", false, "Very noisy verbose output")

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

	prependArgs := []string{"get"}
	if viper.GetBool("update") {
		prependArgs = append(prependArgs, "-u")
	}
	if viper.GetBool("verbose") {
		prependArgs = append(prependArgs, "-v")
	}

	for i := 0; i < len(goGetArgs); i++ {
		goGetArgs[i] = append(prependArgs, goGetArgs[i]...)
	}

	fmt.Printf("Retrieving all Go dependencies using \"go get\":\n\n")

	for _, goGetArg := range goGetArgs {
		fmt.Printf("%s ... ", goGetArg[len(goGetArg)-1])

		exc := exec.Command("go", goGetArg...)
		out, err := exc.CombinedOutput()

		if err != nil {
			fmt.Printf("ERROR\n\n")
		} else {
			fmt.Printf("SUCCESS\n")
		}

		if len(out) > 0 {
			fmt.Print(string(out))
		}

		if err != nil {
			fmt.Printf("%s\n\n", err)
			os.Exit(1)
		}
	}

	prependArgs = []string{"install", "--global"}
	if viper.GetBool("verbose") {
		prependArgs = append(prependArgs, "--verbose")
	}

	for i := 0; i < len(npmInstallArgs); i++ {
		npmInstallArgs[i] = append(prependArgs, npmInstallArgs[i]...)
	}

	fmt.Printf("\nRetrieving all Nodejs dependencies using \"npm install --global\":\n\n")

	_, err = exec.LookPath("npm")
	if err != nil {
		fmt.Printf(`Error: npm could not be found in your $PATH. If you have not already installed nodejs 
and npm you must do so before proceeding. Please follow the instructions at: 
https://docs.npmjs.com/getting-started/installing-node

If you receive permission related errors, please apply the following fix: 
https://docs.npmjs.com/getting-started/fixing-npm-permissions
`)
		os.Exit(1)
	}

	for _, npmInstallArg := range npmInstallArgs {
		fmt.Printf("%s ... ", npmInstallArg[len(npmInstallArg)-1])

		exc := exec.Command("npm", npmInstallArg...)
		out, err := exc.CombinedOutput()

		if err != nil {
			fmt.Printf("ERROR\n\n")
		} else {
			fmt.Printf("SUCCESS\n\n")
		}

		if len(out) > 0 {
			fmt.Print(string(out))
		}

		if err != nil {
			fmt.Printf("%s\n\n", err)
			fmt.Printf(`Note: If you are receiving a permission related exit status or error, please apply the following fix: 
https://docs.npmjs.com/getting-started/fixing-npm-permissions
`)
			os.Exit(1)
		}
	}

	fmt.Printf("All dependencies successfully installed.\n\n")

	return err
}
