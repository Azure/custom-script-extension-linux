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
