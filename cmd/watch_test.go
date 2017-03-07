package cmd

import (
	"testing"

	"github.com/spf13/afero"
)

func TestFindAllDirs(t *testing.T) {
	AppFS = afero.NewMemMapFs()

	AppFS.MkdirAll("/project/assets/js/admin", 0774)
	AppFS.MkdirAll("/project/assets/css", 0774)

	dirs, err := findAllDirs("/project/assets")
	if err != nil {
		t.Error(err)
	}

	set := map[string]struct{}{}
	for _, d := range dirs {
		set[d] = struct{}{}
	}

	expect := []string{
		"/project/assets",
		"/project/assets/js",
		"/project/assets/js/admin",
		"/project/assets/css",
	}

	for _, e := range expect {
		if _, ok := set[e]; !ok {
			t.Errorf("did not find: %s", e)
		}
	}

	if t.Failed() {
		t.Log("set:", set)
	}
}
