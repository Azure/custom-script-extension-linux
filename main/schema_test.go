package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// func TestValidatePublicSettings_fieldHasWrongType(t *testing.T) {
// 	err := validatePublicSettings(`{"commandToExecute": ["foo"]}`)
// 	require.NotNil(t, err)
// 	require.Contains(t, err.Error(), "Invalid type. Expected: string, given: array")
// }

// func TestValidatePublicSettings_unrecognizedField(t *testing.T) {
// 	err := validatePublicSettings(`{"commandToExecute": "date", "alien":0}`)
// 	require.NotNil(t, err)
// 	require.Contains(t, err.Error(), "Additional property alien is not allowed")
// }

// func TestValidatePublicSettings_fileUris(t *testing.T) {
// 	// empty
// 	err := validatePublicSettings(`{"commandToExecute": "date", "fileUris":[]}`)
// 	require.Nil(t, err)

// 	// not a URL
// 	err = validatePublicSettings(`{"commandToExecute": "date", "fileUris":["a"]}`)
// 	require.NotNil(t, err)
// 	require.Contains(t, err.Error(), "Does not match format 'uri'")

// 	// mixed types
// 	err = validatePublicSettings(`{"commandToExecute": "date", "fileUris":["https://a.b/c.txt?d=e&f=g", 0]}`)
// 	require.NotNil(t, err)
// 	require.Contains(t, err.Error(), "Expected: string, given: integer")
// }

// func TestValidatePublicSettings_timestampSupported(t *testing.T) {
// 	require.Nil(t, validatePublicSettings(`{"commandToExecute": "date", "timestamp": 1}`))
// }

func TestValidateProtectedSettings_empty(t *testing.T) {
	require.Nil(t, validateProtectedSettings(""), "empty string")
	require.Nil(t, validateProtectedSettings("{}"), "empty string")
}

func TestValidateProtectedSettings_unrecognizedField(t *testing.T) {
	err := validateProtectedSettings(`{"alien":0}`)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Additional property alien is not allowed")
}

// func TestValidateProtectedSettings_commandToExecute(t *testing.T) {
// 	// Invalid type
// 	err := validateProtectedSettings(`{"commandToExecute": false}`)
// 	require.NotNil(t, err)
// 	require.Contains(t, err.Error(), "Expected: string, given: boolean")

// 	// Valid
// 	require.Nil(t, validateProtectedSettings(`{"commandToExecute":"date"}`))
// }

// func TestValidateProtectedSettings_fileUris(t *testing.T) {
// 	// empty
// 	err := validateProtectedSettings(`{"commandToExecute": "date", "fileUris":[]}`)
// 	require.Nil(t, err)

// 	// not a URL
// 	err = validateProtectedSettings(`{"commandToExecute": "date", "fileUris":["a"]}`)
// 	require.NotNil(t, err)
// 	require.Contains(t, err.Error(), "Does not match format 'uri'")

// 	// mixed types
// 	err = validateProtectedSettings(`{"commandToExecute": "date", "fileUris":["https://a.b/c.txt?d=e&f=g", 0]}`)
// 	require.NotNil(t, err)
// 	require.Contains(t, err.Error(), "Expected: string, given: integer")
// }

func TestValidateProtectedSettings_script(t *testing.T) {
	// Invalid type
	err := validateProtectedSettings(`{"runAsPassword": false}`)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Expected: string, given: boolean")

	// Valid
	require.Nil(t, validateProtectedSettings(`{"runAsPassword":"samplepassword"}`))
}

func TestValidatePublicSettings_skipDos2Unix(t *testing.T) {
	// Invalid type
	err := validatePublicSettings(`{"asyncExecution": "not-a-bool"}`)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Expected: boolean, given: string")

	// Valid
	require.Nil(t, validatePublicSettings(`{"asyncExecution":true}`))
}

// func TestValidateProtectedSettings_storageAccountName(t *testing.T) {
// 	chkPatternMismatch := func(e error, reason string) {
// 		require.NotNil(t, e, reason)
// 		require.Contains(t, e.Error(), "storageAccountName: Does not match pattern", reason)
// 	}

// 	// Specified but empty
// 	chkPatternMismatch(validateProtectedSettings(`{"storageAccountName": ""}`), "empty")

// 	// Too short
// 	chkPatternMismatch(validateProtectedSettings(`{"storageAccountName": "aa"}`), "too short")

