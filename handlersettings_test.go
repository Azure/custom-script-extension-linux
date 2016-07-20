package main

import "testing"
import "github.com/stretchr/testify/require"

func Test_handlerSettingsValidate(t *testing.T) {
	// commandToExecute specified twice
	require.Equal(t, errCmdTooMany, handlerSettings{
		publicSettings{CommandToExecute: "foo"},
		protectedSettings{CommandToExecute: "foo"},
	}.validate())

	// storageAccount name specified; but not key
	require.Equal(t, errStoragePartialCredentials, handlerSettings{
		protectedSettings: protectedSettings{
			StorageAccountName: "foo",
			StorageAccountKey:  ""},
	}.validate())

	// storageAccount key specified; but not name
	require.Equal(t, errStoragePartialCredentials, handlerSettings{
		protectedSettings: protectedSettings{
			StorageAccountName: "",
			StorageAccountKey:  "foo"},
	}.validate())
}
