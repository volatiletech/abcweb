package cmd

import (
	"os"
	"testing"
)

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
