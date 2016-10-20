package app

import "testing"

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	a := appState{}
	err := a.loadConfig()
	if err != nil {
		t.Error(err)
	}
}
