package abcconfig

import (
	"database/sql"
	"io/ioutil"
	"os"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/pflag"
)

func TestNewFlagSet(t *testing.T) {
	flags := NewFlagSet()

	if flags == nil {
		t.Error("expected non-nil")
	}

	val, err := flags.GetString("server.bind")
	if err != nil {
		t.Error(err)
	}
	if val != ":80" {
		t.Errorf("expected %q, got %q", ":80", val)
	}

	val, err = flags.GetString("env")
	if err != nil {
		t.Error(err)
	}
	if val != "prod" {
		t.Errorf("expected %q, got %q", "prod", val)
	}
}

// test custom struct
type CustomConfig struct {
	// The active environment section
	Env       string `toml:"env" mapstructure:"env" env:"ENV"`
	Something string `toml:"something" mapstructure:"something" env:"SOMETHING"`
	Other     string `toml:"other" mapstructure:"other" env:"OTHER"`

	CustomThing MyThing `toml:"custom-thing" mapstructure:"custom-thing"`

	Server ServerConfig `toml:"server" mapstructure:"server"`
}

// test imbedded struct
type ImbeddedConfig struct {
	AppConfig

	// The active environment section
	Something string `toml:"something" mapstructure:"something" env:"SOMETHING"`
	Other     string `toml:"other" mapstructure:"other" env:"OTHER"`

	CustomThing MyThing `toml:"custom-thing" mapstructure:"custom-thing"`
}

type MyThing struct {
	Testy  string `toml:"testy" mapstructure:"testy" env:"CUSTOM_THING_TESTY"`
	Crusty string `toml:"crusty" mapstructure:"crusty" env:"CUSTOM_THING_CRUSTY"`
	Angry  string `toml:"angry" mapstructure:"angry" env:"CUSTOM_THING_ANGRY"`
}

type RealConfig struct {
	AppConfig
}

func TestBindLoadEnv(t *testing.T) {
	// replicate real config file
	contents := []byte(`
[dev]
	[dev.server]
		bind = ":4000"
		tls-cert-file = "cert.pem"
		tls-key-file = "private.key"
		live-reload = true
		prod-logger = false
		assets-manifest = false
		assets-no-cache = true
		render-recompile = true
		sessions-dev-storer = true
	[dev.db]
		db = "postgres"
		user = "username"
		pass = "password"
		dbname = "lolwtf_dev"
		host = "localhost"
		sslmode = "require"
		enforce-migration = false
		blacklist = ["mig_migrations"]
[prod]
	[prod.server]
		bind = ":80"
		tls-bind = ":443"
		tls-cert-file = "cert.pem"
		tls-key-file = "private.key"
	[prod.db]
		pass = "password"
		dbname = "lolwtf_prod"
		host = "localhost"
		sslmode = "require"
		blacklist = ["mig_migrations"]
[test]
	[test.db]
		db = "postgres"
		user = "cooluser"
		pass = "coolpass"
		dbname = "lolwtf_test"
		host = "localhost"
		sslmode = "require"
`)

	file, err := ioutil.TempFile("", "abcconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		n := file.Name()
		file.Close()
		os.Remove(n)
	}()

	if _, err := file.Write(contents); err != nil {
		t.Fatal(err)
	}

	c := &Config{
		File:    file.Name(),
		LoadEnv: "test",
	}

	cfg := &RealConfig{}
	flags := NewFlagSet()

	if _, err := c.Bind(flags, cfg); err != nil {
		t.Error(err)
	}

	if cfg.DB.DB != "postgres" {
		t.Errorf("expected postgres, got %s", cfg.DB.DB)
	}
	if cfg.DB.DBName != "lolwtf_test" {
		t.Errorf("expected lolwtf_test, got %s", cfg.DB.DBName)
	}
	if cfg.DB.User != "cooluser" {
		t.Errorf("expected cooluser, got %s", cfg.DB.User)
	}
	if cfg.DB.Pass != "coolpass" {
		t.Errorf("expected coolpass, got %s", cfg.DB.Pass)
	}
	if cfg.DB.Host != "localhost" {
		t.Errorf("expected localhost, got %s", cfg.DB.Host)
	}
	if cfg.DB.SSLMode != "require" {
		t.Errorf("expected require, got %s", cfg.DB.SSLMode)
	}
	if cfg.DB.Port != 5432 {
		t.Errorf("expected 5432, got %d", cfg.DB.Port)
	}
}

