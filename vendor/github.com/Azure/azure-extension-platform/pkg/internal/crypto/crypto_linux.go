// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"github.com/Azure/azure-extension-platform/pkg/extensionerrors"
	"math/big"
	"os"
	"strings"
	"time"
)

const (
	COUNTRY             = "US"
	LOCALITY            = "Redmond"
	COMMON_NAME         = "Hybrid Runbook Worker"
	ORGANIZATIONAL_UNIT = "Azure Automation"
	ORGANIZATION        = "Microsoft Corporation"
	STATE               = "Washington"
)

type SelfSignedCertificateKey struct {
	Cert    x509.Certificate
	PrivKey rsa.PrivateKey
}

type CertificateOperations interface {
	WriteCertificateToDisk(certificateOutputPath string) error
	WriteKeyToDisk(keyOutputPath string) error
	GetCertificateThumbprint() string
}

func NewSelfSignedx509Certificate() (*SelfSignedCertificateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)

	if err != nil {
		return nil, extensionerrors.AddStackToError(err)
	}

	certSubject := pkix.Name{
		Country:            []string{COUNTRY},
		Locality:           []string{LOCALITY},
		Province:           []string{STATE},
		Organization:       []string{ORGANIZATION},
		OrganizationalUnit: []string{ORGANIZATIONAL_UNIT},
		CommonName:         COMMON_NAME,
		SerialNumber:       "666",
	}
	certTemplate := x509.Certificate{
		Subject:               certSubject,
		NotBefore:             time.Unix(0, 0),
		NotAfter:              time.Now().Add(time.Hour * 24 * 365 * 10), // 10 years from now
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		SerialNumber:          big.NewInt(666),
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &certTemplate, &certTemplate, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, extensionerrors.AddStackToError(err)
	}

	x509Cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return nil, extensionerrors.AddStackToError(err)
	}

	return &SelfSignedCertificateKey{Cert: *x509Cert, PrivKey: *privateKey}, nil
}

func (cert *SelfSignedCertificateKey) WriteCertificateToDisk(certificateOutputPath string) error {
	certFH, err := os.OpenFile(certificateOutputPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return extensionerrors.AddStackToError(err)
	}

	if err = pem.Encode(certFH, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Cert.Raw}); err != nil {
		return extensionerrors.AddStackToError(err)
	}
	if err = certFH.Close(); err != nil {
		return extensionerrors.AddStackToError(err)
	}
	return nil
}

func (cert *SelfSignedCertificateKey) WriteKeyToDisk(keyOutputPath string) error {
	keyFH, err := os.OpenFile(keyOutputPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return extensionerrors.AddStackToError(err)
	}
	if err := pem.Encode(keyFH, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(&cert.PrivKey)}); err != nil {
		return extensionerrors.AddStackToError(err)
	}
	if err := keyFH.Close(); err != nil {
		return extensionerrors.AddStackToError(err)
	}
	return nil
}

func (cert *SelfSignedCertificateKey) GetCertificateThumbprint() string {
	sigBytes := sha1.Sum(cert.Cert.Raw)
	return strings.ToUpper(hex.EncodeToString(sigBytes[:]))
}
