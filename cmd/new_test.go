package cmd

import (
	"os"
	"testing"

	"github.com/spf13/afero"
)

func init() {
	fs = afero.NewMemMapFs()
}

func TestGetAppPath(t *testing.T) {
	t.Parallel()

	gopath := os.Getenv("GOPATH")
	os.Setenv("GOPATH", "testpath/test")

	appPath, importPath, appName, err := getAppPath([]string{"."})
	if err == nil {
		t.Errorf("expected error, but got none: %s - %s", appPath, appName)
	}

	appPath, importPath, appName, err = getAppPath([]string{"/"})
	if err == nil {
		t.Errorf("expected error, but got none: %s - %s", appPath, appName)
	}

	appPath, importPath, appName, err = getAppPath([]string{"/test"})
	if err != nil {
		t.Error(err)
	}
	if appPath != "testpath/test/src/test" {
		t.Errorf("mismatch, got %s", appPath)
	}
	if appName != "test" {
		t.Errorf("mismatch, got %s", appName)
	}
	if importPath != "/test" {
		t.Errorf("mismatch, got %s", importPath)
	}

	appPath, importPath, appName, err = getAppPath([]string{"./stuff/test"})
	if err != nil {
		t.Error(err)
	}
	if appPath != "testpath/test/src/stuff/test" {
		t.Errorf("mismatch, got %s", appPath)
	}
	if appName != "test" {
		t.Errorf("mismatch, got %s", appName)
	}
	if importPath != "stuff/test" {
		t.Errorf("mismatch, got %s", importPath)
	}

	os.Setenv("GOPATH", gopath)
}

func TestGetProcessedPaths(t *testing.T) {
	t.Parallel()

	config := newConfig{
		AppPath: "/test/myapp",
		AppName: "myapp",
	}

	inPath := "/lol/" + templatesDirectory + "/file.tmpl"
	cleanPath, fullPath := getProcessedPaths(inPath, "/", config)
	if cleanPath != "myapp/file" {
		t.Error("mismatch:", cleanPath)
	}
	if fullPath != "/test/myapp/file" {
		t.Error("mismatch:", fullPath)
	}

	config.AppPath = "myapp"
	config.AppName = "myapp"

	cleanPath, fullPath = getProcessedPaths(inPath, "/", config)
	if cleanPath != "myapp/file" {
		t.Error("mismatch:", cleanPath)
	}
	if fullPath != "myapp/file" {
		t.Error("mismatch:", fullPath)
	}
}

func TestProcessSkips(t *testing.T) {
	config := newConfig{
		NoReadme:      true,
		NoGitIgnore:   true,
		NoConfig:      true,
		NoFontAwesome: true,
		Bootstrap:     "none",
		NoBootstrapJS: true,
	}

	// check skip basedir
	err := fs.MkdirAll("/templates", 0755)
	if err != nil {
		t.Fatal(err)
	}
	info, err := fs.Stat("/templates")
	if err != nil {
		t.Fatal(err)
	}
	skip, _ := processSkips(config, "/templates", "/templates", info)
	if skip == false {
		t.Error("expected to skip base path")
	}

	// check skip skipDirs slice
	err = fs.MkdirAll("/templates/i18n", 0755)
	if err != nil {
		t.Fatal(err)
	}
	info, err = fs.Stat("/templates/i18n")
	if err != nil {
		t.Fatal(err)
	}
	skip, err = processSkips(config, "/templates", "/templates/i18n", info)
	if skip != true || err == nil {
		t.Error("expected to skip skipDir and receive skipdir err")
	}

	// check skip readme
	f, err := fs.Create("/templates/README.md")
	if err != nil {
		t.Fatal(err)
	}
	info, err = f.Stat()
	if err != nil {
		t.Fatal(err)
	}
	skip, _ = processSkips(config, "/templates", "/templates/README.md", info)
	if skip != true {
		t.Error("expected to skip skip readme")
	}

	// check skip gitignore
	f, err = fs.Create("/templates/.gitignore")
	if err != nil {
		t.Fatal(err)
	}
	info, err = f.Stat()
	if err != nil {
		t.Fatal(err)
	}
	skip, _ = processSkips(config, "/templates", "/templates/.gitignore", info)
	if skip != true {
		t.Error("expected to skip skip readme")
	}

	// check skip config.toml
	f, err = fs.Create("/templates/config.toml")
	if err != nil {
		t.Fatal(err)
	}
	info, err = f.Stat()
	if err != nil {
		t.Fatal(err)
	}
	skip, _ = processSkips(config, "/templates", "/templates/config.toml", info)
	if skip != true {
		t.Error("expected to skip skip config.toml")
	}

	// check skip fontawesome files
	f, err = fs.Create("/templates/font-awesome.css")
	if err != nil {
		t.Fatal(err)
	}
	info, err = f.Stat()
	if err != nil {
		t.Fatal(err)
	}
	skip, _ = processSkips(config, "/templates", "/templates/font-awesome.css", info)
	if skip != true {
		t.Error("expected to skip skip font-awesome.css")
	}

	// check skip fontawesome files
	f, err = fs.Create("/templates/bootstrap.css")
	if err != nil {
		t.Fatal(err)
	}
	info, err = f.Stat()
	if err != nil {
		t.Fatal(err)
	}
	skip, _ = processSkips(config, "/templates", "/templates/bootstrap.css", info)
	if skip != true {
		t.Error("expected to skip skip bootstrap.css")
	}

	config.Bootstrap = "flex"
	// check skip fontawesome files
	f, err = fs.Create("/templates/bootstrap.js")
	if err != nil {
		t.Fatal(err)
	}
	info, err = f.Stat()
	if err != nil {
		t.Fatal(err)
	}
	skip, _ = processSkips(config, "/templates", "/templates/bootstrap.js", info)
	if skip != true {
		t.Error("expected to skip skip bootstrap.js")
	}

	// check no-skip regular go file
	f, err = fs.Create("/templates/file.go")
	if err != nil {
		t.Fatal(err)
	}
	info, err = f.Stat()
	if err != nil {
		t.Fatal(err)
	}
	skip, _ = processSkips(config, "/templates", "/templates/file.go", info)
	if skip == true {
		t.Error("did not expect skip")
	}

}

func TestNewCmdPreRun(t *testing.T) {

}

func TestNewCmdRun(t *testing.T) {

}

func TestNewCmdWalk(t *testing.T) {

}

func TestGenerateTLSCerts(t *testing.T) {
	config := newConfig{
		AppPath:       "/out/spiders",
		AppName:       "spiders",
		TLSCommonName: "dragons",
		// attempt to create tls certs twice
		// should fail second time if this is false
		TLSCertsOnly: false,
	}

	err := generateTLSCerts(config)
	if err != nil {
		t.Fatal(err)
	}

	info, err := fs.Stat("/out/spiders/cert.pem")
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() == 0 {
		t.Error("expected non-0 size for cert file")
	}

	info, err = fs.Stat("/out/spiders/private.key")
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() == 0 {
		t.Error("expected non-0 size for private key file")
	}
}