func TestBindPublicPathEnv(t *testing.T) {
	// replicate real config file
	contents := []byte(`
[dev]
	[dev.server]
		bind = ":4000"
		tls-cert-file = "cert.pem"
		tls-key-file = "private.key"
		live-reload = true
		prod-logger = false
		assets-manifest = false
		assets-no-cache = true
		render-recompile = true
		sessions-dev-storer = true
	[dev.db]
		db = "postgres"
		user = "username"
		pass = "password"
		dbname = "lolwtf_dev"
		host = "localhost"
		sslmode = "require"
		enforce-migration = false
		blacklist = ["mig_migrations"]
[prod]
	[prod.server]
		bind = ":80"
		tls-bind = ":443"
		tls-cert-file = "cert.pem"
		tls-key-file = "private.key"
	[prod.db]
		pass = "password"
		dbname = "lolwtf_prod"
		host = "localhost"
		sslmode = "require"
		blacklist = ["mig_migrations"]
[test]
	[test.db]
		db = "postgres"
		user = "cooluser"
		pass = "coolpass"
		dbname = "lolwtf_test"
		host = "localhost"
		sslmode = "require"
`)

	var err error
	err = os.Setenv("ABCWEB_ENV", "dev")
	if err != nil {
		t.Error(err)
	}
	defer os.Unsetenv("ABCWEB_ENV")

	err = os.Setenv("ABCWEB_SERVER_PUBLIC_PATH", "testpath")
	if err != nil {
		t.Error(err)
	}
	defer os.Unsetenv("ABCWEB_SERVER_PUBLIC_PATH")

	file, err := ioutil.TempFile("", "abcconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		n := file.Name()
		file.Close()
		os.Remove(n)
	}()

	if _, err := file.Write(contents); err != nil {
		t.Fatal(err)
	}

	c := NewConfig("ABCWEB")
	c.File = file.Name()

	cfg := &RealConfig{}
	flags := NewFlagSet()

	if _, err := c.Bind(flags, cfg); err != nil {
		t.Error(err)
	}

	if cfg.DB.DB != "postgres" {
		t.Errorf("expected postgres, got %s", cfg.DB.DB)
	}
	if cfg.DB.DBName != "lolwtf_dev" {
		t.Errorf("expected lolwtf_dev, got %s", cfg.DB.DBName)
	}
	if cfg.DB.User != "username" {
		t.Errorf("expected username, got %s", cfg.DB.User)
	}
	if cfg.DB.Pass != "password" {
		t.Errorf("expected password, got %s", cfg.DB.Pass)
	}
	if cfg.DB.Host != "localhost" {
		t.Errorf("expected localhost, got %s", cfg.DB.Host)
	}
	if cfg.DB.SSLMode != "require" {
		t.Errorf("expected require, got %s", cfg.DB.SSLMode)
	}
	if cfg.DB.Port != 5432 {
		t.Errorf("expected 5432, got %d", cfg.DB.Port)
	}
	if cfg.Env != "dev" {
		t.Errorf("expected dev, got %s", cfg.Env)
	}
	if cfg.Server.PublicPath != "testpath" {
		t.Errorf("expected testpath, got %s", cfg.Server.PublicPath)
	}
}

