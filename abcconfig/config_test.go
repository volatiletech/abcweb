package abcconfig

import (
	"io/ioutil"
	"os"
	"testing"
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

func TestInitAppConfig(t *testing.T) {
	contents := []byte(`[prod]
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
`)

	file, err := ioutil.TempFile("", "abcconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	if _, err := file.Write(contents); err != nil {
		t.Fatal(err)
	}

	// overwrite filename over in config.go
	filename = file.Name()
	SetEnvAppName("ABCWEB")

	err = os.Setenv("ABCWEB_SERVER_TLS_CERT_FILE", "bananas")
	if err != nil {
		t.Error(err)
	}
	err = os.Setenv("ABCWEB_DB_DB", "postgres")
	if err != nil {
		t.Error(err)
	}

	cfg := NewAppConfig()
	flags := NewFlagSet()
	if err := InitAppConfig(flags, cfg); err != nil {
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

	cfg = NewAppConfig()
	flags = NewFlagSet()

	// test flag override
	if err := flags.Set("env", "cool"); err != nil {
		t.Error(err)
	}
	if err := flags.Set("server.bind", ":9000"); err != nil {
		t.Error(err)
	}

	if err := InitAppConfig(flags, cfg); err != nil {
		t.Error(err)
	}

	// default values should be set appropriately
	if cfg.Server.Bind != ":9000" {
		t.Errorf("expected %q, got %q", ":9000", cfg.Server.Bind)
	}

	// test with db section
	contents = []byte(`[prod]
	[prod.server]
		live-reload = true
		tls-bind = "hahaha"
[cool]
	[cool.db]
		db = "postgres"
`)

	cfg = NewAppConfig()
	flags = NewFlagSet()

	// test flag override
	if err := flags.Set("env", "cool"); err != nil {
		t.Error(err)
	}

	if err := InitAppConfig(flags, cfg); err != nil {
		t.Error(err)
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

func TestGetTagMappings(t *testing.T) {
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
}
