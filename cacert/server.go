package cacert

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"log/slog"
	"math/big"
	"net/http"
	"os"
	"time"
)

func ListenAndServe(listenUri string) error {
    return http.ListenAndServe(listenUri, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/ca.crt" {
            userConfigDir, err := os.UserConfigDir()
            if err != nil {
                slog.Error("Failed to find configuration dir", "error", err)
                return
            }
            configDir := userConfigDir + "/mitmproxy"
            certLocation := configDir + "/ca.crt"
            contents, err := os.ReadFile(certLocation)

            w.Write(contents)
        } else {
            http.Error(w, "Not found", http.StatusNotFound)
        }
    }))
}

func SetupCertificateAuthority (hostname string, certLocation string, privKeyLocation string) error {
    // Write files
    userCfgDir, err := os.UserConfigDir()
    if err != nil {
        return err
    }
    configDir := userCfgDir + "/mitmproxy"

    if _, err := os.Stat(certLocation); errors.Is(err, os.ErrNotExist) {
        os.Mkdir(configDir, 0700)

        ca := &x509.Certificate{
            SerialNumber: big.NewInt(2023136218723618723),
            Subject: pkix.Name{
                Organization: []string{"MITMProxy Inc"},
                Country: []string{},
                Province: []string{},
                Locality: []string{},
                StreetAddress: []string{},
                PostalCode: []string{},
            },
            NotBefore: time.Now(),
            NotAfter: time.Now().AddDate(10, 0, 0),
            IsCA: true,
            ExtKeyUsage: []x509.ExtKeyUsage{
                x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth,
            },
            KeyUsage: x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
            BasicConstraintsValid: true,
        }

        caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
        if err != nil {
            return err
        }

        caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)

        if err != nil {
            return err
        }

        cert := &x509.Certificate{
            SerialNumber: big.NewInt(2023136218723618723),
            Issuer: pkix.Name{
                Organization: []string{"MITMProxy Inc"},
                Country: []string{},
                Province: []string{},
                Locality: []string{},
                StreetAddress: []string{},
                PostalCode: []string{},
            },
            Subject: pkix.Name{
                Organization: []string{"MITMProxy Inc"},
                Country: []string{},
                Province: []string{},
                Locality: []string{},
                StreetAddress: []string{},
                PostalCode: []string{},
            },
            DNSNames: []string{hostname},
            NotBefore: time.Now(),
            NotAfter: time.Now().AddDate(10, 0, 0),
            SubjectKeyId: []byte{1,2,3,4,5,6},
            ExtKeyUsage: []x509.ExtKeyUsage{
                x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth,
            },
            KeyUsage: x509.KeyUsageDigitalSignature,
        }

        certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)

        if err != nil {
            return err
        }

        certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &certPrivKey.PublicKey, certPrivKey)

        if err != nil {
            return err
        }

        certPem := new(bytes.Buffer)
        caPem   := new(bytes.Buffer)
        certPrivKeyPem := new(bytes.Buffer)

        pem.Encode(certPem, &pem.Block{
            Type:  "CERTIFICATE",
            Bytes: certBytes,
        })

        pem.Encode(caPem, &pem.Block{
            Type: "CERTIFICATE",
            Bytes: caBytes,
        })

        pem.Encode(certPrivKeyPem, &pem.Block{
            Type:  "RSA PRIVATE KEY",
            Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
        })

        err = os.WriteFile(configDir + "/ca.crt", certBytes, 0400)
        if err != nil {
            return err
        }
        err = os.WriteFile(configDir + "/ca.pem", certPem.Bytes(), 0400)
        if err != nil {
            return err
        }
        err = os.WriteFile(configDir + "/server.pem", append(certPem.Bytes(), caPem.Bytes()...), 0400)
        if err != nil {
            return err
        }
        err = os.WriteFile(configDir + "/privkey.pem", certPrivKeyPem.Bytes(), 0400)
        if err != nil {
            return err
        }
    }
    return nil
}

