// +build !bindata

package serving

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
)

type assetFile struct {
	bytes.Reader
}

func Open(name string) (ReadSeekCloser, error) {
	return os.Open(name)
}

func ReadFile(name string) ([]byte, error) {
	return ioutil.ReadFile(name)
}

func Stat(name string) (os.FileInfo, error) {
	return AssetInfo(name)
}

// Asset reads the contents of the asset for the given name.
// It returns an error if the asset could not be read.
func Asset(name string) ([]byte, error) {
	return ioutil.ReadFile(name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	b, err := ioutil.ReadFile(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return b
}

// AssetInfo attempts to Stat the asset with the given name.
// It returns an error if the Stat fails.
func AssetInfo(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

// AssetNames returns the names of the assets in the public directory.
func AssetNames() []string {
	var files []string

	err := filepath.Walk("public", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		panic(err)
	}

	err = filepath.Walk("templates", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		panic(err)
	}

	return files
}

// AssetDir returns the file names below a certain directory.
// For example with the following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	// AssetDir called with empty name will return {"public", "templates"}
	// to mimic the behavior of the go-bindata file.
	if name == "" {
		return []string{"public", "templates"}, nil
	}

	var fnames []string

	files, err := ioutil.ReadDir(name)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		fnames = append(fnames, f.Name())
	}

	return fnames, nil
}

// RestoreAsset is a mock implementation of the RestoreAsset generated
// by go-bindata.
func RestoreAsset(dir, name string) error {
	return nil
}

// RestoreAssets is a mock implementation of the RestoreAssets generated
// by go-bindata.
func RestoreAssets(dir, name string) error {
	return nil
}
