package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"os"
	"time"
)

// template is a helper function to create a cert template with a
// serial number and other required fields
func template(appName, commonName string) (*x509.Certificate, error) {
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
func WriteCertFile(path, appName, commonName string, pub interface{}, parentPriv interface{}) error {
	template, err := template(appName, commonName)
	if err != nil {
		return err
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, pub, parentPriv)
	if err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}

	// PEM encode the certificate (this is a standard TLS encoding)
	b := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	}

	return pem.Encode(file, b)
}

// WritePrivateKey writes the private.key private key file
func WritePrivateKey(path string, key *rsa.PrivateKey) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	b := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	return pem.Encode(file, b)
}
