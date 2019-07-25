package download

import (
	"encoding/json"
	"github.com/Azure/azure-extension-foundation/msi"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"
)


// README for running this test
// Assign/create an azure VM with system assigned or user assigned identity
// this is the machine that you'll get the msiJson from
// Assign "Storage Blob Data Reader" permissions to managed identity on a blob


var msiJson = `` // place the msi json here e.g.
// {"access_token":<access token>","client_id":"31b403aa-c364-4240-a7ff-d85fb6cd7232","expires_in":"28799",
// "expires_on":"1563607134","ext_expires_in":"28799","not_before":"1563578034","resource":"https://storage.azure.com/",
// "token_type":"Bearer"}
// Linux command to get msi
// curl 'http://169.254.169.254/metadata/identity/oauth2/token?api-version=2018-02-01&resource=https%3A%2F%2Fstorage.azure.com%2F' -H Metadata:true
// Powershell command to get msi
// Invoke-RestMethod -Method "GET" -Headers @{"Metadata"=$true} "http://169.254.169.254/metadata/identity/oauth2/token?api-version=2018-02-01&resource=https%3A%2F%2Fstorage.azure.com%2F" | ConvertTo-Json


var blobUri = "" // set the blob to download here e.g. https://storageaccount.blob.core.windows.net/container/blobname
var stringToLookFor = "" // the string to look for in you blob

type mockMsiProvider struct { //implements MsiProvider
}

func (self *mockMsiProvider) GetMsi() (msi.Msi, error) {
	msi := msi.Msi{}
	err := json.Unmarshal([]byte(msiJson), &msi)
	return msi, err
}

func Test_realDownloadBlobWithMsiToken(t *testing.T) {
	if msiJson == "" || blobUri == "" || stringToLookFor == "" {
		t.Skip()
	}
	downloader := blobWithMsiToken{blobUri, new(mockMsiProvider)}
	_, stream, err := Download(&downloader)
	require.NoError(t, err, "File download failed")
	defer stream.Close()

	bytes, err := ioutil.ReadAll(stream)
	require.NoError(t, err, "saving file stream to memory failed")

	//verify
	require.Contains(t, string(bytes), stringToLookFor)
}
