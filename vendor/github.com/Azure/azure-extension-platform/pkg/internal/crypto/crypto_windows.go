// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
package crypto

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"
)

const (
	CertHashPropID    = 3
	CrypteEAsn1BadTag = 2148086027
	certNCryptKeySpec = 0xFFFFFFFF
)

var (
	Modcrypt32                            = syscall.NewLazyDLL("crypt32.dll")
	procCertGetCertificateContextProperty = Modcrypt32.NewProc("CertGetCertificateContextProperty")
	ProcCryptDecryptMessage               = Modcrypt32.NewProc("CryptDecryptMessage")
	procCryptAcquireCertificatePrivateKey = Modcrypt32.NewProc("CryptAcquireCertificatePrivateKey")
	procNCryptFreeObject                  = Modcrypt32.NewProc("NCryptFreeObject")
)

type CryptAlgorithmIdentifier struct {
	PszObjID   uintptr
	Parameters CryptObjectIDBlob
}

type cryptIntegerBlob struct {
	cbData uint32
	pbData uintptr
}

type CryptObjectIDBlob struct {
	CbData uint32
	PbData uintptr
}

type cryptBitBlob struct {
	cbData      uint32
	pbData      uintptr
	cUnusedBits uint32
}

type certNameBlob struct {
	cbData uint32
	pbData uintptr
}

type certPublicKeyInfo struct {
	Algorithm CryptAlgorithmIdentifier
	PublicKey cryptBitBlob
}

// This struct is not implemented in syscall, so we need to do this ourselves
type certInfo struct {
	dwVersion            uint32
	serialNumber         cryptIntegerBlob
	signatureAlgorithm   CryptAlgorithmIdentifier
	issuer               certNameBlob
	notBefore            syscall.Filetime
	notAfter             syscall.Filetime
	subject              certNameBlob
	subjectPublicKeyInfo certPublicKeyInfo
	issuerUniqueID       cryptBitBlob
	subjectUniqueID      cryptBitBlob
	cExtension           uint32
	rgExtension          uintptr
}

type CertContext struct {
	EncodingType uint32
	EncodedCert  *byte
	Length       uint32
	CertInfo     *certInfo
	Store        syscall.Handle
}

// We look for a cert with the following
// - Not expired
// - Has a private key
// Note that the dev code uses syscall.CertContext. However that doesn't have the CERT_INFO
// structure, so we need to find the cert manually, then convert it to the syscall structure
func GetAUsableCert(handle syscall.Handle) (cert *syscall.CertContext, _ error) {
	var testCert *CertContext
	var prevCert *CertContext
	procCertEnumCertificatesInStore := Modcrypt32.NewProc("CertEnumCertificatesInStore")

	for {
		ret, _, _ := syscall.Syscall(
			procCertEnumCertificatesInStore.Addr(),
			2,
			uintptr(handle),
			uintptr(unsafe.Pointer(prevCert)),
			0)

		// Not that we don't handle ENotFound, since that's an error case for us (we couldn't find a cert)
		testCert = (*CertContext)(unsafe.Pointer(ret))
		usable := isAUsableCert(testCert)
		if usable {
			// We need a syscall.CertContext
			syscallContext := (*syscall.CertContext)(unsafe.Pointer(ret))
			return syscallContext, nil
		}

		prevCert = testCert
	}
}

func isAUsableCert(cert *CertContext) (usable bool) {
	// First check if the cert has expired
	ended := time.Unix(0, cert.CertInfo.notAfter.Nanoseconds())
	started := time.Unix(0, cert.CertInfo.notBefore.Nanoseconds())
	now := time.Now()
	if now.After(ended) || now.Before(started) {
		return false
	}

	// Check that it has a private key
	if !hasPrivateKey(cert) {
		return false
	}

	return true
}

func hasPrivateKey(cert *CertContext) bool {
	var ncryptKeyHandle uintptr
	var dwKeySpec uint32
	var fCallerFreeProvOrNCryptKey uint32
	ret, _, err := syscall.Syscall6(
		procCryptAcquireCertificatePrivateKey.Addr(),
		6,
		uintptr(unsafe.Pointer(cert)),
		uintptr(0),
		uintptr(0),
		uintptr(unsafe.Pointer(&ncryptKeyHandle)),
		uintptr(unsafe.Pointer(&dwKeySpec)),
		uintptr(unsafe.Pointer(&fCallerFreeProvOrNCryptKey)))
	if ret == 0 {
		if err > 0 {
			// If for some reason we can't retrieve the private key, move on
			return false
		}
	}

	// Figure out if we need to release the handle
	if fCallerFreeProvOrNCryptKey != 0 {
		if dwKeySpec == certNCryptKeySpec {
			// We received an CERT_NCRYPT_KEY_SPEC
			syscall.Syscall(
				procNCryptFreeObject.Addr(),
				1,
				uintptr(ncryptKeyHandle),
				0,
				0)
		} else {
			handle := syscall.Handle(ncryptKeyHandle)
			syscall.CryptReleaseContext(handle, 0)
		}
	}

	return true
}

func GetCertificateThumbprint(cert *syscall.CertContext) ([]byte, error) {
	// Call it once to retrieve the thumbprint size
	var cbComputedHash uint32
	ret, _, err := syscall.Syscall6(
		procCertGetCertificateContextProperty.Addr(),
		4,
		uintptr(unsafe.Pointer(cert)),            // pCertContext
		uintptr(CertHashPropID),                  // dwPropId
		uintptr(0),                               // pvData)
		uintptr(unsafe.Pointer(&cbComputedHash)), // pcbData
		0,
		0,
	)

	if ret == 0 {
		return nil, fmt.Errorf("VmExtension: Could not hash certificate due to '%d'", syscall.Errno(err))
	}

	// Create our buffer
	if cbComputedHash == 0 {
		return nil, nil
	}

	var computedHashBuffer = make([]byte, cbComputedHash)
	var pComputedHash *byte
	pComputedHash = &computedHashBuffer[0]
	ret, _, err = syscall.Syscall6(
		procCertGetCertificateContextProperty.Addr(),
		4,
		uintptr(unsafe.Pointer(cert)),            // pCertContext
		uintptr(CertHashPropID),                  // dwPropId
		uintptr(unsafe.Pointer(pComputedHash)),   // pvData)
		uintptr(unsafe.Pointer(&cbComputedHash)), // pcbData
		0,
		0,
	)
	if ret == 0 {
		return nil, fmt.Errorf("VmExtension: Could not hash certificate due to '%d'", syscall.Errno(err))
	}

	return computedHashBuffer, nil
}
