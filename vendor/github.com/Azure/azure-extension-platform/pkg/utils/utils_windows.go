// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
package utils

import (
	"os"
	"path"
)

func GetDataFolder(name string, version string) string {
	systemDriveFolder := os.Getenv("SystemDrive")
	return path.Join(systemDriveFolder, "Packages\\Plugins", name, version, "Downloads")
}
