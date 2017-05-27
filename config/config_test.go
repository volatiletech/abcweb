package config

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

func init() {
	AppFS = afero.NewMemMapFs()
	testHarnessViperReadConfig = func(modeViper *viper.Viper) error {
		return modeViper.ReadConfig(bytes.NewBuffer([]byte(fileOut)))
	}
}

const fileOut = `[dev]
[dev.db]
	db = "db1"
	host = "host1"
	port = 1
	dbname="dbname1"
	user="user1"
	pass="pass1"
	sslmode="sslmode1"
	blacklist=["blacklist1"]
	whitelist=["whitelist1"]
	tag=["tag1"]
	basedir="basedir1"
	output="output1"
	pkgname="pkgname1"
	schema="schema1"
	tinyint-not-bool=true
	no-auto-timestamps=true
	debug=true
	no-hooks=true
	no-tests=true
[prod]
[prod.db]
	db = "db2"
	host = "host2"
	port = 2
	dbname="dbname2"
	user="user2"
	pass="pass2"
	sslmode="sslmode2"
	blacklist=["blacklist2"]
	whitelist=["whitelist2"]
	tag=["tag2"]
	basedir="basedir2"
	output="output2"
	pkgname="pkgname2"
	schema="schema2"
	tinyint-not-bool=true
	no-auto-timestamps=true
	debug=true
	no-hooks=true
	no-tests=true
`

func TestNewModeViper(t *testing.T) {
	t.Parallel()

	appPath, err := getAppPath()
	if err != nil {
		t.Fatal(err)
	}
	appName := getAppName(appPath)

	modeViper := NewModeViper(appPath, appName, "prod")
	modeViper.RegisterAlias("sql", "migrations.sql")

	val := modeViper.GetString("basedir")
	if val != "basedir2" {
		t.Errorf("expected %q, got %q", "basedir2", val)
	}

	b := modeViper.GetBool("tinyint-not-bool")
	if !b {
		t.Error("expected true, got false")
	}
}

func TestGetActiveEnv(t *testing.T) {
	appPath, err := getAppPath()
	if err != nil {
		t.Fatal(err)
	}
	appName := getAppName(appPath)

	configPath := filepath.Join(appPath, "config.toml")

	// File has to be present to prevent fatal error
	afero.WriteFile(AppFS, configPath, []byte(""), 0644)
	envVal := os.Getenv("ABCWEB_ENV")
	os.Setenv("ABCWEB_ENV", "")

	env := getActiveEnv(appPath, appName)
	if env != "" {
		t.Errorf("Expected %q, got %q", "", env)
	}

	afero.WriteFile(AppFS, configPath, []byte("env=\"dog\"\n"), 0644)

	env = getActiveEnv(appPath, appName)
	if env != "dog" {
		t.Errorf("Expected %q, got %q", "dog", env)
	}

	os.Setenv("ABCWEB_ENV", "cat")

	env = getActiveEnv(appPath, appName)
	if env != "cat" {
		t.Errorf("Expected %q, got %q", "cat", env)
	}

	// Reset env var for other tests
	os.Setenv("ABCWEB_ENV", envVal)
	AppFS = afero.NewMemMapFs()
}

func TestGetAppPath(t *testing.T) {
	t.Parallel()

	path, err := getAppPath()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(path, "abcweb") {
		t.Error("Expected path to end with abcweb, but didnt. Got:", path)
	}
}

func TestGetAppName(t *testing.T) {
	t.Parallel()

	path, err := getAppPath()
	if err != nil {
		t.Fatal(err)
	}

	appName := getAppName(path)
	if appName != "abcweb" {
		t.Errorf("Expected appName %q, got %q", "abcweb", appName)
	}
}
