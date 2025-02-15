package cacert

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
	"log/slog"
	"math/big"
	"net/http"
	"os"
	"time"

	"com.github.redawl.mitmproxy/db"
	"com.github.redawl.mitmproxy/util"
)

func ListenAndServe(listenUri string) error {
    return http.ListenAndServe(listenUri, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        slog.Debug("Request to Cacert server", "path", r.URL.Path)
        if r.URL.Path == "/ca.crt" {
            configDir, err := util.GetConfigDir()

            if err != nil {
                slog.Error("Error getting config dir", "error", err)
                return
            }

            certLocation := configDir + "/ca.crt"
            contents, err := os.ReadFile(certLocation)

            if err != nil {
                slog.Error("Error getting ca cert", "error", err)
                w.WriteHeader(http.StatusInternalServerError)
                return
            }

            w.Write(contents)
        } else if r.URL.Path == "/proxy.pac" {
            contents, err := os.ReadFile("www/proxy.pac")

            if err != nil {
                slog.Error("Error getting proxy file", "error", err)
                w.WriteHeader(http.StatusInternalServerError)
                return
            }

            w.Write(contents)
        } else {
            http.Error(w, "Not found", http.StatusNotFound)
        }
    }))
}

func AddHostname (hostname string) error {
    ca, caPrivKey, err := getCaCert()

    serialNumber, err := createSerialNumer()
    
    if err != nil {
        return err
    }

    subjectKeyId := sha1.Sum(serialNumber.Bytes())

    cert := &x509.Certificate{
        SerialNumber: serialNumber,
        Issuer: *getName(),
        Subject: pkix.Name{
            CommonName: hostname,
        },
        DNSNames: []string{hostname},
        NotBefore: time.Now(),
        NotAfter: time.Now().AddDate(1, 0, 0),
        SubjectKeyId: subjectKeyId[:],
        ExtKeyUsage: []x509.ExtKeyUsage{
            x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth,
        },
        KeyUsage: x509.KeyUsageDigitalSignature,
    }

    certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)

    if err != nil {
        return err
    }

    certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &certPrivKey.PublicKey, caPrivKey)

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
        Type:  "CERTIFICATE",
        Bytes: ca.Raw,
    })

    pem.Encode(certPrivKeyPem, &pem.Block{
        Type:  "RSA PRIVATE KEY",
        Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
    })

    err = db.AddDomain(hostname, append(certPem.Bytes(), caPem.Bytes()...), certPrivKeyPem.Bytes())

    if err != nil {
        return err
    }

    return nil
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
            Subject: *getName(),
            NotBefore: time.Now(),
            NotAfter: time.Now().AddDate(1, 0, 0),
            IsCA: true,
            ExtKeyUsage: []x509.ExtKeyUsage{
                x509.ExtKeyUsageServerAuth,
            },
            KeyUsage: x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
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

        caPem   := new(bytes.Buffer)
        caPrivKeyPem   := new(bytes.Buffer)
        pem.Encode(caPem, &pem.Block{
            Type: "CERTIFICATE",
            Bytes: caBytes,
        })
        pem.Encode(caPrivKeyPem, &pem.Block{
            Type:  "RSA PRIVATE KEY",
            Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
        })
        err = os.WriteFile(configDir + "/ca.crt", caBytes, 0400)
        if err != nil {
            return nil, nil, err
        }
        err = os.WriteFile(configDir + "/ca.pem", caPem.Bytes(), 0400)
        if err != nil {
            return nil, nil, err
        }
        err = os.WriteFile(configDir + "/privkey.pem", caPrivKeyPem.Bytes(), 0400)
        if err != nil {
            return nil, nil, err
        }
    }

    caPem, err := os.ReadFile(configDir + "/ca.pem") 
    
    if err != nil {
        return nil, nil, err
    }

    caBlock, rest := pem.Decode(caPem)

    if caBlock == nil || len(rest) > 0 {
        return nil, nil, fmt.Errorf("Error parsing ca.pem")
    }

    privKeyPem, err := os.ReadFile(configDir + "/privkey.pem")

    if err != nil {
        return nil, nil, err
    }

    privKeyBlock, rest := pem.Decode(privKeyPem)

    if privKeyBlock == nil || len(rest) > 0 {
        return nil, nil, fmt.Errorf("Error parsing privkey.pem")
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

func getName() (*pkix.Name) {
    return &pkix.Name{
        CommonName: "MITMProxy Inc",
        OrganizationalUnit: []string{"MITMProxy Inc"},
        Organization: []string{"MITMProxy Inc"},
        Country: []string{"MITMProxy Inc"},
        Province: []string{"MITMProxy Inc"},
        Locality: []string{"MITMProxy Inc"},
    }
}
