// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
package utils

import (
	"os"
	"path/filepath"
)

// GetCurrentProcessWorkingDir returns the absolute path of the running process.
func GetCurrentProcessWorkingDir() (string, error) {
	p, err := filepath.Abs(os.Args[0])
	if err != nil {
		return "", err
	}
	return filepath.Dir(p), nil
}
