package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"math/big"
	"os"
	"time"
)

var certs = make(map[string]*tls.Certificate)

func FetchCertificate(site string) (*tls.Certificate, error) {
	if cert, exists := certs[site]; exists {
		return cert, nil
	}

	// Load CA certificate
	caCertPEM, err := ioutil.ReadFile("data/root_ca/ca.crt")
	if err != nil {
		return nil, err
	}
	caPrivKeyPEM, err := ioutil.ReadFile("data/root_ca/ca.key")
	if err != nil {
		return nil, err
	}

	// Decode CA cert
	block, _ := pem.Decode(caCertPEM)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, errors.New("could not decode certificate")
	}
	caCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}

	// Decode CA private key
	block, _ = pem.Decode(caPrivKeyPEM)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("could not decode private key")
	}
	caPrivKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	// Generate private key for site
	certKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	// Generate serial number
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, err
	}

	certTemplate := x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:   site,
			Organization: []string{"Org"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0), // Valid for 1 year
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:              []string{site},
		BasicConstraintsValid: true,
	}

	// Sign the certificate with the CA
	certBytes, err := x509.CreateCertificate(rand.Reader, &certTemplate, caCert, &certKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, err
	}

	// Encode cert and cert key to PEM
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	certKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(certKey)})

	cert, err := tls.X509KeyPair(certPEM, certKeyPEM)
	if err != nil {
		return nil, err
	}

	certs[site] = &cert
	return &cert, nil
}

func GenerateRootCA() error {
	// Create private key for CA
	priv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}

	// Generate serial number
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return err
	}

	template := x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Organization: []string{"Proxy"},
			CommonName:   "Proxy",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0), // Valid for 10 years
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{},
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            0,
	}

	cert, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return err
	}

	f, err := os.Create("root_ca/ca.crt")
	if err != nil {
		return err
	}

	pem.Encode(f, &pem.Block{Type: "CERTIFICATE", Bytes: cert})
	f.Close()

	f, err = os.Create("root_ca/ca.key")
	if err != nil {
		return err
	}

	pem.Encode(f, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	f.Close()

	return nil
}

func RootCAExists() bool {
	_, err := os.Stat("data/root_ca/ca.crt")
	if err != nil {
		return false
	}

	_, err = os.Stat("data/root_ca/ca.key")
	if err != nil {
		return false
	}

	return true
}
