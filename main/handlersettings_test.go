package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_handlerSettingsValidate(t *testing.T) {

	// scriptUri and script both specified
	require.Equal(t, errSourceNotSpecified, handlerSettings{
		publicSettings{Source: &scriptSource{Script: "foo", ScriptURI: "bar"}},
		protectedSettings{},
	}.validate())

	// 	// commandToExecute not specified
	// 	require.Equal(t, errCmdMissing, handlerSettings{
	// 		publicSettings{},
	// 		protectedSettings{},
	// 	}.validate())

	// 	// commandToExecute specified twice
	// 	require.Equal(t, errCmdTooMany, handlerSettings{
	// 		publicSettings{CommandToExecute: "foo"},
	// 		protectedSettings{CommandToExecute: "foo"},
	// 	}.validate())

	// commandToExecute and script both specified
	// require.Equal(t, errCmdAndScript, handlerSettings{
	// 	publicSettings{CommandToExecute: "foo"},
	// 	protectedSettings{Source: {Script: "foo"}},
	// }.validate())

	// require.Equal(t, errCmdAndScript, handlerSettings{
	// 	publicSettings{Source: {Script: "foo"}},
	// 	protectedSettings{CommandToExecute: "foo"},
	// }.validate())

	// storageAccount name specified; but not key
	// 	require.Equal(t, errStoragePartialCredentials, handlerSettings{
	// 		protectedSettings: protectedSettings{
	// 			CommandToExecute:   "date",
	// 			StorageAccountName: "foo",
	// 			StorageAccountKey:  ""},
	// 	}.validate())

	// 	// storageAccount key specified; but not name
	// 	require.Equal(t, errStoragePartialCredentials, handlerSettings{
	// 		protectedSettings: protectedSettings{
	// 			CommandToExecute:   "date",
	// 			StorageAccountName: "",
	// 			StorageAccountKey:  "foo"},
	// 	}.validate())
	// }

	// func Test_commandToExecutePrivateIfNotPublic(t *testing.T) {
	// 	testSubject := handlerSettings{
	// 		publicSettings{},
	// 		protectedSettings{CommandToExecute: "bar"},
	// 	}

	// 	require.Equal(t, "bar", testSubject.commandToExecute())
	// }

	// func Test_scriptPrivateIfNotPublic(t *testing.T) {
	// 	testSubject := handlerSettings{
	// 		publicSettings{},
	// 		protectedSettings{Script: "bar"},
	// 	}

	// 	require.Equal(t, "bar", testSubject.script())
}

func Test_scriptUri(t *testing.T) {
	testSubject := handlerSettings{
		publicSettings{Source: &scriptSource{ScriptURI: "bar"}},
		protectedSettings{},
	}

	require.Equal(t, "bar", testSubject.Source.ScriptURI)
}

// func Test_syncExecutionDefaultsToFalse(t *testing.T) {
// 	testSubject := handlerSettings{
// 		publicSettings{CommandToExecute: "/bin/ls"},
// 		protectedSettings{},
// 	}

// 	require.Equal(t, false, testSubject.AsyncExecution)
// }

// func Test_managedIdentityVerification(t *testing.T) {
// 	require.NoError(t, handlerSettings{publicSettings{}, protectedSettings{
// 		CommandToExecute: "echo hi",
// 		FileURLs:         []string{"file1", "file2"},
// 		ManagedIdentity: &clientOrObjectId{
// 			ClientId: "31b403aa-c364-4240-a7ff-d85fb6cd7232",
// 		},
// 	}}.validate(), "validation failed for settings with MSI")

// 	require.NoError(t, handlerSettings{publicSettings{}, protectedSettings{
// 		CommandToExecute: "echo hi",
// 		ManagedIdentity: &clientOrObjectId{
// 			ObjectId: "31b403aa-c364-4240-a7ff-d85fb6cd7232",
// 		},
// 	}}.validate(), "validation failed for settings with MSI")

// 	require.Equal(t, errUsingBothKeyAndMsi,
// 		handlerSettings{publicSettings{},
// 			protectedSettings{
// 				CommandToExecute:   "echo hi",
// 				StorageAccountName: "name",
// 				StorageAccountKey:  "key",
// 				ManagedIdentity: &clientOrObjectId{
// 					ObjectId: "31b403aa-c364-4240-a7ff-d85fb6cd7232",
// 				},
// 			}}.validate(), "validation didn't fail for settings with both MSI and storage account")

// 	require.Equal(t, errUsingBothClientIdAndObjectId,
// 		handlerSettings{publicSettings{},
// 			protectedSettings{
// 				CommandToExecute: "echo hi",
// 				ManagedIdentity: &clientOrObjectId{
// 					ObjectId: "31b403aa-c364-4240-a7ff-d85fb6cd7232",
// 					ClientId: "31b403aa-c364-4240-a7ff-d85fb6cd7232",
// 				},
// 			}}.validate(), "validation didn't fail for settings with both MSI and storage account")
// }

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