func TestBind(t *testing.T) {
	contents := []byte(`
[prod]
	[prod.server]
		live-reload = true
		tls-bind = "hahaha"
	[prod.db]
		user = "a"
		host = "b"
		dbname = "c"
[cool]
	[cool.db]
		user = "a"
		host = "b"
		dbname = "c"
[custom]
	other = "global"
	something = "aaa"
	[custom.server]
		live-reload = true
		tls-bind = "1"
	[custom.custom-thing]
		testy = "bbb"
[imbedded]
	other = "global"
	something = "aaa"
	[imbedded.server]
		live-reload = true
		tls-bind = "1"
	[imbedded.custom-thing]
		testy = "bbb"
`)

	file, err := ioutil.TempFile("", "abcconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		n := file.Name()
		file.Close()
		os.Remove(n)
	}()

	if _, err := file.Write(contents); err != nil {
		t.Fatal(err)
	}

	err = os.Setenv("ABCWEB_SERVER_TLS_CERT_FILE", "bananas")
	if err != nil {
		t.Error(err)
	}
	defer os.Unsetenv("ABCWEB_SERVER_TLS_CERT_FILE")

	err = os.Setenv("ABCWEB_DB_DB", "postgres")
	if err != nil {
		t.Error(err)
	}
	defer os.Unsetenv("ABCWEB_DB_DB")

	cfg := &AppConfig{}
	flags := NewFlagSet()
	c := NewConfig("ABCWEB")
	c.File = file.Name()

	if _, err := c.Bind(flags, cfg); err != nil {
		t.Error(err)
	}

	// default values should be set appropriately
	if cfg.Server.Bind != ":80" {
		t.Errorf("expected bind %q, got %q", ":80", cfg.Server.Bind)
	}
	if cfg.Server.LiveReload != true {
		t.Error("expected livereload true, got false")
	}
	if cfg.Server.TLSBind != "hahaha" {
		t.Errorf("expected %q, got %q", "hahaha", cfg.Server.TLSBind)
	}
	if cfg.Server.TLSCertFile != "bananas" {
		t.Errorf("expected env var to set tls cert file to %q, got %q", "bananas", cfg.Server.TLSCertFile)
	}
	if cfg.Env != "prod" {
		t.Errorf("expected env to be prod, got %s", cfg.Env)
	}

	cfg = &AppConfig{}
	flags = NewFlagSet()

	// test flag override
	if err := flags.Set("env", "cool"); err != nil {
		t.Error(err)
	}
	if err := flags.Set("server.bind", ":9000"); err != nil {
		t.Error(err)
	}

	if _, err := c.Bind(flags, cfg); err != nil {
		t.Error(err)
	}

	// default values should be set appropriately
	if cfg.Server.Bind != ":9000" {
		t.Errorf("expected %q, got %q", ":9000", cfg.Server.Bind)
	}
	if cfg.Env != "cool" {
		t.Errorf("expected env to be cool, got %s", cfg.Env)
	}
	if cfg.DB.DB != "postgres" {
		t.Errorf("expected postgres, got %s", cfg.DB.DB)
	}
	if cfg.DB.Port != 5432 {
		t.Errorf("expected port 5432, got %d", cfg.DB.Port)
	}

	cfg = &AppConfig{}
	flags = NewFlagSet()

	// test flag override
	if err := flags.Set("env", "cool"); err != nil {
		t.Error(err)
	}

	if _, err := c.Bind(flags, cfg); err != nil {
		t.Error(err)
	}
	if cfg.Env != "cool" {
		t.Errorf("expected env to be cool, got %s", cfg.Env)
	}

	custom := &CustomConfig{}
	flags = NewFlagSet()

	if err := flags.Set("env", "custom"); err != nil {
		t.Error(err)
	}

	newFlags := &pflag.FlagSet{}
	newFlags.StringP("custom-thing.testy", "", "yyy", "test flag")
	newFlags.StringP("custom-thing.crusty", "", "xxx", "test flag")
	flags.AddFlagSet(newFlags)

	err = os.Setenv("ABCWEB_SERVER_TLS_CERT_FILE", "z")
	if err != nil {
		t.Error(err)
	}
	defer os.Unsetenv("ABCWEB_SERVER_TLS_CERT_FILE")

	err = os.Setenv("ABCWEB_CUSTOM_THING_ANGRY", "zzz")
	if err != nil {
		t.Error(err)
	}
	defer os.Unsetenv("ABCWEB_CUSTOM_THING_ANGRY")

	if _, err := c.Bind(flags, custom); err != nil {
		t.Error(err)
	}

	// default values should be set appropriately
	if custom.Server.Bind != ":80" {
		t.Errorf("expected bind %q, got %q", ":80", custom.Server.Bind)
	}
	if custom.Server.LiveReload != true {
		t.Error("expected livereload true, got false")
	}
	if custom.Server.TLSBind != "1" {
		t.Errorf("expected %q, got %q", "1", custom.Server.TLSBind)
	}
	if custom.Server.TLSCertFile != "z" {
		t.Errorf("expected env var to set tls cert file to %q, got %q", "z", custom.Server.TLSCertFile)
	}
	if custom.Env != "custom" {
		t.Errorf("expected env to be custom, got %s", custom.Env)
	}
	if custom.Other != "global" {
		t.Errorf("expected other to be global, got %s", custom.Other)
	}
	if custom.Something != "aaa" {
		t.Errorf("expected something to be aaa, got %s", custom.Something)
	}
	if custom.CustomThing.Testy != "bbb" {
		t.Errorf("expected testy to be bbb, got %s", custom.CustomThing.Testy)
	}
	// test flag default value
	if custom.CustomThing.Crusty != "xxx" {
		t.Errorf("expected crusty to be xxx, got %s", custom.CustomThing.Crusty)
	}
	// test env overwrite
	if custom.CustomThing.Angry != "zzz" {
		t.Errorf("expected angry to be zzz, got %s", custom.CustomThing.Angry)
	}

	imbedded := &ImbeddedConfig{}
	flags = NewFlagSet()

	if err := flags.Set("env", "imbedded"); err != nil {
		t.Error(err)
	}

	newFlags = &pflag.FlagSet{}
	newFlags.StringP("custom-thing.testy", "", "yyy", "test flag")
	newFlags.StringP("custom-thing.crusty", "", "xxx", "test flag")
	flags.AddFlagSet(newFlags)

	err = os.Setenv("ABCWEB_SERVER_TLS_CERT_FILE", "z")
	if err != nil {
		t.Error(err)
	}
	defer os.Unsetenv("ABCWEB_SERVER_TLS_CERT_FILE")

	err = os.Setenv("ABCWEB_CUSTOM_THING_ANGRY", "zzz")
	if err != nil {
		t.Error(err)
	}
	defer os.Unsetenv("ABCWEB_CUSTOM_THING_ANGRY")

	if _, err := c.Bind(flags, imbedded); err != nil {
		t.Error(err)
	}

	// default values should be set appropriately
	if imbedded.Server.Bind != ":80" {
		t.Errorf("expected bind %q, got %q", ":80", imbedded.Server.Bind)
	}
	if imbedded.Server.LiveReload != true {
		t.Error("expected livereload true, got false")
	}
	if imbedded.Server.TLSBind != "1" {
		t.Errorf("expected %q, got %q", "1", imbedded.Server.TLSBind)
	}
	if imbedded.Server.TLSCertFile != "z" {
		t.Errorf("expected env var to set tls cert file to %q, got %q", "z", imbedded.Server.TLSCertFile)
	}
	if imbedded.Env != "imbedded" {
		t.Errorf("expected env to be imbedded, got %s", imbedded.Env)
	}
	if imbedded.Other != "global" {
		t.Errorf("expected other to be global, got %s", imbedded.Other)
	}
	if imbedded.Something != "aaa" {
		t.Errorf("expected something to be aaa, got %s", imbedded.Something)
	}
	if imbedded.CustomThing.Testy != "bbb" {
		t.Errorf("expected testy to be bbb, got %s", imbedded.CustomThing.Testy)
	}
	// test flag default value
	if imbedded.CustomThing.Crusty != "xxx" {
		t.Errorf("expected crusty to be xxx, got %s", imbedded.CustomThing.Crusty)
	}
	// test env overwrite
	if imbedded.CustomThing.Angry != "zzz" {
		t.Errorf("expected angry to be zzz, got %s", imbedded.CustomThing.Angry)
	}
}

