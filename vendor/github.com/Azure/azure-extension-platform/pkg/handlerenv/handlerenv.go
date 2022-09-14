// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
package handlerenv

import (
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-extension-platform/pkg/extensionerrors"
	"github.com/Azure/azure-extension-platform/pkg/utils"
	"io/ioutil"
	"os"
	"path/filepath"
)

const handlerEnvFileName = "HandlerEnvironment.json"

// HandlerEnvironment describes the handler environment configuration for an extension
type HandlerEnvironment struct {
	HeartbeatFile       string
	StatusFolder        string
	ConfigFolder        string
	LogFolder           string
	DataFolder          string
	EventsFolder        string
	DeploymentID        string
	RoleName            string
	Instance            string
	HostResolverAddress string
}

// HandlerEnvironment describes the handler environment configuration presented
// to the extension handler by the Azure Guest Agent.
type handlerEnvironmentInternal struct {
	Version            float64 `json:"version"`
	Name               string  `json:"name"`
	HandlerEnvironment struct {
		HeartbeatFile       string `json:"heartbeatFile"`
		StatusFolder        string `json:"statusFolder"`
		ConfigFolder        string `json:"configFolder"`
		LogFolder           string `json:"logFolder"`
		EventsFolder        string `json:"eventsFolder"`
		EventsFolderPreview string `json:"eventsFolder_preview"`
		DeploymentID        string `json:"deploymentid"`
		RoleName            string `json:"rolename"`
		Instance            string `json:"instance"`
		HostResolverAddress string `json:"hostResolverAddress"`
	}
}

// GetHandlerEnv locates the HandlerEnvironment.json file by assuming it lives
// next to or one level above the extension handler (read: this) executable,
// reads, parses and returns it.
func GetHandlerEnvironment(name, version string) (he *HandlerEnvironment, _ error) {
	contents, _, err := findAndReadFile(handlerEnvFileName)
	if err != nil {
		return nil, err
	}

	handlerEnvInternal, err := parseHandlerEnv(contents)
	if err != nil {
		return nil, err
	}

	// TODO: before this API goes public, remove the eventsfolder_preview
	// This is only used for private preview of the events
	eventsFolder := handlerEnvInternal.HandlerEnvironment.EventsFolder
	if eventsFolder == "" {
		eventsFolder = handlerEnvInternal.HandlerEnvironment.EventsFolderPreview
	}

	dataFolder := utils.GetDataFolder(name, version)
	return &HandlerEnvironment{
		HeartbeatFile:       handlerEnvInternal.HandlerEnvironment.HeartbeatFile,
		StatusFolder:        handlerEnvInternal.HandlerEnvironment.StatusFolder,
		ConfigFolder:        handlerEnvInternal.HandlerEnvironment.ConfigFolder,
		LogFolder:           handlerEnvInternal.HandlerEnvironment.LogFolder,
		DataFolder:          dataFolder,
		EventsFolder:        eventsFolder,
		DeploymentID:        handlerEnvInternal.HandlerEnvironment.DeploymentID,
		RoleName:            handlerEnvInternal.HandlerEnvironment.RoleName,
		Instance:            handlerEnvInternal.HandlerEnvironment.Instance,
		HostResolverAddress: handlerEnvInternal.HandlerEnvironment.HostResolverAddress,
	}, nil
}

// ParseHandlerEnv parses the HandlerEnvironment.json format.
func parseHandlerEnv(b []byte) (*handlerEnvironmentInternal, error) {
	var hf []handlerEnvironmentInternal

	if err := json.Unmarshal(b, &hf); err != nil {
		return nil, fmt.Errorf("vmextension: failed to parse handler env: %v", err)
	}
	if len(hf) != 1 {
		return nil, fmt.Errorf("vmextension: expected 1 config in parsed HandlerEnvironment, found: %v", len(hf))
	}
	return &hf[0], nil
}

// findAndReadFile locates the specified file on disk relative to our currently
// executing process and attempts to read the file
func findAndReadFile(fileName string) (b []byte, fileLoc string, _ error) {
	dir, err := utils.GetCurrentProcessWorkingDir()
	if err != nil {
		return nil, "", fmt.Errorf("vmextension: cannot find base directory of the running process: %v", err)
	}

	paths := []string{
		filepath.Join(dir, fileName),       // this level (i.e. executable is in [EXT_NAME]/.)
		filepath.Join(dir, "..", fileName), // one up (i.e. executable is in [EXT_NAME]/bin/.)
	}

	for _, p := range paths {
		o, err := ioutil.ReadFile(p)
		if err != nil && !os.IsNotExist(err) {
			return nil, "", fmt.Errorf("vmextension: error examining '%s' at '%s': %v", fileName, p, err)
		} else if err == nil {
			fileLoc = p
			b = o
			break
		}
	}

	if b == nil {
		return nil, "", extensionerrors.ErrNotFound
	}

	return b, fileLoc, nil
}
