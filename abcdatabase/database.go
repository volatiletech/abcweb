package abcdatabase

import (
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/bdb/drivers"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/abcweb/abcconfig"
	"github.com/volatiletech/helpers/git"
	"github.com/volatiletech/mig"
)

var (
	rgxMigrationVersion = regexp.MustCompile(`([0-9]+).*\.sql`)
	// ErrNoMigrations occurs if no migration files can be found on disk
	ErrNoMigrations = errors.New("no migrations found")
)

// TestdataFunc is the function signature for the sql testdata function
// that performs database operations once the test suite is initialized
// with migrations and the sql testdata file.
type TestdataFunc func(driver, conn string) error

// GetConnStr returns a connection string for the database software used
func GetConnStr(cfg abcconfig.DBConfig) (string, error) {
	if len(cfg.DB) == 0 {
		return "", errors.New("db field in config.toml must be provided")
	}

	if cfg.DB == "postgres" {
		return drivers.PostgresBuildQueryString(cfg.User, cfg.Pass, cfg.DBName, cfg.Host, cfg.Port, cfg.SSLMode), nil
	} else if cfg.DB == "mysql" {
		return drivers.MySQLBuildQueryString(cfg.User, cfg.Pass, cfg.DBName, cfg.Host, cfg.Port, cfg.SSLMode), nil
	}

	return "", fmt.Errorf("cannot get connection string for unknown database %q", cfg.DB)
}

// SetupTestSuite fully initializes a test database for you and returns
// a database connection to that test database and the number of migrations
// executed. It executes migrations testdata.sql and a TestdataFunc against
// your test database if present.
//
// SetupTestSuite retrieves the test database configuration from the
// config.toml file located in the root of the app, and loads the "test"
// section of the config file. It also sets the environment prefix to
// the name of the app (in uppercase, i.e. "my app" -> "MY_APP").
//
// SetupTestSuite will run the migrations located at approot/db/migrations
// and then execute the testdata.sql file located at approot/db/testdata.sql
// if it exists. If a TestdataFunc is provided to SetupTestSuite it will
// execute this handler after executing the testdata.sql file.
//
// SetupTestSuite returns a connection to the test database, the number
// of migrations executed, and an error if present.
func SetupTestSuite(testdata TestdataFunc) (*sql.DB, int, error) {
	var db *sql.DB
	appPath := git.GetAppPath()

	c := &abcconfig.Config{
		File:      filepath.Join(appPath, "config.toml"),
		LoadEnv:   "test",
		EnvPrefix: git.GetAppEnvName(),
	}

	cfg := &abcconfig.AppConfig{}

	// Load the database config from test env into cfg
	_, err := c.Bind(nil, cfg)
	if err != nil {
		return nil, 0, errors.Wrap(err, "cannot load environment named test from config.toml")
	}

	// Set the sqlboiler debugmode
	boil.DebugMode = cfg.DB.DebugMode

	err = createTestDB(cfg.DB)
	if err != nil {
		return nil, 0, errors.Wrap(err, "unable to create test database")
	}

	count, err := RunMigrations(cfg.DB, filepath.Join(appPath, "db", "migrations"))
	if err != nil {
		return nil, count, errors.Wrap(err, "cannot execute migrations up")
	}

	testdataFile := filepath.Join(appPath, "db", "testdata.sql")
	_, err = os.Stat(testdataFile)
	if err != nil && !os.IsNotExist(err) {
		return nil, count, errors.Wrapf(err, "unexpected error occurred trying to load %s", testdataFile)
	} else if err != nil { // file not exists error
		// set the file to empty string to tell SetupTestdata not to execute it
		testdataFile = ""
	}

	err = SetupTestdata(cfg.DB, testdataFile, testdata)
	if err != nil {
		return nil, 0, errors.Wrap(err, "unable to setup testdata")
	}

	connStr, err := GetConnStr(cfg.DB)
	if err != nil {
		return nil, count, err
	}

	db, err = sql.Open(cfg.DB.DB, connStr)
	if err != nil {
		return nil, count, errors.Wrap(err, "cannot connect to test database")
	}

	return db, count, nil
}

