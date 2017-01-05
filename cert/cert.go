package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"time"

	"github.com/spf13/afero"
)

// Template is a helper function to create a cert template with a
// serial number and other required fields
func Template(appName, commonName string) (*x509.Certificate, error) {
	// generate a random serial number (a real cert authority would have some logic behind this)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, errors.New("failed to generate serial number: " + err.Error())
	}

	tmpl := x509.Certificate{
		IsCA:         true,
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{appName},
			CommonName:   commonName,
		},

		SignatureAlgorithm:    x509.SHA256WithRSA,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour * 24 * 365 * 4), // valid for 4 years
		BasicConstraintsValid: true,
	}

	return &tmpl, nil
}

// WriteCertFile writes the cert.pem certificate file
func WriteCertFile(outFile afero.File, template *x509.Certificate, pub interface{}, priv interface{}) error {
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, pub, priv)
	if err != nil {
		return err
	}

	// PEM encode the certificate (this is a standard TLS encoding)
	b := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	}

	return pem.Encode(outFile, b)
}

// WritePrivateKey writes the private.key private key file
func WritePrivateKey(outFile afero.File, key *rsa.PrivateKey) error {
	b := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	return pem.Encode(outFile, b)
}
