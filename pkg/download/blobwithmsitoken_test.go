package download

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/Azure/azure-extension-foundation/msi"
	"github.com/stretchr/testify/require"
)

// README
// to run this test, assign/create an azure VM with system assigned or user assigned identity
// this is the machine that you'll get the msiJson from
// assign "Storage Blob Data Reader" permissions to managed identity on a blob

var msiJson = `` // place the msi json here e.g.
// {"access_token":<access token>","client_id":"31b403aa-c364-4240-a7ff-d85fb6cd7232","expires_in":"28799",
// "expires_on":"1563607134","ext_expires_in":"28799","not_before":"1563578034","resource":"https://storage.azure.com/",
// "token_type":"Bearer"}

// Linux command to get msi
// curl 'http://169.254.169.254/metadata/identity/oauth2/token?api-version=2018-02-01&resource=https%3A%2F%2Fstorage.azure.com%2F' -H Metadata:true
// curl 'http://169.254.169.254/metadata/identity/oauth2/token?api-version=2018-02-01&resource=https%3A%2F%2Fstorage.azure.com%2F&client_id=<client_id>' -H Metadata:true
// curl 'http://169.254.169.254/metadata/identity/oauth2/token?api-version=2018-02-01&resource=https%3A%2F%2Fstorage.azure.com%2F&object_id=<object_id>' -H Metadata:true

// Powershell command to get msi
// Invoke-RestMethod -Method "GET" -Headers @{"Metadata"=$true} "http://169.254.169.254/metadata/identity/oauth2/token?api-version=2018-02-01&resource=https%3A%2F%2Fstorage.azure.com%2F" | ConvertTo-Json
// Invoke-RestMethod -Method "GET" -Headers @{"Metadata"=$true} "http://169.254.169.254/metadata/identity/oauth2/token?api-version=2018-02-01&resource=https%3A%2F%2Fstorage.azure.com%2F&client_id=<client_id>" | ConvertTo-Json
// Invoke-RestMethod -Method "GET" -Headers @{"Metadata"=$true} "http://169.254.169.254/metadata/identity/oauth2/token?api-version=2018-02-01&resource=https%3A%2F%2Fstorage.azure.com%2F&object_id=<object_id>" | ConvertTo-Json

// the first command gets the system managed identity, or the user assigned identity if the VM has only one user assigned identity and no system assigned identity
// the second command gets user assigned identity with its client id
// the third command gets user assigned identity with its object id

var blobUri = ""         // set the blob to download here e.g. https://storageaccount.blob.core.windows.net/container/blobname
var stringToLookFor = "" // the string to look for in you blob

func Test_realDownloadBlobWithMsiToken(t *testing.T) {
	if msiJson == "" || blobUri == "" || stringToLookFor == "" {
		t.Skip()
	}
	downloader := blobWithMsiToken{blobUri, func() (msi.Msi, error) {
		msi := msi.Msi{}
		err := json.Unmarshal([]byte(msiJson), &msi)
		return msi, err
	}}
	_, stream, err := Download(testctx, &downloader)
	require.NoError(t, err, "File download failed")
	defer stream.Close()

	bytes, err := ioutil.ReadAll(stream)
	require.NoError(t, err, "saving file stream to memory failed")
	require.Contains(t, string(bytes), stringToLookFor)
}

func Test_realDownloadBlobWithMsiToken404(t *testing.T) {
	if msiJson == "" || blobUri == "" || stringToLookFor == "" {
		t.Skip()
	}
	var badBlobUri = blobUri[0 : len(blobUri)-1]
	downloader := blobWithMsiToken{badBlobUri, func() (msi.Msi, error) {
		msi := msi.Msi{}
		err := json.Unmarshal([]byte(msiJson), &msi)
		return msi, err
	}}
	code, _, err := Download(testctx, &downloader)
	require.NotNil(t, err, "File download succeeded but was not supposed to")
	require.Equal(t, http.StatusNotFound, code)
	require.Contains(t, err.Error(), MsiDownload404ErrorString)
	require.Contains(t, err.Error(), "Service request ID:") // should have a service request ID since downloading from Azure Storage
}

func Test_isAzureStorageBlobUri(t *testing.T) {
	require.True(t, IsAzureStorageBlobUri("https://a.blob.core.windows.net/container/blobname"))
	require.True(t, IsAzureStorageBlobUri("http://mystorageaccountcn.blob.core.chinacloudapi.cn"))
	require.True(t, IsAzureStorageBlobUri("https://blackforestsa.blob.core.couldapi.de/c/b/x"))
	require.True(t, IsAzureStorageBlobUri("https://another.blob.core.future.store/c/b/x"))
	require.False(t, IsAzureStorageBlobUri("https://github.com/Azure-Samples/storage-blobs-go-quickstart/blob/master/README.md"))
	require.False(t, IsAzureStorageBlobUri("http://github.com/Azure-Samples/storage-blobs-go-quickstart/blob/master/README.md"))
	require.False(t, IsAzureStorageBlobUri("file:\\\\C:\\scripts\\Script.ps1"))
}
