package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Azure/custom-script-extension-linux/pkg/errorutil"
	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func Test_handlerSettingsValidate(t *testing.T) {
	// commandToExecute not specified
	require.Equal(t, errCmdMissing, handlerSettings{
		publicSettings{},
		protectedSettings{},
	}.validate().Err)

	// commandToExecute specified twice
	require.Equal(t, errCmdTooMany, handlerSettings{
		publicSettings{CommandToExecute: "foo"},
		protectedSettings{CommandToExecute: "foo"},
	}.validate().Err)

	// script specified twice
	require.Equal(t, errScriptTooMany, handlerSettings{
		publicSettings{Script: "foo"},
		protectedSettings{Script: "foo"},
	}.validate().Err)

	// storageAccount name specified; but not key
	require.Equal(t, errStoragePartialCredentials, handlerSettings{
		protectedSettings: protectedSettings{
			CommandToExecute:   "date",
			StorageAccountName: "foo",
			StorageAccountKey:  ""},
	}.validate().Err)

	// storageAccount key specified; but not name
	require.Equal(t, errStoragePartialCredentials, handlerSettings{
		protectedSettings: protectedSettings{
			CommandToExecute:   "date",
			StorageAccountName: "",
			StorageAccountKey:  "foo"},
	}.validate().Err)
}

func Test_handlerSettingsValidate_commandAndScript(t *testing.T) {
	tests := []struct {
		name     string
		settings handlerSettings
		conflict bool
	}{
		{
			name:     "both public",
			settings: handlerSettings{publicSettings: publicSettings{CommandToExecute: "date", Script: "ZGF0ZQ=="}},
			conflict: true,
		},
		{
			name:     "both protected",
			settings: handlerSettings{protectedSettings: protectedSettings{CommandToExecute: "date", Script: "ZGF0ZQ=="}},
			conflict: true,
		},
		{
			name: "public command and protected script",
			settings: handlerSettings{
				publicSettings:    publicSettings{CommandToExecute: "date"},
				protectedSettings: protectedSettings{Script: "ZGF0ZQ=="},
			},
			conflict: true,
		},
		{
			name: "protected command and public script",
			settings: handlerSettings{
				publicSettings:    publicSettings{Script: "ZGF0ZQ=="},
				protectedSettings: protectedSettings{CommandToExecute: "date"},
			},
			conflict: true,
		},
		{
			name:     "public command only",
			settings: handlerSettings{publicSettings: publicSettings{CommandToExecute: "date"}},
		},
		{
			name:     "protected command only",
			settings: handlerSettings{protectedSettings: protectedSettings{CommandToExecute: "date"}},
		},
		{
			name:     "public script only",
			settings: handlerSettings{publicSettings: publicSettings{Script: "ZGF0ZQ=="}},
		},
		{
			name:     "protected script only",
			settings: handlerSettings{protectedSettings: protectedSettings{Script: "ZGF0ZQ=="}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ewc := tt.settings.validate()
			if !tt.conflict {
				require.Nil(t, ewc)
				return
			}
			require.NotNil(t, ewc)
			require.Equal(t, errCmdAndScript, ewc.Err)
			require.Equal(t, errorutil.CustomerInput_commandToExecuteAndScriptBothSpecified, ewc.ErrorCode)
		})
	}
}

func Test_parseAndValidateSettings_logicalValidation(t *testing.T) {
	tests := []struct {
		name     string
		settings publicSettings
		err      error
		code     int
	}{
		{
			name:     "command and script",
			settings: publicSettings{CommandToExecute: "date", Script: "ZGF0ZQ=="},
			err:      errCmdAndScript,
			code:     errorutil.CustomerInput_commandToExecuteAndScriptBothSpecified,
		},
		{
			name: "neither command nor script",
			err:  errCmdMissing,
			code: errorutil.CustomerInput_commandToExecuteAndScriptNotSpecified,
		},
		{
			name:     "command only",
			settings: publicSettings{CommandToExecute: "date"},
		},
		{
			name:     "script only",
			settings: publicSettings{Script: "ZGF0ZQ=="},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			tt.settings.FileURLs = []string{}
			data, err := json.Marshal(map[string]interface{}{
				"runtimeSettings": []interface{}{
					map[string]interface{}{
						"handlerSettings": map[string]interface{}{"publicSettings": tt.settings},
					},
				},
			})
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(filepath.Join(dir, "0.settings"), data, 0600))

			settings, ewc := parseAndValidateSettings(log.NewContext(log.NewNopLogger()), dir, 0)
			if tt.err == nil {
				require.Nil(t, ewc)
				require.Equal(t, tt.settings, settings.publicSettings)
				return
			}
			require.NotNil(t, ewc)
			require.Equal(t, tt.err, errors.Cause(ewc.Err))
			require.Equal(t, tt.code, ewc.ErrorCode)
			require.Contains(t, ewc.Err.Error(), "invalid configuration")
		})
	}
}

