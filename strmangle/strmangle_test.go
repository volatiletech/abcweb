package strmangle

import "testing"

func TestEnvAppName(t *testing.T) {
	s := EnvAppName("My cool#app")
	if s != "MY_COOL_APP" {
		t.Errorf("Expected %s, got %s", "MY_COOL_APP", s)
	}

	s = EnvAppName("abcweb")
	if s != "ABCWEB" {
		t.Errorf("Expected %s, got %s", "ABCWEB", s)
	}
}

func TestSanitizeAppName(t *testing.T) {
	s := SanitizeAppName("My_app@cool-Name")
	if s != "My-app-cool-Name" {
		t.Errorf("Expected %s, got %s", "My-app-cool-Name", s)
	}
}

func TestDBAppName(t *testing.T) {
	s := DBAppName("My-app Name")
	if s != "my_app_name" {
		t.Errorf("Expected %s, got %s", "my_app_name", s)
	}
}
