package main

import (
	"encoding/json"

	"github.com/Azure/azure-docker-extension/pkg/vmextension"
	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
)

var (
	errStoragePartialCredentials    = errors.New("both 'storageAccountName' and 'storageAccountKey' must be specified")
	errCmdTooMany                   = errors.New("'commandToExecute' was specified both in public and protected settings; it must be specified only once")
	errScriptTooMany                = errors.New("'script' was specified both in public and protected settings; it must be specified only once")
	errCmdAndScript                 = errors.New("'commandToExecute' and 'script' were both specified, but only one is validate at a time")
	errCmdMissing                   = errors.New("'commandToExecute' is not specified")
	errUsingBothKeyAndMsi           = errors.New("'storageAccountName' or 'storageAccountKey' must not be specified with 'managedServiceIdentity'")
	errUsingBothClientIdAndObjectId = errors.New("only one of 'clientId' or 'objectId' must be specified with 'managedServiceIdentity'")
)

// handlerSettings holds the configuration of the extension handler.
type handlerSettings struct {
	publicSettings
	protectedSettings
}

func (s *handlerSettings) commandToExecute() string {
	if s.publicSettings.CommandToExecute != "" {
		return s.publicSettings.CommandToExecute
	}
	return s.protectedSettings.CommandToExecute
}

func (s *handlerSettings) script() string {
	if s.publicSettings.Script != "" {
		return s.publicSettings.Script
	}
	return s.protectedSettings.Script
}

func (s *handlerSettings) fileUrls() []string {
	if len(s.publicSettings.FileURLs) > 0 {
		return s.publicSettings.FileURLs
	}
	return s.protectedSettings.FileURLs
}

// validate makes logical validation on the handlerSettings which already passed
// the schema validation.
func (h handlerSettings) validate() error {
	if h.commandToExecute() == "" && h.script() == "" {
		return errCmdMissing
	}
	if h.publicSettings.CommandToExecute != "" && h.protectedSettings.CommandToExecute != "" {
		return errCmdTooMany
	}

	if h.publicSettings.Script != "" && h.protectedSettings.Script != "" {
		return errScriptTooMany
	}

	if h.commandToExecute() != "" && h.script() != "" {
		return errCmdAndScript
	}

	if (h.protectedSettings.StorageAccountName != "") !=
		(h.protectedSettings.StorageAccountKey != "") {
		return errStoragePartialCredentials
	}

	if (h.protectedSettings.StorageAccountKey != "" || h.protectedSettings.StorageAccountName != "") && !h.protectedSettings.ManagedServiceIdentity.isEmpty() {
		return errUsingBothKeyAndMsi
	}

	if h.protectedSettings.ManagedServiceIdentity.ClientId != "" && h.protectedSettings.ManagedServiceIdentity.ObjectId != "" {
		return errUsingBothClientIdAndObjectId
	}

	return nil
}

// publicSettings is the type deserialized from public configuration section of
// the extension handler. This should be in sync with publicSettingsSchema.
type publicSettings struct {
	SkipDos2Unix     bool     `json:"skipDos2Unix"`
	CommandToExecute string   `json:"commandToExecute"`
	Script           string   `json:"script"`
	FileURLs         []string `json:"fileUris"`
}

// protectedSettings is the type decoded and deserialized from protected
// configuration section. This should be in sync with protectedSettingsSchema.
type protectedSettings struct {
	CommandToExecute       string           `json:"commandToExecute"`
	Script                 string           `json:"script"`
	FileURLs               []string         `json:"fileUris"`
	StorageAccountName     string           `json:"storageAccountName"`
	StorageAccountKey      string           `json:"storageAccountKey"`
	ManagedServiceIdentity clientOrObjectId `json:managedServiceIdentity`
}

type clientOrObjectId struct {
	ObjectId string `json:"objectId"`
	ClientId string `json:"clientId"`
}

func (self *clientOrObjectId) isEmpty() bool {
	return self.ClientId == "" && self.ObjectId == ""
}

// parseAndValidateSettings reads configuration from configFolder, decrypts it,
// runs JSON-schema and logical validation on it and returns it back.
func parseAndValidateSettings(ctx *log.Context, configFolder string) (h handlerSettings, _ error) {
	ctx.Log("event", "reading configuration")
	pubJSON, protJSON, err := readSettings(configFolder)
	if err != nil {
		return h, err
	}
	ctx.Log("event", "read configuration")

	ctx.Log("event", "validating json schema")
	if err := validateSettingsSchema(pubJSON, protJSON); err != nil {
		return h, errors.Wrap(err, "json validation error")
	}
	ctx.Log("event", "json schema valid")

	ctx.Log("event", "parsing configuration json")
	if err := vmextension.UnmarshalHandlerSettings(pubJSON, protJSON, &h.publicSettings, &h.protectedSettings); err != nil {
		return h, errors.Wrap(err, "json parsing error")
	}
	ctx.Log("event", "parsed configuration json")

	ctx.Log("event", "validating configuration logically")
	if err := h.validate(); err != nil {
		return h, errors.Wrap(err, "invalid configuration")
	}
	ctx.Log("event", "validated configuration")
	return h, nil
}

// readSettings uses specified configFolder (comes from HandlerEnvironment) to
// decrypt and parse the public/protected settings of the extension handler into
// JSON objects.
func readSettings(configFolder string) (pubSettingsJSON, protSettingsJSON map[string]interface{}, err error) {
	pubSettingsJSON, protSettingsJSON, err = vmextension.ReadSettings(configFolder)
	err = errors.Wrapf(err, "error reading extension configuration")
	return
}

// validateSettings takes publicSettings and protectedSettings as JSON objects
// and runs JSON schema validation on them.
func validateSettingsSchema(pubSettingsJSON, protSettingsJSON map[string]interface{}) error {
	pubJSON, err := toJSON(pubSettingsJSON)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal public settings into json")
	}
	protJSON, err := toJSON(protSettingsJSON)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal protected settings into json")
	}

	if err := validatePublicSettings(pubJSON); err != nil {
		return err
	}
	if err := validateProtectedSettings(protJSON); err != nil {
		return err
	}
	return nil
}

// toJSON converts given in-memory JSON object representation into a JSON object string.
func toJSON(o map[string]interface{}) (string, error) {
	if o == nil { // instead of JSON 'null' assume empty object '{}'
		return "{}", nil
	}
	b, err := json.Marshal(o)
	return string(b), errors.Wrap(err, "failed to marshal into json")
}
