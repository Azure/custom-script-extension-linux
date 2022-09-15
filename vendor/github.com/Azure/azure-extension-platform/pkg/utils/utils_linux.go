// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
package utils

import "path"

// agentDir is where the agent is located, a subdirectory of which we use as the data directory
const agentDir = "/var/lib/waagent"

func GetDataFolder(name string, version string) string {
	return path.Join(agentDir, name)
}