type Test struct {
	A string `mapstructure:"a" env:"A"`
	B TestB  `mapstructure:"b"`
}

type TestB struct {
	BA string `mapstructure:"ba" env:"B_BA"`
	BB string `mapstructure:"bb" env:"B_BB"`
	C  TestC  `mapstructure:"c"`
	// unexported field test
	zz string
	// unexported pointer test
	xx *string
	// exported non-struct pointer test
	YY *string `mapstructure:"yy" env:"B_YY"`

	// test pointers to structs
	D *TestD `mapstructure:"d"`
}

type TestC struct {
	CA string `mapstructure:"ca" env:"B_C_CA"`
	CB string `mapstructure:"cb" env:"B_C_CB"`
}

type TestD struct {
	DA string `mapstructure:"da" env:"B_D_DA"`
}

// Test non-exported struct with some non-exported fields and a non-exported ptr.
// Also test how it handles only reading into DB and not Server as well.
type otherConfig struct {
	dbConn *sql.DB

	Env string   `toml:"env" mapstructure:"env" env:"ENV"`
	DB  DBConfig `toml:"db" mapstructure:"db"`

	pgPassFile string
	testDBName string
}

func TestGetTagMappings(t *testing.T) {
	t.Parallel()

	i := ""
	j := ""
	cfg := &Test{
		B: TestB{
			xx: &i,
			YY: &j,
			C:  TestC{},
			D:  &TestD{},
		},
	}

	mappings, err := GetTagMappings(cfg)
	if err != nil {
		t.Fatal(err)
	}

	expected := Mappings{
		{chain: "a", env: "A"},
		{chain: "b.ba", env: "B_BA"},
		{chain: "b.bb", env: "B_BB"},
		{chain: "b.c.ca", env: "B_C_CA"},
		{chain: "b.c.cb", env: "B_C_CB"},
		{chain: "b.yy", env: "B_YY"},
		{chain: "b.d.da", env: "B_D_DA"},
	}

	if len(mappings) != len(expected) {
		t.Errorf("expected len %d, got %d", len(expected), len(mappings))
	}

	for i, m := range mappings {
		if expected[i].chain != m.chain {
			t.Errorf("expected chain: %s, got: %s", expected[i].chain, m.chain)
		}
		if expected[i].env != m.env {
			t.Errorf("expected env: %s, got: %s", expected[i].env, m.env)
		}
	}

	rcfg := &RealConfig{}

	mappings, err = GetTagMappings(rcfg)
	if err != nil {
		t.Fatal(err)
	}

	expected = Mappings{
		{chain: "env", env: "ENV"},
		{chain: "server.live-reload", env: "SERVER_LIVE_RELOAD"},
		{chain: "server.prod-logger", env: "SERVER_PROD_LOGGER"},
		{chain: "server.bind", env: "SERVER_BIND"},
		{chain: "server.tls-bind", env: "SERVER_TLS_BIND"},
		{chain: "server.tls-cert-file", env: "SERVER_TLS_CERT_FILE"},
		{chain: "server.tls-key-file", env: "SERVER_TLS_KEY_FILE"},
		{chain: "server.read-timeout", env: "SERVER_READ_TIMEOUT"},
		{chain: "server.write-timeout", env: "SERVER_WRITE_TIMEOUT"},
		{chain: "server.idle-timeout", env: "SERVER_IDLE_TIMEOUT"},
		{chain: "server.assets-manifest", env: "SERVER_ASSETS_MANIFEST"},
		{chain: "server.assets-no-cache", env: "SERVER_ASSETS_NO_CACHE"},
		{chain: "server.render-recompile", env: "SERVER_RENDER_RECOMPILE"},
		{chain: "server.sessions-dev-storer", env: "SERVER_SESSIONS_DEV_STORER"},
		{chain: "server.public-path", env: "SERVER_PUBLIC_PATH"},
		{chain: "db.debug-mode", env: "DB_DEBUG_MODE"},
		{chain: "db.db", env: "DB_DB"},
		{chain: "db.dbname", env: "DB_DBNAME"},
		{chain: "db.host", env: "DB_HOST"},
		{chain: "db.port", env: "DB_PORT"},
		{chain: "db.user", env: "DB_USER"},
		{chain: "db.pass", env: "DB_PASS"},
		{chain: "db.sslmode", env: "DB_SSLMODE"},
		{chain: "db.enforce-migration", env: "DB_ENFORCE_MIGRATION"},
	}

	if len(mappings) != len(expected) {
		t.Errorf("expected len %d, got %d", len(expected), len(mappings))
	}

	for i, m := range mappings {
		if expected[i].chain != m.chain {
			t.Errorf("expected chain: %s, got: %s", expected[i].chain, m.chain)
		}
		if expected[i].env != m.env {
			t.Errorf("expected env: %s, got: %s", expected[i].env, m.env)
		}
	}

	ocfg := &otherConfig{}

	mappings, err = GetTagMappings(ocfg)
	if err != nil {
		t.Fatal(err)
	}

	spew.Dump(mappings)
}

