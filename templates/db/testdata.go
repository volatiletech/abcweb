package db

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/nullbio/helpers/git"
)

// goTestdata is a function that can be edited and used to insert testdata
// into your test database after the migrations have finished executing
// when running unit tests.
//
// This can be used as an alternative to testdata.sql if you require go logic
// or would prefer to use your generated SQLBoiler model to perform db operations.
//
// To use this function just uncomment the code and perform your db operations
// using the uncommented db handle which will be connected to the test database.
func goTestdata(driver string, conn string) error {
	// db, err := sql.Open(driver, conn)
	// if err != nil {
	//	return err
	// }
	// defer db.Close()

	// database operations here to insert testdata using above db handle
	return nil
}

// sqlTestdata executes the testdata.sql file SQL against the passed in test db.
// This is not a function to be edited.
func sqlTestdata(cfg *Config) error {
	appPath, err := git.GetAppPath()
	if err != nil {
		return err
	}

	sqlFilePath := filepath.Join(appPath, "db", "testdata.sql")
	if cfg.DB == "postgres" {
		passFile, err := pgPassFile(cfg)
		if err != nil {
			return err
		}
		defer os.Remove(passFile)

		cmd := exec.Command("psql", cfg.DBName, sqlFilePath)
		cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSFILE=%s", passFile))
		return cmd.Run()
	} else if cfg.DB == "mysql" {
		passFile, err := mysqlPassFile(cfg)
		if err != nil {
			return err
		}
		defer os.Remove(passFile)

		cmd := exec.Command("mysql", fmt.Sprintf("--defaults-file=%s", passFile), "--database", cfg.DBName)
		return cmd.Run()
	}

	return fmt.Errorf("cannot import sql testdata, incompatible database software %s", cfg.DB)
}
