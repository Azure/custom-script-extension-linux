package main

import (
	"github.com/Azure/azure-docker-extension/pkg/vmextension"
	"github.com/pkg/errors"
)

var (
	errStoragePartialCredentials = errors.New("both 'storageAccountName' and 'storageAccountKey' must be specified")
	errCmdTooMany                = errors.New("'commandToExecute' was specified both in public and protected settings; it must be specified only once")
)

// handlerSettings holds the configuration of the extension handler.
type handlerSettings struct {
	publicSettings
	protectedSettings
}

// publicSettings is the type deserialized from public configuration section of
// the extension handler. This should be in sync with publicSettingsSchema.
type publicSettings struct {
	CommandToExecute string   `json:"commandToExecute"`
	FileURLs         []string `json:"fileUris"`
}

// protectedSettings is the type decoded and deserialized from protected
// configuration section. This should be in sync with protectedSettingsSchema.
type protectedSettings struct {
	CommandToExecute   string `json:"commandToExecute"`
	StorageAccountName string `json:"storageAccountName"`
	StorageAccountKey  string `json:"storageAccountKey"`
}

// parseSettings uses specified configFolder (comes from HandlerEnvironment) to
// decrypt and parse the public/protected settings of the extension handler.
func parseSettings(configFolder string) (h handlerSettings, err error) {
	err = vmextension.UnmarshalHandlerSettings(configFolder, &h.publicSettings, &h.protectedSettings)
	return h, errors.Wrapf(err, "error parsing extension configuration")
}

// validate makes logical valiation on the handlerSettings which already passed
// the schema validation.
func (h handlerSettings) validate() error {
	if h.publicSettings.CommandToExecute != "" &&
		h.protectedSettings.CommandToExecute != "" {
		return errCmdTooMany
	}

	if (h.protectedSettings.StorageAccountName != "") !=
		(h.protectedSettings.StorageAccountKey != "") {
		return errStoragePartialCredentials
	}

	return nil
}
