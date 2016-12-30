package cmd

import "testing"

func TestGetAppPath(t *testing.T) {
	t.Parallel()

	_, _, err := getAppPath([]string{"."})
	if err == nil {
		t.Error("expected error, but got none")
	}

	_, _, err = getAppPath([]string{"/"})
	if err == nil {
		t.Error("expected error, but got none")
	}

	appPath, appName, err := getAppPath([]string{"/test"})
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
