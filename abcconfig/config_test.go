package abcconfig

import (
	"io/ioutil"
	"testing"
)

func TestNewFlagSet(t *testing.T) {
	flags := NewFlagSet()

	if flags == nil {
		t.Error("expected non-nil")
	}

	val, err := flags.GetString("bind")
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
[cool]`)

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

	cfg := NewAppConfig()
	flags := NewFlagSet()
	if err := InitAppConfig(flags, cfg); err != nil {
		t.Error(err)
	}

	// default values should be set appropriately
	if cfg.Server.ActiveEnv != "prod" {
		t.Errorf("expected activeenv %q, got %q", "prod", cfg.Server.ActiveEnv)
	}
	if cfg.Server.Bind != ":80" {
		t.Errorf("expected bind %q, got %q", ":80", cfg.Server.Bind)
	}
	if cfg.Server.LiveReload != true {
		t.Error("expected livereload true, got false")
	}

	//	cfg = NewAppConfig()
	//
	//	// test flag override
	//	if err := flags.Set("env", "cool"); err != nil {
	//		t.Error(err)
	//	}
	//	if err := flags.Set("bind", ":9000"); err != nil {
	//		t.Error(err)
	//	}
	//
	//	if err := InitAppConfig(flags, cfg); err != nil {
	//		t.Error(err)
	//	}
	//
	//	// default values should be set appropriately
	//	if cfg.Server.ActiveEnv != "cool" {
	//		t.Errorf("expected %q, got %q", "cool", cfg.Server.ActiveEnv)
	//	}
	//	if cfg.Server.Bind != ":9000" {
	//		t.Errorf("expected %q, got %q", ":9000", cfg.Server.Bind)
	//	}
}
