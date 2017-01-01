package cmd

import "testing"

func TestGetAppPath(t *testing.T) {
	t.Parallel()

	appPath, appName, err := getAppPath([]string{"."})
	if err == nil {
		t.Errorf("expected error, but got none: %s - %s", appPath, appName)
	}

	appPath, appName, err = getAppPath([]string{"/"})
	if err == nil {
		t.Errorf("expected error, but got none: %s - %s", appPath, appName)
	}

	appPath, appName, err = getAppPath([]string{"/test"})
	if err != nil {
		t.Error(err)
	}
	if appPath != "/test" {
		t.Errorf("mismatch, got %s", appPath)
	}
	if appName != "test" {
		t.Errorf("mismatch, got %s", appName)
	}

	appPath, appName, err = getAppPath([]string{"./stuff/test"})
	if err != nil {
		t.Error(err)
	}
	if appPath != "stuff/test" {
		t.Errorf("mismatch, got %s", appPath)
	}
	if appName != "test" {
		t.Errorf("mismatch, got %s", appName)
	}

	appPath, appName, err = getAppPath([]string{"~/test/thing/"})
	if err != nil {
		t.Error(err)
	}
	if appPath != "~/test/thing" {
		t.Errorf("mismatch, got %s", appPath)
	}
	if appName != "thing" {
		t.Errorf("mismatch, got %s", appName)
	}
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
