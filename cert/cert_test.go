package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/spf13/afero"
	"github.com/volatiletech/abcweb/config"
)

func init() {
	config.AppFS = afero.NewMemMapFs()
}

func TestTemplate(t *testing.T) {
	t.Parallel()

	template, err := Template("test1", "test2")
	if template == nil {
		t.Error("got nil template")
	}
	if err != nil {
		t.Error(err)
	}
	if template.Subject.Organization[0] != "test1" {
		t.Errorf("Expected org to be %q, got %q", "test1", template.Subject.Organization[0])
	}
	if template.Subject.CommonName != "test2" {
		t.Errorf("Expected commonname to be %q, got %q", "test2", template.Subject.CommonName)
	}
}

func TestWriteCertFile(t *testing.T) {
	template, err := Template("appname", "localhost")

	fh, err := config.AppFS.Create("cert.pem")
	if err != nil {
		t.Error(err)
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Error(err)
	}

	err = WriteCertFile(fh, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Error(err)
	}
	err = fh.Close()
	if err != nil {
		t.Error(err)
	}

	out, err := afero.ReadFile(config.AppFS, "cert.pem")
	if err != nil {
		t.Error(err)
	}
	if len(out) == 0 {
		t.Error("key not written to file")
	}

	config.AppFS = afero.NewMemMapFs()
}

func TestWritePrivateKey(t *testing.T) {
	fh, err := config.AppFS.Create("private.key")
	if err != nil {
		t.Error(err)
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Error(err)
	}

	err = WritePrivateKey(fh, privateKey)
	if err != nil {
		t.Error(err)
	}
	err = fh.Close()
	if err != nil {
		t.Error(err)
	}

	out, err := afero.ReadFile(config.AppFS, "private.key")
	if err != nil {
		t.Error(err)
	}
	if len(out) == 0 {
		t.Error("key not written to file")
	}

	config.AppFS = afero.NewMemMapFs()
}
