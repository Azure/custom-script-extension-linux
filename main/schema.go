package main

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

// Refer to http://json-schema.org/ on how to use JSON Schemas.

const (
	publicSettingsSchema = `{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "title": "Custom Script - Public Settings",
  "type": "object",
  "properties": {
    "commandToExecute": {
      "description": "Command to be executed",
      "type": "string"
    },
    "script": {
      "description": "Script to be executed",
      "type": "string"
    },
    "skipDos2Unix": {
      "description": "Skip DOS2UNIX and BOM removal for download files and script",
      "type": "boolean"
    },
    "fileUris": {
      "description": "List of files to be downloaded",
      "type": "array",
      "items": {
        "type": "string",
        "format": "uri"
      }
    },
    "timestamp": {
      "description": "An integer, intended to trigger re-execution of the script when changed",
      "type": "integer"
    }
  },
  "additionalProperties": false
}`

	protectedSettingsSchema = `{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "title": "Custom Script - Protected Settings",
  "type": "object",
  "properties": {
    "commandToExecute": {
      "description": "Command to be executed",
      "type": "string"
    },
    "fileUris": {
      "description": "List of files to be downloaded",
      "type": "array",
      "items": {
        "type": "string",
        "format": "uri"
      }
    },
    "script": {
      "description": "Script to be executed",
      "type": "string"
    },
    "storageAccountName": {
      "description": "Name of the Azure Storage Account (3-24 characters of lowercase letters or digits)",
      "type": "string",
      "pattern": "^[a-z0-9]{3,24}$"
    },
    "storageAccountKey": {
      "description": "Key for the Azure Storage Account (a base64 encoded string)",
      "type": "string",
      "pattern": "^(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=|[A-Za-z0-9+/]{4})$"
    },
    "managedIdentity": {
      "description": "Setting to use Managed Service Identity to try to download fileUri from azure blob",
      "type": "object",
      "properties": {
        "objectId": {
          "description": "Object id that identifies the user created managed identity",
          "type": "string",
          "pattern": "^(?:[0-9A-Fa-f]{8}[-][0-9A-Fa-f]{4}[-][0-9A-Fa-f]{4}[-][0-9A-Fa-f]{4}[-][0-9A-Fa-f]{12})$"
        },
        "clientId": {
          "description": "Client id that identifies the user created managed identity",
          "type": "string",
          "pattern": "^(?:[0-9A-Fa-f]{8}[-][0-9A-Fa-f]{4}[-][0-9A-Fa-f]{4}[-][0-9A-Fa-f]{4}[-][0-9A-Fa-f]{12})$"
        }
      }
    }
  },
  "additionalProperties": false
}`
)

// validateObjectJSON validates the specified json with schemaJSON.
// If json is empty string, it will be converted into an empty JSON object
// before being validated.
func validateObjectJSON(schema *gojsonschema.Schema, json string) error {
	if json == "" {
		json = "{}"
	}

	doc := gojsonschema.NewStringLoader(json)
	res, err := schema.Validate(doc)
	if err != nil {
		return err
	}
	if !res.Valid() {
		for _, err := range res.Errors() {
			// return with the first error
			return fmt.Errorf("%s", err)
		}
	}
	return nil
}

func validateSettingsObject(settingsType, schemaJSON, docJSON string) error {
	schema, err := gojsonschema.NewSchema(gojsonschema.NewStringLoader(schemaJSON))
	if err != nil {
		return errors.Wrapf(err, "failed to load %s settings schema", settingsType)
	}
	if err := validateObjectJSON(schema, docJSON); err != nil {
		return errors.Wrapf(err, "invalid %s settings JSON", settingsType)
	}
	return nil
}

func validatePublicSettings(json string) error {
	return validateSettingsObject("public", publicSettingsSchema, json)
}

func validateProtectedSettings(json string) error {
	return validateSettingsObject("protected", protectedSettingsSchema, json)
}
