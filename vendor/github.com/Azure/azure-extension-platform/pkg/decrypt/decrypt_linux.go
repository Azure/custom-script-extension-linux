// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
package decrypt

import (
	"bytes"
	"fmt"
	"os/exec"
	"path"
	"path/filepath"
)

var getCertificateDir = func(configFolder string) (certificateFolder string) {
	return path.Join(configFolder, "..", "..")
}

// decryptProtectedSettings decrypts the read protected settigns using certificates
func DecryptProtectedSettings(configFolder string, thumbprint string, decoded []byte) (string, error) {
	// go two levels up where certs are placed (/var/lib/waagent)
	crt := filepath.Join(getCertificateDir(configFolder), fmt.Sprintf("%s.crt", thumbprint))
	prv := filepath.Join(getCertificateDir(configFolder), fmt.Sprintf("%s.prv", thumbprint))

	// we use os/exec instead of azure-docker-extension/pkg/executil here as
	// other extension handlers depend on this package for parsing handler
	// settings.
	cmd := exec.Command("openssl", "smime", "-inform", "DER", "-decrypt", "-recip", crt, "-inkey", prv)
	var bOut, bErr bytes.Buffer
	cmd.Stdin = bytes.NewReader(decoded)
	cmd.Stdout = &bOut
	cmd.Stderr = &bErr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("decrypting protected settings failed: error=%v stderr=%s", err, string(bErr.Bytes()))
	}

	v := bOut.String()
	return v, nil
}