// 	// Too long
// 	chkPatternMismatch(validateProtectedSettings(`{"storageAccountName": "1234567890123456789012345"}`), "too long")

// 	// invalid chars
// 	chkPatternMismatch(validateProtectedSettings(`{"storageAccountName": "foo-bar"}`), "invalid char")
// 	chkPatternMismatch(validateProtectedSettings(`{"storageAccountName": " foobar"}`), "invalid char")
// 	chkPatternMismatch(validateProtectedSettings(`{"storageAccountName": "foobar "}`), "invalid char")
// 	chkPatternMismatch(validateProtectedSettings(`{"storageAccountName": "foo.bar"}`), "invalid char")

// 	// ok storage account name
// 	require.Nil(t, validateProtectedSettings(`{"storageAccountName": "0foobarquuxyz12345"}`))
// }

// func TestValidateProtectedSettings_storageAccountKey(t *testing.T) {
// 	chkPatternMismatch := func(e error, reason string) {
// 		require.NotNil(t, e, reason)
// 		require.Contains(t, e.Error(), "storageAccountKey: Does not match pattern", reason)
// 	}

// 	// empty
// 	chkPatternMismatch(validateProtectedSettings(`{"storageAccountKey": ""}`), "empty")

// 	// bad string
// 	chkPatternMismatch(validateProtectedSettings(`{"storageAccountKey": "NotABase64Really!"}`), "not b64")

// 	// for a base64 string ending with '==', removing one of the '=' is not valid b64, the schema validation should catch that
// 	chkPatternMismatch(validateProtectedSettings(`{"storageAccountKey": "OllwYfXmC0mSMhWg4x+lUdLg6Eoa/d44+PxPTXBaadO5l87L4JzgkyyVvQr8r60WIzG2X8r6LLxkhNBQaHa3XQ="}`), "bad b64")

// 	// spacing
// 	chkPatternMismatch(validateProtectedSettings(`{"storageAccountKey": "OllwYfXmC0mSMhWg4x+lUdLg6Eoa/d44+PxPTXBaadO5l87L4JzgkyyVvQr8r60WIzG2X8r6LLxkhNBQaHa3XQ== "}`), "whitespace")
// 	chkPatternMismatch(validateProtectedSettings(`{"storageAccountKey": " OllwYfXmC0mSMhWg4x+lUdLg6Eoa/d44+PxPTXBaadO5l87L4JzgkyyVvQr8r60WIzG2X8r6LLxkhNBQaHa3XQ=="}`), "whitespace")

// 	// ok
// 	require.Nil(t, validateProtectedSettings(`{"storageAccountKey": "OllwYfXmC0mSMhWg4x+lUdLg6Eoa/d44+PxPTXBaadO5l87L4JzgkyyVvQr8r60WIzG2X8r6LLxkhNBQaHa3XQ=="}`), "ok")
// 	require.Nil(t, validateProtectedSettings(`{"storageAccountKey": "A+hMRrsZQ6COPXTYX/EiKiF2HVtfhCfLDo3Dkc3ekKoX3jA58zXVG2QRe/C1+zdEFSrVX6FZsKyivsSlnwmWOw=="}`), "ok")
// 	require.Nil(t, validateProtectedSettings(`{"storageAccountKey": "/yGnx6KyxQ8Pjzk0QXeY+66Du0BeTWaCt83la59w72hu/81e6TzskXXvL/IlO3q6g0k0kJrR9MYQNi+cNR3SXA=="}`), "ok")
// }

// func TestValidateProtectedSettings_managedServiceIdentity(t *testing.T) {
// 	require.NoError(t, validateProtectedSettings(`{"managedIdentity": { "clientId": "31b403aa-c364-4240-a7ff-d85fb6cd7232"}}`),
// 		"couldn't parse msi proprety with lowercase guid")
// 	require.NoError(t, validateProtectedSettings(`{"managedIdentity": { "objectId": "31B403AA-C364-4240-A7FF-D85FB6CD7232"}}`),
// 		"couldn't parse msi property with uppercase guid")
// 	require.NoError(t, validateProtectedSettings(`{"managedIdentity": { }}`),
// 		"couldn't parse msi property without clientId or objectId")

// 	require.Error(t, validateProtectedSettings(`{"managedIdentity": { "clientId": "notaguid"}}`),
// 		"guid validation succeded when expected to fail")
// }
