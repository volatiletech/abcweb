package abcrender

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/unrolled/render"
)

func TestNew(t *testing.T) {
	t.Parallel()

	o := New(render.Options{IsDevelopment: true}, map[string]string{"test": "test"})
	if o == nil {
		t.Error("did not expect nil")
	}

	if o.(*Render).assetsManifest["test"] != "test" {
		t.Error("expected test key to have value test")
	}
}

func TestGetManifest(t *testing.T) {
	t.Parallel()

	json := `{"a": "b"}`

	_, err := GetManifest("zxovasgfju")
	if err == nil {
		t.Error("expected error, got nil")
	}

	dir, err := ioutil.TempDir("", "manifesttest")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(dir)

	err = os.Mkdir(filepath.Join(dir, "assets"), 0755)
	if err != nil {
		t.Error(err)
	}

	err = ioutil.WriteFile(filepath.Join(dir, "assets", "manifest.json"), []byte(json), 0755)
	if err != nil {
		t.Error(err)
	}

	m, err := GetManifest(dir)
	if err != nil {
		t.Error(err)
	}

	if m["a"] != "b" {
		t.Errorf("expected m[a] to be b, got %#v", m)
	}
	if len(m) != 1 {
		t.Errorf("expected len of m to be 1, got %d", len(m))
	}
}

func TestAppHelpers(t *testing.T) {
	t.Parallel()

	fnmap := AppHelpers(map[string]string{"css/a": "css/a123"})
	if len(fnmap) == 0 {
		t.Error("expected funcmap to be returned, got len 0")
	}

	// test asset serving from manifest
	r := fnmap["cssPath"].(func(string) string)("a")
	if r != "/assets/css/a123" {
		t.Errorf("mismatch, got %s", r)
	}

	// test asset serving without manifest
	r = fnmap["cssPath"].(func(string) string)("b")
	if r != "/assets/css/b" {
		t.Errorf("mismatch, got %s", r)
	}
}

func TestLiveReloadHelper(t *testing.T) {
	t.Parallel()
	x := liveReloadHelper("path", "123")
	if x != "/assets/js/path?host=123" {
		t.Error("mismatch, got:", x)
	}
}

func TestCSSTag(t *testing.T) {
	t.Parallel()
	x := cssTag("path")
	if x != "<link href=\"path\" rel=\"stylesheet\">" {
		t.Error("mismatch, got:", x)
	}
}

func TestJSTag(t *testing.T) {
	t.Parallel()
	x := jsTag("path")
	if x != "<script src=\"path\"></script>" {
		t.Error("mismatch, got:", x)
	}
}

func TestJSBootstrap(t *testing.T) {
	t.Parallel()
	x := jsBootstrap()
	if len(x) == 0 {
		t.Error("expected contents back")
	}
}
