package socks5

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/redawl/gitm/internal/db"
	"github.com/redawl/gitm/internal/util"
)

// AddHostname creates a certificate for hostname, and adds it to the sqlite db stored in the config dir.
func AddHostname(hostname string) (*db.DomainInfo, error) {
	ca, caPrivKey, err := getCaCert()
	if err != nil {
		return nil, err
	}

	serialNumber, err := createSerialNumer()
	if err != nil {
		return nil, err
	}

	subjectKeyId := sha1.Sum(serialNumber.Bytes())

	cert := &x509.Certificate{
		SerialNumber: serialNumber,
		Issuer:       *getName(),
		Subject: pkix.Name{
			CommonName: hostname,
		}, DNSNames: []string{hostname},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		SubjectKeyId: subjectKeyId[:],
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth,
		},
		KeyUsage: x509.KeyUsageDigitalSignature,
	}

	certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &certPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, err
	}

	certPem := new(bytes.Buffer)
	caPem := new(bytes.Buffer)
	certPrivKeyPem := new(bytes.Buffer)

	if err := pem.Encode(certPem, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	}); err != nil {
		return nil, err
	}

	if err := pem.Encode(caPem, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: ca.Raw,
	}); err != nil {
		return nil, err
	}

	if err := pem.Encode(certPrivKeyPem, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	}); err != nil {
		return nil, err
	}
	domainCert := append(certPem.Bytes(), caPem.Bytes()...)
	if err := db.AddDomain(hostname, domainCert, certPrivKeyPem.Bytes()); err != nil {
		return nil, err
	}

	return &db.DomainInfo{
		Domain:  hostname,
		Cert:    domainCert,
		PrivKey: certPrivKeyPem.Bytes(),
	}, nil
}

func getCaCert() (*x509.Certificate, *rsa.PrivateKey, error) {
	configDir, err := util.GetConfigDir()
	if err != nil {
		return nil, nil, err
	}

	certLocation := configDir + "/ca.crt"

	if _, err := os.Stat(certLocation); errors.Is(err, os.ErrNotExist) {
		serialNumber, err := createSerialNumer()
		if err != nil {
			return nil, nil, err
		}

		ca := &x509.Certificate{
			SerialNumber: serialNumber,
			Subject:      *getName(),
			NotBefore:    time.Now(),
			NotAfter:     time.Now().AddDate(1, 0, 0),
			IsCA:         true,
			ExtKeyUsage: []x509.ExtKeyUsage{
				x509.ExtKeyUsageServerAuth,
			},
			KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
			BasicConstraintsValid: true,
		}

		caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			return nil, nil, err
		}

		caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
		if err != nil {
			return nil, nil, err
		}

		caPem := new(bytes.Buffer)
		caPrivKeyPem := new(bytes.Buffer)
		if err := pem.Encode(caPem, &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: caBytes,
		}); err != nil {
			return nil, nil, err
		}

		if err := pem.Encode(caPrivKeyPem, &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
		}); err != nil {
			return nil, nil, err
		}

		if err := os.WriteFile(configDir+"/ca.crt", caBytes, 0400); err != nil {
			return nil, nil, err
		}
		if err := os.WriteFile(configDir+"/ca.pem", caPem.Bytes(), 0400); err != nil {
			return nil, nil, err
		}
		if err := os.WriteFile(configDir+"/privkey.pem", caPrivKeyPem.Bytes(), 0400); err != nil {
			return nil, nil, err
		}
	}

	caPem, err := os.ReadFile(configDir + "/ca.pem")
	if err != nil {
		return nil, nil, err
	}

	caBlock, rest := pem.Decode(caPem)

	if caBlock == nil || len(rest) > 0 {
		return nil, nil, fmt.Errorf("parsing ca.pem, leftover bytes")
	}

	privKeyPem, err := os.ReadFile(configDir + "/privkey.pem")
	if err != nil {
		return nil, nil, err
	}

	privKeyBlock, rest := pem.Decode(privKeyPem)

	if privKeyBlock == nil || len(rest) > 0 {
		return nil, nil, fmt.Errorf("parsing privkey.pem, leftover bytes")
	}

	caCert, err := x509.ParseCertificate(caBlock.Bytes)
	if err != nil {
		return nil, nil, err
	}

	privKey, err := x509.ParsePKCS1PrivateKey(privKeyBlock.Bytes)
	if err != nil {
		return nil, nil, err
	}

	return caCert, privKey, nil
}

func createSerialNumer() (*big.Int, error) {
	return rand.Int(rand.Reader, big.NewInt(999999999999999999))
}

func getName() *pkix.Name {
	return &pkix.Name{
		CommonName:         "GITM Inc",
		OrganizationalUnit: []string{"GITM Inc"},
		Organization:       []string{"GITM Inc"},
		Country:            []string{"GITM Inc"},
		Province:           []string{"GITM Inc"},
		Locality:           []string{"GITM Inc"},
	}
}

func InitCaCert() error {
	_, _, err := getCaCert()
	if err != nil {
		return fmt.Errorf("init ca cert: %v", err)
	}

	return nil
}
