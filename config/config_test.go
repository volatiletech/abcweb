package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/nullbio/shift"
	"github.com/spf13/afero"
)

func init() {
	AppFS = afero.NewMemMapFs()
	testHarnessShiftLoad = testShiftLoadOverride
}

const fileOut = `[dev]
	db = "db1"
	host = "host1"
	port = 1
	db_name="dbname1"
	user="user1"
	pass="pass1"
	ssl_mode="sslmode1"
	blacklist=["blacklist1"]
	whitelist=["whitelist1"]
	tag=["tag1"]
	base_dir="basedir1"
	output="output1"
	pkg_name="pkgname1"
	schema="schema1"
	tinyint_not_bool=true
	no_auto_timestamps=true
	debug=true
	no_hooks=true
	no_tests=true
[prod]
	db = "db2"
	host = "host2"
	port = 2
	db_name="dbname2"
	user="user2"
	pass="pass2"
	ssl_mode="sslmode2"
	blacklist=["blacklist2"]
	whitelist=["whitelist2"]
	tag=["tag2"]
	base_dir="basedir2"
	output="output2"
	pkg_name="pkgname2"
	schema="schema2"
	tinyint_not_bool=true
	no_auto_timestamps=true
	debug=true
	no_hooks=true
	no_tests=true
`

func testShiftLoadOverride(c interface{}, file, prefix, env string) error {
	contents, err := afero.ReadFile(AppFS, file)
	if err != nil {
		return err
	}

	var decoded interface{}
	_, err = toml.Decode(string(contents), &decoded)
	if err != nil {
		return err
	}

	return shift.LoadWithDecoded(c, decoded, prefix, env)
}

func TestLoadDBConfig(t *testing.T) {
	appPath := GetAppPath()
	configPath := filepath.Join(appPath, "database.toml")

	afero.WriteFile(AppFS, configPath, []byte(fileOut), 0644)
	config := LoadDBConfig(appPath, "dev")

	orig := &DBConfig{
		DB:               "db1",
		Host:             "host1",
		DBName:           "dbname1",
		Port:             1,
		User:             "user1",
		Pass:             "pass1",
		SSLMode:          "sslmode1",
		Blacklist:        []string{"blacklist1"},
		Whitelist:        []string{"whitelist1"},
		Tag:              []string{"tag1"},
		BaseDir:          "basedir1",
		Output:           "output1",
		PkgName:          "pkgname1",
		Schema:           "schema1",
		TinyintNotBool:   true,
		NoAutoTimestamps: true,
		Debug:            true,
		NoHooks:          true,
		NoTests:          true,
	}

	if !reflect.DeepEqual(config, orig) {
		t.Errorf("mismatch between structs:\n%#v\n%#v\n", orig, config)
	}

	config = LoadDBConfig(appPath, "prod")

	orig = &DBConfig{
		DB:               "db2",
		Host:             "host2",
		DBName:           "dbname2",
		Port:             2,
		User:             "user2",
		Pass:             "pass2",
		SSLMode:          "sslmode2",
		Blacklist:        []string{"blacklist2"},
		Whitelist:        []string{"whitelist2"},
		Tag:              []string{"tag2"},
		BaseDir:          "basedir2",
		Output:           "output2",
		PkgName:          "pkgname2",
		Schema:           "schema2",
		TinyintNotBool:   true,
		NoAutoTimestamps: true,
		Debug:            true,
		NoHooks:          true,
		NoTests:          true,
	}

	if !reflect.DeepEqual(config, orig) {
		t.Errorf("mismatch between structs:\n%#v\n%#v\n", orig, config)
	}
}

func TestGetActiveEnv(t *testing.T) {
	appPath := GetAppPath()
	configPath := filepath.Join(appPath, "config.toml")

	// File has to be present to prevent fatal error
	afero.WriteFile(AppFS, configPath, []byte(""), 0644)

	env := GetActiveEnv(appPath)
	if env != "" {
		t.Errorf("Expected %q, got %q", "", env)
	}

	afero.WriteFile(AppFS, configPath, []byte("default_env=\"dog\"\n"), 0644)

	env = GetActiveEnv(appPath)
	if env != "dog" {
		t.Errorf("Expected %q, got %q", "dog", env)
	}

	envVal := os.Getenv("ABCWEB_ENV")
	os.Setenv("ABCWEB_ENV", "cat")

	env = GetActiveEnv(appPath)
	if env != "cat" {
		t.Errorf("Expected %q, got %q", "cat", env)
	}

	// Reset env var for other tests
	os.Setenv("ABCWEB_ENV", envVal)
	AppFS = afero.NewMemMapFs()
}

func TestGetAppPath(t *testing.T) {
	t.Parallel()

	path := GetAppPath()
	if !strings.HasSuffix(path, "abcweb") {
		t.Error("Expected path to end with abcweb, but didnt. Got:", path)
	}
}

func TestGetAppName(t *testing.T) {
	t.Parallel()

	path := GetAppPath()

	appName := GetAppName(path)
	if appName != "abcweb" {
		t.Errorf("Expected appName %q, got %q", "abcweb", appName)
	}
}
