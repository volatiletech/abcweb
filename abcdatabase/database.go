package abcdatabase

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/pkg/errors"
	"github.com/vattle/sqlboiler/bdb/drivers"
	"github.com/volatiletech/abcweb/abcconfig"
	"github.com/volatiletech/helpers/git"
	"github.com/volatiletech/mig"
)

var (
	rgxMigrationVersion = regexp.MustCompile(`([0-9]+).*\.sql`)
	// ErrNoMigrations occurs if no migration files can be found on disk
	ErrNoMigrations = errors.New("no migrations found")
)

// GetConnStr returns a connection string for the database software used
func GetConnStr(cfg *abcconfig.DBConfig) (string, error) {
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

// IsMigrated returns true if the database is migrated to the
// latest migration in the db/migrations folder. It also
// returns the current database migration version number.
func IsMigrated(cfg *abcconfig.DBConfig) (bool, int64, error) {
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
func pgEnv(cfg *abcconfig.DBConfig, passFilePath string) []string {
	return []string{
		fmt.Sprintf("PGHOST=%s", cfg.Host),
		fmt.Sprintf("PGPORT=%d", cfg.Port),
		fmt.Sprintf("PGUSER=%s", cfg.User),
		fmt.Sprintf("PGPASSFILE=%s", passFilePath),
	}
}

// pgPassFile creates a file in the temp directory containing the connection
// details and password for the database to be passed into the mysql cmdline cmd
func pgPassFile(cfg *abcconfig.DBConfig) (string, error) {
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
func mysqlPassFile(cfg *abcconfig.DBConfig) (string, error) {
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

// SQLTestdata executes the testdata.sql file SQL against the passed in test db.
func SQLTestdata(cfg *abcconfig.DBConfig) error {
	appPath, err := git.GetAppPath()
	if err != nil {
		return err
	}

	fh, err := os.Open(filepath.Join(appPath, "db", "testdata.sql"))
	if err != nil {
		return fmt.Errorf("cannot open testdata.sql file: %v", err)
	}
	defer fh.Close()

	if cfg.DB == "postgres" {
		passFilePath, err := pgPassFile(cfg)
		if err != nil {
			return err
		}
		defer os.Remove(passFilePath)

		cmd := exec.Command("psql", cfg.DBName, "-v", "ON_ERROR_STOP=1")
		cmd.Stdin = fh
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
		cmd.Stdin = fh

		res, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println(string(res))
		}

		return err
	}

	return fmt.Errorf("cannot import sql testdata, incompatible database %q", cfg.DB)
}
