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
func GetConnStr(cfg abcconfig.DBConfig) (string, error) {
	return psqlBuildQueryString(cfg.User, cfg.Pass, cfg.DBName, cfg.Host, cfg.Port, cfg.SSLMode), nil
}

func psqlBuildQueryString(user, pass, dbname, host string, port int, sslmode string) string {
	parts := []string{}
	if len(user) != 0 {
		parts = append(parts, fmt.Sprintf("user=%s", user))
	}
	if len(pass) != 0 {
		parts = append(parts, fmt.Sprintf("password=%s", pass))
	}
	if len(dbname) != 0 {
		parts = append(parts, fmt.Sprintf("dbname=%s", dbname))
	}
	if len(host) != 0 {
		parts = append(parts, fmt.Sprintf("host=%s", host))
	}
	if port != 0 {
		parts = append(parts, fmt.Sprintf("port=%d", port))
	}
	if len(sslmode) != 0 {
		parts = append(parts, fmt.Sprintf("sslmode=%s", sslmode))
	}

	return strings.Join(parts, " ")
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