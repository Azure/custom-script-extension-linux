package download

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/Azure/custom-script-extension-linux/pkg/blobutil"
	"github.com/stretchr/testify/require"
)

func Test_blobDownload_validateInputs(t *testing.T) {
	type sas interface {
		getURL() (string, error)
	}

	_, err := NewBlobDownload("", "", blobutil.AzureBlobRef{}).GetRequest()
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "failed to initialize azure storage client")
	require.Contains(t, err.Error(), "account name required")

	_, err = NewBlobDownload("account", "", blobutil.AzureBlobRef{}).GetRequest()
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "failed to initialize azure storage client")
	require.Contains(t, err.Error(), "account key required")

	_, err = NewBlobDownload("account", "Zm9vCg==", blobutil.AzureBlobRef{}).GetRequest()
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "failed to initialize azure storage client")

	_, err = NewBlobDownload("account", "Zm9vCg==", blobutil.AzureBlobRef{
		StorageBase: storage.DefaultBaseURL,
	}).GetRequest()
	require.Nil(t, err)
}

func Test_blobDownload_getURL(t *testing.T) {
	type sas interface {
		getURL() (string, error)
	}

	d := NewBlobDownload("account", "Zm9vCg==", blobutil.AzureBlobRef{
		StorageBase: "test.core.windows.net",
		Container:   "",
		Blob:        "blob.txt",
	})

	v, ok := d.(blobDownload)
	require.True(t, ok)

	url, err := v.getURL()
	require.Nil(t, err)
	require.Contains(t, url, "https://", "missing https scheme")
	require.Contains(t, url, "/account.blob.test.core.windows.net/", "missing/wrong host")
	require.Contains(t, url, "/$root/", "missing container in url")
	require.Contains(t, url, "/blob.txt", "missing blob name in url")
	for _, v := range []string{"sig", "se", "sr", "sp", "sv"} { // SAS query parameters
		require.Contains(t, url, v+"=", "missing SAS query '%s' in url", v)
	}
}

func Test_blobDownload_fails_badCreds(t *testing.T) {
	d := NewBlobDownload("example", "Zm9vCg==", blobutil.AzureBlobRef{
		StorageBase: storage.DefaultBaseURL,
		Blob:        "fooBlob.txt",
		Container:   "foocontainer",
	})

	status, _, err := Download(d)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "unexpected status code: got=403")
	require.Equal(t, status, http.StatusForbidden)
}

func Test_blobDownload_fails_urlNotFound(t *testing.T) {
	d := NewBlobDownload("accountname", "Zm9vCg==", blobutil.AzureBlobRef{
		StorageBase: ".example.com",
		Blob:        "fooBlob.txt",
		Container:   "foocontainer",
	})

	_, _, err := Download(d)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "http request failed:")
}

func Test_blobDownload_actualBlob(t *testing.T) {
	acct := os.Getenv("AZURE_STORAGE_ACCOUNT")
	key := os.Getenv("AZURE_STORAGE_ACCESS_KEY")
	if acct == "" || key == "" {
		t.Skipf("Skipping: AZURE_STORAGE_ACCOUNT or AZURE_STORAGE_ACCESS_KEY not specified to run this test")
	}
	base := storage.DefaultBaseURL

	// Create a blob first
	cl, err := storage.NewClient(acct, key, base, storage.DefaultAPIVersion, true)
	require.Nil(t, err)
	bs := cl.GetBlobService()

	var (
		n         = 1024 * 64
		name      = "blob.txt"
		container = fmt.Sprintf("custom-script-test-%d", rand.New(rand.NewSource(time.Now().UnixNano())).Int63())
		chunk     = make([]byte, n)
	)
	_, err = bs.DeleteContainerIfExists(container)
	require.Nil(t, err)
	_, err = bs.CreateContainerIfNotExists(container, storage.ContainerAccessTypePrivate)
	require.Nil(t, err)
	defer bs.DeleteContainer(container)
	require.Nil(t, bs.PutAppendBlob(container, name, nil))
	rand.Read(chunk)
	require.Nil(t, bs.AppendBlock(container, name, chunk, nil))

	// Get the blob via downloader
	d := NewBlobDownload(acct, key, blobutil.AzureBlobRef{
		Container:   container,
		Blob:        name,
		StorageBase: base,
	})
	_, body, err := Download(d)
	require.Nil(t, err)
	defer body.Close()
	b, err := ioutil.ReadAll(body)
	require.Nil(t, err)
	require.EqualValues(t, chunk, b, "retrieved body is different body=%d chunk=%d", len(b), len(chunk))
}
