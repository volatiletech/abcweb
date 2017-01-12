package cmd

import "testing"

func TestEnvAppName(t *testing.T) {
	s := envAppName("My cool#app")
	if s != "MY_COOL_APP" {
		t.Errorf("Expected %s, got %s", "MY_COOL_APP", s)
	}
}

func TestSanitizeAppName(t *testing.T) {
	s := sanitizeAppName("My_app@cool-Name")
	if s != "My-app-cool-Name" {
		t.Errorf("Expected %s, got %s", "My-app-cool-Name", s)
	}
}

func TestDBAppName(t *testing.T) {
	s := dbAppName("My-app Name")
	if s != "my_app_name" {
		t.Errorf("Expected %s, got %s", "my_app_name", s)
	}
}
