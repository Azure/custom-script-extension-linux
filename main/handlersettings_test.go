package main

import "testing"
import "github.com/stretchr/testify/require"

func Test_handlerSettingsValidate(t *testing.T) {
	// commandToExecute not specified
	require.Equal(t, errCmdMissing, handlerSettings{
		publicSettings{},
		protectedSettings{},
	}.validate())

	// commandToExecute specified twice
	require.Equal(t, errCmdTooMany, handlerSettings{
		publicSettings{CommandToExecute: "foo"},
		protectedSettings{CommandToExecute: "foo"},
	}.validate())

	// script specified twice
	require.Equal(t, errScriptTooMany, handlerSettings{
		publicSettings{Script: "foo"},
		protectedSettings{Script: "foo"},
	}.validate())

	// commandToExecute and script both specified
	require.Equal(t, errCmdAndScript, handlerSettings{
		publicSettings{CommandToExecute: "foo"},
		protectedSettings{Script: "foo"},
	}.validate())

	require.Equal(t, errCmdAndScript, handlerSettings{
		publicSettings{Script: "foo"},
		protectedSettings{CommandToExecute: "foo"},
	}.validate())

	// storageAccount name specified; but not key
	require.Equal(t, errStoragePartialCredentials, handlerSettings{
		protectedSettings: protectedSettings{
			CommandToExecute:   "date",
			StorageAccountName: "foo",
			StorageAccountKey:  ""},
	}.validate())

	// storageAccount key specified; but not name
	require.Equal(t, errStoragePartialCredentials, handlerSettings{
		protectedSettings: protectedSettings{
			CommandToExecute:   "date",
			StorageAccountName: "",
			StorageAccountKey:  "foo"},
	}.validate())
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

func Test_managedSystemIdentityVerification(t *testing.T) {
	require.NoError(t, handlerSettings{publicSettings{}, protectedSettings{
		CommandToExecute: "echo hi",
		FileURLs:         []string{"file1", "file2"},
		ManagedServiceIdentity: &clientOrObjectId{
			ClientId: "31b403aa-c364-4240-a7ff-d85fb6cd7232",
		},
	}}.validate(), "validation failed for settings with MSI")

	require.NoError(t, handlerSettings{publicSettings{}, protectedSettings{
		CommandToExecute: "echo hi",
		ManagedServiceIdentity: &clientOrObjectId{
			ObjectId: "31b403aa-c364-4240-a7ff-d85fb6cd7232",
		},
	}}.validate(), "validation failed for settings with MSI")

	require.Equal(t, errUsingBothKeyAndMsi,
		handlerSettings{publicSettings{},
			protectedSettings{
				CommandToExecute:   "echo hi",
				StorageAccountName: "name",
				StorageAccountKey:  "key",
				ManagedServiceIdentity: &clientOrObjectId{
					ObjectId: "31b403aa-c364-4240-a7ff-d85fb6cd7232",
				},
			}}.validate(), "validation didn't fail for settings with both MSI and storage account")

	require.Equal(t, errUsingBothClientIdAndObjectId,
		handlerSettings{publicSettings{},
			protectedSettings{
				CommandToExecute:   "echo hi",
				ManagedServiceIdentity: &clientOrObjectId{
					ObjectId: "31b403aa-c364-4240-a7ff-d85fb6cd7232",
					ClientId: "31b403aa-c364-4240-a7ff-d85fb6cd7232",
				},
			}}.validate(), "validation didn't fail for settings with both MSI and storage account")
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