// test a situation like we have in abcweb's generated app with the
// custom migrate command that only uses flags to set values
func TestCustomCommandExample(t *testing.T) {
	contents := []byte(`
[prod]
	cat = "rawr"
	[prod.server]
		bind = ":80"
		live-reload = true
		tls-bind = "hahaha"
	[prod.db]
		user = "a"
		host = "b"
		dbname = "c"
`)

	file, err := ioutil.TempFile("", "abcconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		n := file.Name()
		file.Close()
		os.Remove(n)
	}()

	if _, err := file.Write(contents); err != nil {
		t.Fatal(err)
	}

	err = os.Setenv("ABCWEB_DB_DB", "postgres")
	if err != nil {
		t.Error(err)
	}
	defer os.Unsetenv("ABCWEB_DB_DB")

	cfg := &AppConfig{}
	c := NewConfig("ABCWEB")
	c.File = file.Name()

	flags := &pflag.FlagSet{}
	flags.BoolP("down", "d", false, "Roll back the database migration version by one")
	flags.StringP("dog", "", "woof", "Testy test")
	flags.StringP("cat", "", "meow", "Testy test")
	flags.StringP("env", "e", "prod", "The database config file environment to load")

	// Not compulsory, but allows users to pass in db settings to command
	flags.AddFlagSet(NewDBFlagSet())

	v, err := c.Bind(flags, cfg)
	if err != nil {
		t.Error(err)
	}

	if v.Get("dog") != "woof" {
		t.Errorf("expected dog %q, got %q", "woof", v.Get("dog"))
	}
	if v.Get("cat") != "rawr" {
		t.Errorf("expected cat %q, got %q", "rawr", v.Get("cat"))
	}

	if cfg.Server.Bind != ":80" {
		t.Errorf("expected bind %q, got %q", ":80", cfg.Server.Bind)
	}
	if cfg.Env != "prod" {
		t.Errorf("expected env to be prod, got %s", cfg.Env)
	}
	if cfg.DB.User != "a" {
		t.Errorf("expected db.user a, got %s", cfg.DB.User)
	}
	if cfg.DB.Host != "b" {
		t.Errorf("expected db.host b, got %s", cfg.DB.Host)
	}
	if cfg.DB.DBName != "c" {
		t.Errorf("expected db.dbname c, got %s", cfg.DB.DBName)
	}
	if cfg.DB.Port != 5432 {
		t.Errorf("expected db.port 5432, got %d", cfg.DB.Port)
	}
}
