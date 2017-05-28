package abcserver

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/zap"

	"github.com/unrolled/render"
	"github.com/volatiletech/abcweb/abcconfig"
	"github.com/volatiletech/abcweb/abcmiddleware"
	"github.com/volatiletech/abcweb/abcrender"
)

// setup the temp dir public folder to give the NotFound handler files
// to work with
func testSetupAssets() (string, error) {
	// this is the testing public dir
	dir, err := ioutil.TempDir("", "handlertestspublic")
	if err != nil {
		return dir, err
	}

	// make robots.txt non-compiled asset file
	err = ioutil.WriteFile(filepath.Join(dir, "robots.txt"), []byte{}, 0755)
	if err != nil {
		return dir, err
	}

	err = os.MkdirAll(filepath.Join(dir, "assets", "css"), 0755)
	if err != nil {
		return dir, err
	}

	// make main.css compiled asset file
	err = ioutil.WriteFile(filepath.Join(dir, filepath.FromSlash("assets/css"), "main.css"), []byte{}, 0755)
	if err != nil {
		return dir, err
	}

	return dir, nil
}

// setup the temp dir templates folder to give the NotFound handler files
// to work with
func testSetupTemplates() (string, error) {
	// this is the testing public dir
	dir, err := ioutil.TempDir("", "handlerteststemplates")
	if err != nil {
		return dir, err
	}

	err = os.MkdirAll(filepath.Join(dir, "errors"), 0755)
	if err != nil {
		return dir, err
	}

	err = ioutil.WriteFile(filepath.Join(dir, "errors", "404.tmpl"), []byte{}, 0755)
	if err != nil {
		return dir, err
	}

	err = ioutil.WriteFile(filepath.Join(dir, "errors", "405.tmpl"), []byte{}, 0755)
	if err != nil {
		return dir, err
	}

	err = ioutil.WriteFile(filepath.Join(dir, "errors", "500.tmpl"), []byte{}, 0755)
	if err != nil {
		return dir, err
	}

	return dir, nil
}

func TestNotFound(t *testing.T) {
	t.Parallel()

	assetsDir, err := testSetupAssets()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(assetsDir)

	templatesDir, err := testSetupTemplates()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(templatesDir)

	// create config
	serverCfg := abcconfig.ServerConfig{
		PublicPath: assetsDir,
	}

	// creater render
	render := &abcrender.Render{
		Render: render.New(render.Options{
			Directory:                 templatesDir,
			Extensions:                []string{".tmpl"},
			IsDevelopment:             true,
			DisableHTTPErrorRendering: true,
		}),
	}

	// create logger
	log, err := zap.NewDevelopment()
	if err != nil {
		t.Error(err)
	}

	// test the non-compiled assets hotpath first
	r := httptest.NewRequest("GET", "/robots.txt", nil)
	w := httptest.NewRecorder()

	n := NewNotFoundHandler(nil)
	notFound := n.Handler(serverCfg, render)

	// set the logger on the context so calls to abcmiddleware.Log don't fail
	r = r.WithContext(context.WithValue(r.Context(), abcmiddleware.CtxLoggerKey, log))

	// Call the handler
	notFound(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected http 200, but got http %d", w.Code)
	}
	loc, err := w.Result().Location()
	if err == nil {
		t.Error("did not expect a redirect, but got one to:", loc.String())
	}

	// test the compiled assets hotpath with non-manifest
	r = httptest.NewRequest("GET", "/assets/css/main.css", nil)
	w = httptest.NewRecorder()

	// set the logger on the context so calls to abcmiddleware.Log don't fail
	r = r.WithContext(context.WithValue(r.Context(), abcmiddleware.CtxLoggerKey, log))

	// Call the handler
	notFound(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected http 200, but got http %d", w.Code)
	}
	loc, err = w.Result().Location()
	if err == nil {
		t.Error("did not expect a redirect, but got one to:", loc.String())
	}

	// test the compiled assets hotpath with manifest
	r = httptest.NewRequest("GET", "/assets/css/main-manifestmagic.css", nil)
	w = httptest.NewRecorder()

	r = r.WithContext(context.WithValue(r.Context(), abcmiddleware.CtxLoggerKey, log))

	// Set asset manifest to test manifest hotpath
	manifest := map[string]string{
		"css/main-manifestmagic.css": "css/main.css",
	}
	serverCfg.AssetsManifest = true

	n = NewNotFoundHandler(manifest)
	notFound = n.Handler(serverCfg, render)

	// Call the handler
	notFound(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected http 200, but got http %d", w.Code)
	}
	loc, err = w.Result().Location()
	if err == nil {
		t.Error("did not expect a redirect, but got one to:", loc.String())
	}
}
