package cert

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"time"
)

// Config contains files paths and some cert data
type Config struct {
	CertFileName string
	KeyFileName  string
	Organization []string
	Country      []string
}

// New creates and generates cert files
func New(cfg Config) (err error) {
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			Organization: cfg.Organization,
			Country:      cfg.Country,
		},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return fmt.Errorf("failed to generate rsa key: %w", err)
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %w", err)
	}

	var certPEMBuf bytes.Buffer
	pem.Encode(&certPEMBuf, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	certPEMFile, err := os.Create(cfg.CertFileName)
	if err != nil {
		return fmt.Errorf("failed to create cert.pem: %w", err)
	}

	_, err = certPEMFile.Write(certPEMBuf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to write cert.pem: %w", err)
	}

	var privateKeyPEMBuf bytes.Buffer
	pem.Encode(&privateKeyPEMBuf, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	privateKeyPEMFile, err := os.Create(cfg.KeyFileName)
	if err != nil {
		return fmt.Errorf("failed to create key.pem: %w", err)
	}

	_, err = privateKeyPEMFile.Write(privateKeyPEMBuf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to write key.pem: %w", err)
	}

	return nil
}