func Test_commandToExecutePrivateIfNotPublic(t *testing.T) {
	testSubject := handlerSettings{
		publicSettings{},
		protectedSettings{CommandToExecute: "bar"},
	}

	require.Equal(t, "bar", testSubject.commandToExecute())
}

func Test_scriptPrivateIfNotPublic(t *testing.T) {
	testSubject := handlerSettings{
		publicSettings{},
		protectedSettings{Script: "bar"},
	}

	require.Equal(t, "bar", testSubject.script())
}

func Test_fileURLsPrivateIfNotPublic(t *testing.T) {
	testSubject := handlerSettings{
		publicSettings{},
		protectedSettings{FileURLs: []string{"bar"}},
	}

	require.Equal(t, []string{"bar"}, testSubject.fileUrls())
}

func Test_skipDos2UnixDefaultsToFalse(t *testing.T) {
	testSubject := handlerSettings{
		publicSettings{CommandToExecute: "/bin/ls"},
		protectedSettings{},
	}

	require.Equal(t, false, testSubject.SkipDos2Unix)
}

func Test_managedIdentityVerification(t *testing.T) {
	err := handlerSettings{publicSettings{}, protectedSettings{
		CommandToExecute: "echo hi",
		FileURLs:         []string{"file1", "file2"},
		ManagedIdentity: &clientOrObjectId{
			ClientId: "31b403aa-c364-4240-a7ff-d85fb6cd7232",
		},
	}}.validate()
	require.Nil(t, err, "validation failed for settings with MSI")

	err = handlerSettings{publicSettings{}, protectedSettings{
		CommandToExecute: "echo hi",
		ManagedIdentity: &clientOrObjectId{
			ObjectId: "31b403aa-c364-4240-a7ff-d85fb6cd7232",
		},
	}}.validate()
	require.Nil(t, err, "validation failed for settings with MSI")

	require.Equal(t, errUsingBothKeyAndMsi,
		handlerSettings{publicSettings{},
			protectedSettings{
				CommandToExecute:   "echo hi",
				StorageAccountName: "name",
				StorageAccountKey:  "key",
				ManagedIdentity: &clientOrObjectId{
					ObjectId: "31b403aa-c364-4240-a7ff-d85fb6cd7232",
				},
			}}.validate().Err, "validation didn't fail for settings with both MSI and storage account")

	require.Equal(t, errUsingBothClientIdAndObjectId,
		handlerSettings{publicSettings{},
			protectedSettings{
				CommandToExecute: "echo hi",
				ManagedIdentity: &clientOrObjectId{
					ObjectId: "31b403aa-c364-4240-a7ff-d85fb6cd7232",
					ClientId: "31b403aa-c364-4240-a7ff-d85fb6cd7232",
				},
			}}.validate().Err, "validation didn't fail for settings with both MSI and storage account")
}

func Test_toJSON_empty(t *testing.T) {
	s, err := toJSON(nil)
	require.Nil(t, err)
	require.Equal(t, "{}", s)
}

func Test_toJSON(t *testing.T) {
	s, err := toJSON(map[string]interface{}{
		"a": 3})
	require.Nil(t, err)
	require.Equal(t, `{"a":3}`, s)
}