// SetupDBData executes the migrations "up" against the passed in database
// and also inserts the test data defined in testdata.sql and executes
// the passed in TestdataFunc handler if present, and then returns the
// number of migrations run.
func SetupDBData(cfg abcconfig.DBConfig, testdata TestdataFunc) (int, error) {
	var err error

	appPath := git.GetAppPath()

	err = createTestDB(cfg)
	if err != nil {
		return 0, err
	}

	count, err := RunMigrations(cfg, filepath.Join(appPath, "db", "migrations"))
	if err != nil {
		return count, errors.Wrap(err, "cannot execute migrations up")
	}

	testdataFile := filepath.Join(appPath, "db", "testdata.sql")
	err = SetupTestdata(cfg, testdataFile, testdata)

	return count, err
}

// createTestDB drops the test database if it exists, then recreates it
func createTestDB(cfg abcconfig.DBConfig) error {
	// copy cfg into cfgNew
	cfgNew := cfg

	// Drop and create the database if it exists so each test run starts
	// on a clean slate
	if cfg.DB == "postgres" {
		// Set the db to the default postgres db so we have something to connect to
		// to create a new database or drop old databases
		cfgNew.DBName = "postgres"
	} else {
		// Other databases like MySQL don't have a default db to connect to
		cfgNew.DBName = ""
	}

	connStr, err := GetConnStr(cfgNew)
	if err != nil {
		return err
	}

	db, err := sql.Open(cfgNew.DB, connStr)
	if err != nil {
		return err
	}

	_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s;", cfg.DBName))
	if err != nil {
		return errors.Wrap(err, "drop if exists failed")
	}

	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s;", cfg.DBName))
	if err != nil {
		return errors.Wrap(err, "create database failed")
	}

	db.Close()
	return nil
}

// RunMigrations executes the migrations "up" against the passed in database
// and returns the number of migrations run.
func RunMigrations(cfg abcconfig.DBConfig, migrationsPath string) (int, error) {
	connStr, err := GetConnStr(cfg)
	if err != nil {
		return 0, err
	}

	count, err := mig.Up(cfg.DB, connStr, migrationsPath)
	if err != nil {
		return count, err
	}

	return count, nil
}

// SetupTestdata executes the passed in sql file against the passed in database
// and then executes the testdata handler function.
//
// SetupTestdata takes an db config, and an optional testdata.sql file path
// and optional testdata function handler.
//
// If path or handler is nil they will be skipped. If both are nil an error
// will be thrown.
func SetupTestdata(cfg abcconfig.DBConfig, testdataPath string, testdataFunc TestdataFunc) error {
	if len(testdataPath) == 0 && testdataFunc == nil {
		return errors.New("no testdata resource provided")
	}
	var err error

	connStr, err := GetConnStr(cfg)
	if err != nil {
		return err
	}

	if len(testdataPath) > 0 {
		contents, err := ioutil.ReadFile(testdataPath)
		if err != nil {
			return err
		}
		if len(contents) > 0 {
			if err := ExecuteScript(cfg, contents); err != nil {
				return err
			}
		}
	}

	// call the users testdata handler func if provided
	if testdataFunc != nil {
		err = testdataFunc(cfg.DB, connStr)
	}

	return err
}

// ExecuteScript executes the passed in SQL script against the passed in db
func ExecuteScript(cfg abcconfig.DBConfig, script []byte) error {
	rdr := bytes.NewReader(script)

	if cfg.DB == "postgres" {
		passFilePath, err := pgPassFile(cfg)
		if err != nil {
			return err
		}
		defer os.Remove(passFilePath)

		cmd := exec.Command("psql", cfg.DBName, "-v", "ON_ERROR_STOP=1")
		cmd.Stdin = rdr
		cmd.Env = append(os.Environ(), pgEnv(cfg, passFilePath)...)

		res, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf(string(res))
		}

		return err
	} else if cfg.DB == "mysql" {
		passFile, err := mysqlPassFile(cfg)
		if err != nil {
			return err
		}
		defer os.Remove(passFile)

		cmd := exec.Command("mysql", fmt.Sprintf("--defaults-file=%s", passFile), "--database", cfg.DBName)
		cmd.Stdin = rdr

		res, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println(string(res))
		}

		return err
	}

	return fmt.Errorf("cannot execute sql script, incompatible database %q", cfg.DB)
}

