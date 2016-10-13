package main

import "testing"

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	_, err := loadConfig()
	if err != nil {
		t.Error(err)
	}
}
