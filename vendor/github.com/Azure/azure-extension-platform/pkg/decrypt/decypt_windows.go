// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
package decrypt

import (
	"bytes"
	"fmt"
	"strconv"
	"syscall"
	"unsafe"

	"github.com/Azure/azure-extension-platform/pkg/extensionerrors"
	"github.com/Azure/azure-extension-platform/pkg/internal/crypto"
	"golang.org/x/sys/windows"
)

type cryptDecryptMessagePara struct {
	cbSize                   uint32
	dwMsgAndCertEncodingType uint32
	cCertStore               uint32
	rghCertStore             uintptr
	dwFlags                  uint32
}

// decryptProtectedSettings decrypts the read protected settings using certificates
func DecryptProtectedSettings(configFolder string, thumbprint string, decoded []byte) (string, error) {
	// Open My/Local
	handle, err := syscall.CertOpenStore(windows.CERT_STORE_PROV_SYSTEM, 0, 0, windows.CERT_SYSTEM_STORE_LOCAL_MACHINE, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("MY"))))
	if err != nil {
		return "", fmt.Errorf("VMextension: Cannot open certificate store due to '%v'", err)
	}
	if handle == 0 {
		return "", extensionerrors.ErrMustRunAsAdmin
	}
	defer syscall.CertCloseStore(handle, 0)

	// Convert the thumbprint to bytes. We do byte comparison vs string comparison because otherwise we'd need to normalize the strings
	decodedThumbprint, err := thumbprintStringToHex(thumbprint)
	if err != nil {
		return "", fmt.Errorf("VmExtension: Invalid thumbprint")
	}

	// Find the certificate by thumbprint
	const crypteENotFound = 0x80092004
	var cert *syscall.CertContext
	var prevContext *syscall.CertContext
	found := false
	for {
		// Keep retrieving the next certificate
		cert, err := syscall.CertEnumCertificatesInStore(handle, prevContext)
		if err != nil {
			if errno, ok := err.(syscall.Errno); ok {
				if errno == crypteENotFound {
					// We've reached the last certificate
					break
				}
			}
			return "", fmt.Errorf("VmExtension: Could not enumerate certificates due to '%v'", err)
		}

		if cert == nil {
			break
		}

		// Determine the cert thumbprint
		foundthumbprint, err := crypto.GetCertificateThumbprint(cert)
		if err == nil && foundthumbprint != nil {
			// TODO: consider logging if we have an error. For now, we just ignore the cert
			if bytes.Compare(decodedThumbprint, foundthumbprint) == 0 {
				found = true
				break
			}
		}

		prevContext = cert
	}

	if !found {
		return "", extensionerrors.ErrCertWithThumbprintNotFound
	}

	// Decrypt the protected settings
	decryptedBytes, err := decryptDataWithCert(decoded, cert, uintptr(handle))
	if err != nil {
		return "", err
	}

	// Now deserialize the data
	v := string(decryptedBytes[:])
	return v, err
}

func thumbprintStringToHex(s string) ([]byte, error) {
	// Remove the UTF mark if we have one
	runes := []rune(s)
	if len(runes)%2 == 1 {
		runes = []rune(s)[1:]
	}

	length := len(runes) / 2
	parts := make([]byte, length)
	for count := 0; count < length; count++ {
		r := runes[count*2 : count*2+2]
		sp := string(r)
		bp, err := strconv.ParseUint(sp, 16, 16)
		if err == nil {
			parts[count] = byte(bp)
		}
	}

	return parts, nil
}

// decryptDataWithCert calls the Windows APIs to do the decryption
func decryptDataWithCert(decoded []byte, cert *syscall.CertContext, storeHandle uintptr) ([]byte, error) {
	var cryptDecryptMessagePara cryptDecryptMessagePara
	cryptDecryptMessagePara.cbSize = uint32(len(decoded))
	cryptDecryptMessagePara.dwMsgAndCertEncodingType = uint32(windows.X509_ASN_ENCODING | windows.PKCS_7_ASN_ENCODING)
	cryptDecryptMessagePara.cCertStore = uint32(1)
	cryptDecryptMessagePara.rghCertStore = uintptr(unsafe.Pointer(&storeHandle))
	cryptDecryptMessagePara.dwFlags = uint32(0)

	// Call it once to get the decrypted data size
	var pbEncryptedBlob *byte
	var cbDecryptedBlob uint32
	pbEncryptedBlob = &decoded[0]
	raw, _, err := syscall.Syscall6(
		crypto.ProcCryptDecryptMessage.Addr(),
		6,
		uintptr(unsafe.Pointer(&cryptDecryptMessagePara)),
		uintptr(unsafe.Pointer(pbEncryptedBlob)),
		uintptr(len(decoded)),
		uintptr(0),
		uintptr(unsafe.Pointer(&cbDecryptedBlob)),
		uintptr(0),
	)
	if raw == 0 {
		errno := syscall.Errno(err)
		if errno == crypto.CrypteEAsn1BadTag {
			return nil, extensionerrors.ErrInvalidProtectedSettingsData
		}

		return nil, fmt.Errorf("VmExtension: Could not decrypt data due to '%d'", syscall.Errno(err))
	}

	// Create our buffer
	if cbDecryptedBlob == 0 {
		return nil, nil
	}

	var decryptedBytes = make([]byte, cbDecryptedBlob)
	var pdecryptedBytes *byte
	pdecryptedBytes = &decryptedBytes[0]

	raw, _, err = syscall.Syscall6(
		crypto.ProcCryptDecryptMessage.Addr(),
		6,
		uintptr(unsafe.Pointer(&cryptDecryptMessagePara)),
		uintptr(unsafe.Pointer(pbEncryptedBlob)),
		uintptr(len(decoded)),
		uintptr(unsafe.Pointer(pdecryptedBytes)),
		uintptr(unsafe.Pointer(&cbDecryptedBlob)),
		uintptr(0),
	)
	if raw == 0 {
		return nil, fmt.Errorf("VmExtension: Could not decrypt data due to '%d'", syscall.Errno(err))
	}

	// Get rid of the null terminator or deserialization will fail
	returnedBytes := decryptedBytes[:cbDecryptedBlob]

	return returnedBytes, nil
}
