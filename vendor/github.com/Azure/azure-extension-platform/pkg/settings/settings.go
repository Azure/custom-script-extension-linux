// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
package settings

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/Azure/azure-extension-platform/pkg/decrypt"
	"github.com/Azure/azure-extension-platform/pkg/extensionerrors"
	"github.com/Azure/azure-extension-platform/pkg/handlerenv"
	"github.com/Azure/azure-extension-platform/pkg/logging"
)

const (
	settingsFileSuffix = ".settings"
	disableFileName    = "disabled"
)

// HandlerSettings contains the decrypted settings for the extension
type HandlerSettings struct {
	// string containing the json sent for the public settings
	PublicSettings string

	// string containing the decrypted protected settings
	ProtectedSettings string
}

// handlerSettings is an internal structure used to deserialize the file
type handlerSettings struct {
	PublicSettings          interface{} `json:"publicSettings"`
	ProtectedSettingsBase64 string      `json:"protectedSettings"`
	SettingsCertThumbprint  string      `json:"protectedSettingsCertThumbprint"`
}

type handlerSettingsFile struct {
	RuntimeSettings []handlerSettingsContainer `json:"runtimeSettings"`
}

type handlerSettingsContainer struct {
	HandlerSettings handlerSettings `json:"handlerSettings"`
}

// GetHandlerSettings reads and parses the handler's settings in an OS independent manner
func GetHandlerSettings(el *logging.ExtensionLogger, he *handlerenv.HandlerEnvironment, seqNo uint) (hs *HandlerSettings, _ error) {
	// The file will be under the config folder with the path {seqNo}.settings
	settingsFileName := filepath.Join(he.ConfigFolder, fmt.Sprintf("%d%s", seqNo, settingsFileSuffix))
	parsedHs, err := parseHandlerSettingsFile(el, settingsFileName)
	if err != nil {
		return hs, err
	}

	protectedSettings, err := unmarshalProtectedSettings(el, he.ConfigFolder, parsedHs)
	if err != nil {
		return hs, err
	}

	var publicSettingJsonString = ""
	// parsedHS.PublicSettings is an interface, has to be marshaled to get the string representation
	if parsedHs.PublicSettings != nil {
		jsonBytes, err := json.Marshal(parsedHs.PublicSettings)
		if err != nil {
			return hs, err
		}
		publicSettingJsonString = string(jsonBytes)
	}

	hs = &HandlerSettings{
		PublicSettings:    publicSettingJsonString,
		ProtectedSettings: protectedSettings,
	}

	return hs, nil
}

// unmarshalProtectedSettings decodes the protected settings from handler
// runtime settings JSON file, decrypts it using the certificates and unmarshals
// into the given struct v.
func unmarshalProtectedSettings(el *logging.ExtensionLogger, configFolder string, hs handlerSettings) (string, error) {
	if hs.ProtectedSettingsBase64 == "" {
		// No protected settings
		return "", nil
	}
	if hs.SettingsCertThumbprint == "" {
		el.Error("parseHandlerSettingsFile failed due to no settings cert thumbprint")
		return "", extensionerrors.ErrNoCertificateThumbprint
	}

	decoded, err := base64.StdEncoding.DecodeString(hs.ProtectedSettingsBase64)
	if err != nil {
		el.Error("parseHandlerSettingsFile failed to decode base64: %v", err)
		return "", extensionerrors.ErrInvalidProtectedSettingsData
	}

	v, err := decrypt.DecryptProtectedSettings(configFolder, hs.SettingsCertThumbprint, decoded)
	return v, err
}

// parseHandlerSettings parses a handler settings file (e.g. 0.settings) and
// returns it as a structured object.
func parseHandlerSettingsFile(el *logging.ExtensionLogger, path string) (h handlerSettings, _ error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		el.Error("parseHandlerSettingsFile failed. Error reading %s: %v", path, err)
		return h, extensionerrors.ErrInvalidSettingsFile
	}
	if len(b) == 0 { // if no config is specified, we get an empty file
		return h, nil
	}

	var f handlerSettingsFile
	if err := json.Unmarshal(b, &f); err != nil {
		el.Error("parseHandlerSettingsFile failed. error parsing json: %v", err)
		return h, extensionerrors.ErrInvalidSettingsFile
	}
	if len(f.RuntimeSettings) != 1 {
		el.Error("parseHandlerSettingsFile failed. wrong runtimeSettings count. expected:1, got:%d", len(f.RuntimeSettings))
		return h, extensionerrors.ErrInvalidSettingsRuntimeSettingsCount
	}

	return f.RuntimeSettings[0].HandlerSettings, nil
}

// CleanUpSettings replaces the protected settings for all settings files [ex: 0.settings, etc] to ensure no
// protected settings are logged in VM
func CleanUpSettings(el *logging.ExtensionLogger, configFolder string) {
	configDir, err := ioutil.ReadDir(configFolder)
	if err != nil {
		el.Error("error clearing config file: %v", err)
		return
	}
	content := []byte("")
	for _, file := range configDir {
		if strings.Compare(filepath.Ext(file.Name()), settingsFileSuffix) == 0 { //checking if its a settings file
			filePath := filepath.Join(configFolder, file.Name())
			err = ioutil.WriteFile(filePath, content, 0644)
			if err != nil {
				el.Error("error clearing %s, err %v", file.Name(), err)
			} else {
				el.Info("%s cleared successfully", file.Name())
			}
		}
	}
}