// func Test_toJSONUmarshallForManagedIdentity(t *testing.T) {
// 	testString := `{"commandToExecute" : "echo hello", "fileUris":["https://a.com/file.txt", "https://b.com/file2.txt"]}`
// 	require.NoError(t, validateProtectedSettings(testString), "protected settings should be valid")
// 	protSettings := new(protectedSettings)
// 	err := json.Unmarshal([]byte(testString), protSettings)
// 	require.NoError(t, err, "error while deserializing json")
// 	require.Nil(t, protSettings.ManagedIdentity, "ProtectedSettings.ManagedIdentity was expected to be nil")
// 	h := handlerSettings{publicSettings{}, *protSettings}
// 	require.NoError(t, h.validate(), "settings should be valid")

// 	testString = `{"commandToExecute" : "echo hello", "fileUris":["https://a.com/file.txt"], "managedIdentity": { }}`
// 	require.NoError(t, validateProtectedSettings(testString), "protected settings should be valid")
// 	protSettings = new(protectedSettings)
// 	err = json.Unmarshal([]byte(testString), protSettings)
// 	require.NoError(t, err, "error while deserializing json")
// 	require.NotNil(t, protSettings.ManagedIdentity, "ProtectedSettings.ManagedIdentity was expected to not be nil")
// 	require.Equal(t, protSettings.ManagedIdentity.ClientId, "")
// 	require.Equal(t, protSettings.ManagedIdentity.ObjectId, "")
// 	h = handlerSettings{publicSettings{}, *protSettings}
// 	require.NoError(t, h.validate(), "settings should be valid")

// 	testString = `{"commandToExecute" : "echo hello", "fileUris":["https://a.com/file.txt", "https://b.com/file2.txt"], "managedIdentity": { "clientId": "31b403aa-c364-4240-a7ff-d85fb6cd7232"}}`
// 	require.NoError(t, validateProtectedSettings(testString), "protected settings should be valid")
// 	protSettings = new(protectedSettings)
// 	err = json.Unmarshal([]byte(testString), protSettings)
// 	require.NoError(t, err, "error while deserializing json")
// 	require.NotNil(t, protSettings.ManagedIdentity, "ProtectedSettings.ManagedIdentity was expected to not be nil")
// 	require.Equal(t, protSettings.ManagedIdentity.ClientId, "31b403aa-c364-4240-a7ff-d85fb6cd7232")
// 	require.Equal(t, protSettings.ManagedIdentity.ObjectId, "")
// 	h = handlerSettings{publicSettings{}, *protSettings}
// 	require.NoError(t, h.validate(), "settings should be valid")

// 	testString = `{"commandToExecute" : "echo hello", "fileUris":["https://a.com/file.txt"], "managedIdentity": { "objectId": "31b403aa-c364-4240-a7ff-d85fb6cd7232"}}`
// 	require.NoError(t, validateProtectedSettings(testString), "protected settings should be valid")
// 	protSettings = new(protectedSettings)
// 	err = json.Unmarshal([]byte(testString), protSettings)
// 	require.NoError(t, err, "error while deserializing json")
// 	require.NotNil(t, protSettings.ManagedIdentity, "ProtectedSettings.ManagedIdentity was expected to not be nil")
// 	require.Equal(t, protSettings.ManagedIdentity.ObjectId, "31b403aa-c364-4240-a7ff-d85fb6cd7232")
// 	require.Equal(t, protSettings.ManagedIdentity.ClientId, "")
// 	h = handlerSettings{publicSettings{}, *protSettings}
// 	require.NoError(t, h.validate(), "settings should be valid")

// 	testString = `{"commandToExecute" : "echo hello", "fileUris":["https://a.com/file.txt", "https://b.com/file2.txt"], "managedIdentity": { "clientId": "31b403aa-c364-4240-a7ff-d85fb6cd7232", "objectId": "41b403aa-c364-4240-a7ff-d85fb6cd7232"}}`
// 	require.NoError(t, validateProtectedSettings(testString), "protected settings should be valid")
// 	protSettings = new(protectedSettings)
// 	err = json.Unmarshal([]byte(testString), protSettings)
// 	require.NoError(t, err, "error while deserializing json")
// 	require.NotNil(t, protSettings.ManagedIdentity, "ProtectedSettings.ManagedIdentity was expected to not be nil")
// 	require.Equal(t, protSettings.ManagedIdentity.ClientId, "31b403aa-c364-4240-a7ff-d85fb6cd7232")
// 	require.Equal(t, protSettings.ManagedIdentity.ObjectId, "41b403aa-c364-4240-a7ff-d85fb6cd7232")
// 	h = handlerSettings{publicSettings{}, *protSettings}
// 	require.Error(t, h.validate(), "settings should be invalid")
// }