// IsMigrated returns true if the database is migrated to the
// latest migration in the db/migrations folder. It also
// returns the current database migration version number.
func IsMigrated(cfg abcconfig.DBConfig) (bool, int64, error) {
	files, err := ioutil.ReadDir(filepath.Join("db", "migrations"))
	if err != nil || len(files) == 0 {
		return false, 0, ErrNoMigrations
	}

	connStr, err := GetConnStr(cfg)
	if err != nil {
		return false, 0, err
	}

	version, err := mig.Version(cfg.DB, connStr)
	if err != nil {
		return false, version, err
	}

	return isLatestVersion(version, files), version, nil
}

// isLatestVersion loops over all passed in files and determines whether
// dbVersion is the latest migration file version by checking filenames
func isLatestVersion(dbVersion int64, files []os.FileInfo) bool {
	highestVersion := int64(0)
	for _, file := range files {
		a := rgxMigrationVersion.FindStringSubmatch(file.Name())
		if len(a) <= 1 {
			continue
		}
		fileVersion := a[1]

		fileVersionNum, err := strconv.ParseInt(fileVersion, 10, 64)
		if err != nil {
			continue
		}

		if fileVersionNum > highestVersion {
			highestVersion = fileVersionNum
		}
	}

	return highestVersion == dbVersion
}

// pgEnv returns a slice of the connection related environment variables
func pgEnv(cfg abcconfig.DBConfig, passFilePath string) []string {
	return []string{
		fmt.Sprintf("PGHOST=%s", cfg.Host),
		fmt.Sprintf("PGPORT=%d", cfg.Port),
		fmt.Sprintf("PGUSER=%s", cfg.User),
		fmt.Sprintf("PGPASSFILE=%s", passFilePath),
	}
}

// pgPassFile creates a file in the temp directory containing the connection
// details and password for the database to be passed into the mysql cmdline cmd
func pgPassFile(cfg abcconfig.DBConfig) (string, error) {
	tmp, err := ioutil.TempFile("", "pgpass")
	if err != nil {
		return "", errors.New("failed to create postgres pass file")
	}
	defer tmp.Close()

	fmt.Fprintf(tmp, "%s:%d:%s:%s", cfg.Host, cfg.Port, cfg.DBName, cfg.User)
	if len(cfg.Pass) != 0 {
		fmt.Fprintf(tmp, ":%s", cfg.Pass)
	}
	fmt.Fprintln(tmp)

	return tmp.Name(), nil
}

// mysqlPassFile creates a file in the temp directory containing the connection
// details and password for the database to be passed into the mysql cmdline cmd
func mysqlPassFile(cfg abcconfig.DBConfig) (string, error) {
	tmp, err := ioutil.TempFile("", "mysqlpass")
	if err != nil {
		return "", errors.Wrap(err, "failed to create mysql pass file")
	}
	defer tmp.Close()

	fmt.Fprintln(tmp, "[client]")
	fmt.Fprintf(tmp, "host=%s\n", cfg.Host)
	fmt.Fprintf(tmp, "port=%d\n", cfg.Port)
	fmt.Fprintf(tmp, "user=%s\n", cfg.User)
	fmt.Fprintf(tmp, "password=%s\n", cfg.Pass)

	var sslMode string
	switch cfg.SSLMode {
	case "true":
		sslMode = "REQUIRED"
	case "false":
		sslMode = "DISABLED"
	default:
		sslMode = "PREFERRED"
	}

	fmt.Fprintf(tmp, "ssl-mode=%s\n", sslMode)

	return tmp.Name(), nil
}
