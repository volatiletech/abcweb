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

	_, err := GetConnStr(cfg)
	if err == nil {
		t.Error("expected error")
	}

	cfg.DB = "postgres"
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

func TestPGEnv(t *testing.T) {
	t.Parallel()

	cfg := abcconfig.DBConfig{
		Host: "a",
		Port: 1,
		User: "b",
	}

	slice := pgEnv(cfg, "c")

	expected := []string{
		"PGHOST=a",
		"PGPORT=1",
		"PGUSER=b",
		"PGPASSFILE=c",
	}

	if len(slice) != len(expected) {
		t.Errorf("len mismatch, got %d, expect %d", len(slice), len(expected))
	}

	for i := 0; i < len(slice); i++ {
		if slice[i] != expected[i] {
			t.Errorf("value mismatch, got %s, expect %s", slice[i], expected[i])
		}
	}
}

func TestPGPassFile(t *testing.T) {
	t.Parallel()

	cfg := abcconfig.DBConfig{
		Host:   "a",
		Port:   1,
		DBName: "b",
		User:   "c",
		Pass:   "d",
	}

	expected := "a:1:b:c:d\n"

	name, err := pgPassFile(cfg)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(name)

	contents, err := ioutil.ReadFile(name)
	if err != nil {
		t.Error(err)
	}

	if string(contents) != expected {
		t.Errorf("expected %s, got %s", expected, contents)
	}
}

func TestMySQLPassFile(t *testing.T) {
	t.Parallel()

	cfg := abcconfig.DBConfig{
		Host: "a",
		Port: 1,
		User: "b",
		Pass: "c",
	}

	expected := "[client]\nhost=a\nport=1\nuser=b\npassword=c\nssl-mode=PREFERRED\n"

	name, err := mysqlPassFile(cfg)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(name)

	contents, err := ioutil.ReadFile(name)
	if err != nil {
		t.Error(err)
	}

	if string(contents) != expected {
		t.Errorf("expected %s, got %s", expected, contents)
	}
}