func Test_toJSONUmarshallForManagedIdentity(t *testing.T) {
	testString := `{"commandToExecute" : "echo hello", "fileUris":["https://a.com/file.txt", "https://b.com/file2.txt"]}`
	require.NoError(t, validateProtectedSettings(testString), "protected settings should be valid")
	protSettings := new(protectedSettings)
	err := json.Unmarshal([]byte(testString), protSettings)
	require.NoError(t, err, "error while deserializing json")
	require.Nil(t, protSettings.ManagedIdentity, "ProtectedSettings.ManagedIdentity was expected to be nil")
	h := handlerSettings{publicSettings{}, *protSettings}
	require.Nil(t, h.validate(), "settings should be valid")

	testString = `{"commandToExecute" : "echo hello", "fileUris":["https://a.com/file.txt"], "managedIdentity": { }}`
	require.NoError(t, validateProtectedSettings(testString), "protected settings should be valid")
	protSettings = new(protectedSettings)
	err = json.Unmarshal([]byte(testString), protSettings)
	require.NoError(t, err, "error while deserializing json")
	require.NotNil(t, protSettings.ManagedIdentity, "ProtectedSettings.ManagedIdentity was expected to not be nil")
	require.Equal(t, protSettings.ManagedIdentity.ClientId, "")
	require.Equal(t, protSettings.ManagedIdentity.ObjectId, "")
	h = handlerSettings{publicSettings{}, *protSettings}
	require.Nil(t, h.validate(), "settings should be valid")

	testString = `{"commandToExecute" : "echo hello", "fileUris":["https://a.com/file.txt", "https://b.com/file2.txt"], "managedIdentity": { "clientId": "31b403aa-c364-4240-a7ff-d85fb6cd7232"}}`
	require.NoError(t, validateProtectedSettings(testString), "protected settings should be valid")
	protSettings = new(protectedSettings)
	err = json.Unmarshal([]byte(testString), protSettings)
	require.NoError(t, err, "error while deserializing json")
	require.NotNil(t, protSettings.ManagedIdentity, "ProtectedSettings.ManagedIdentity was expected to not be nil")
	require.Equal(t, protSettings.ManagedIdentity.ClientId, "31b403aa-c364-4240-a7ff-d85fb6cd7232")
	require.Equal(t, protSettings.ManagedIdentity.ObjectId, "")
	h = handlerSettings{publicSettings{}, *protSettings}
	require.Nil(t, h.validate(), "settings should be valid")

	testString = `{"commandToExecute" : "echo hello", "fileUris":["https://a.com/file.txt"], "managedIdentity": { "objectId": "31b403aa-c364-4240-a7ff-d85fb6cd7232"}}`
	require.NoError(t, validateProtectedSettings(testString), "protected settings should be valid")
	protSettings = new(protectedSettings)
	err = json.Unmarshal([]byte(testString), protSettings)
	require.NoError(t, err, "error while deserializing json")
	require.NotNil(t, protSettings.ManagedIdentity, "ProtectedSettings.ManagedIdentity was expected to not be nil")
	require.Equal(t, protSettings.ManagedIdentity.ObjectId, "31b403aa-c364-4240-a7ff-d85fb6cd7232")
	require.Equal(t, protSettings.ManagedIdentity.ClientId, "")
	h = handlerSettings{publicSettings{}, *protSettings}
	require.Nil(t, h.validate(), "settings should be valid")

	testString = `{"commandToExecute" : "echo hello", "fileUris":["https://a.com/file.txt", "https://b.com/file2.txt"], "managedIdentity": { "clientId": "31b403aa-c364-4240-a7ff-d85fb6cd7232", "objectId": "41b403aa-c364-4240-a7ff-d85fb6cd7232"}}`
	require.NoError(t, validateProtectedSettings(testString), "protected settings should be valid")
	protSettings = new(protectedSettings)
	err = json.Unmarshal([]byte(testString), protSettings)
	require.NoError(t, err, "error while deserializing json")
	require.NotNil(t, protSettings.ManagedIdentity, "ProtectedSettings.ManagedIdentity was expected to not be nil")
	require.Equal(t, protSettings.ManagedIdentity.ClientId, "31b403aa-c364-4240-a7ff-d85fb6cd7232")
	require.Equal(t, protSettings.ManagedIdentity.ObjectId, "41b403aa-c364-4240-a7ff-d85fb6cd7232")
	h = handlerSettings{publicSettings{}, *protSettings}
	require.Error(t, h.validate(), "settings should be invalid")
}
