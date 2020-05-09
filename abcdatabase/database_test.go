package abcdatabase

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/volatiletech/abcweb/abcconfig"
)

func TestGetConnStr(t *testing.T) {
	t.Parallel()

	cfg := abcconfig.DBConfig{}

	cfg.User = "a"
	cfg.Pass = "b"
	cfg.DBName = "c"
	cfg.Host = "1"
	cfg.Port = 2
	cfg.SSLMode = "d"

	c, err := GetConnStr(cfg)
	if err != nil {
		t.Error(err)
	}

	if c != "user=a password=b dbname=c host=1 port=2 sslmode=d" {
		t.Error("invalid value", c)
	}
}

func TestIsMigrated(t *testing.T) {
	t.Parallel()

	_, _, err := IsMigrated(abcconfig.DBConfig{})
	if err != ErrNoMigrations {
		t.Error("expected no migrations error since db/migrations doesnt exist")
	}
}

func TestIsLatestVersion(t *testing.T) {
	t.Parallel()

	dir, err := ioutil.TempDir("", "islatesttest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	fileNames := []string{
		filepath.Join(dir, "1.sql"),
		filepath.Join(dir, "2.sql"),
		filepath.Join(dir, "3.sql"),
		filepath.Join(dir, "4.sql"),
	}

	for _, fname := range fileNames {
		err := ioutil.WriteFile(fname, []byte{}, 0755)
		if err != nil {
			t.Error(err)
		}
	}

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		t.Error(err)
	}

	if isLatestVersion(1, files) {
		t.Error("expected isLatest false, got true")
	}

	if !isLatestVersion(4, files) {
		t.Error("expected isLatest true, got false")
	}

	if isLatestVersion(5, files) {
		t.Error("expected isLatest false, got true")
	}
}
