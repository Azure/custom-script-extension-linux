package download

import (
	"fmt"
	"net/http"
	url2 "net/url"
	"strings"

	"github.com/Azure/azure-extension-foundation/httputil"
	"github.com/Azure/azure-extension-foundation/msi"
	"github.com/pkg/errors"
)

const (
	xMsVersionHeaderName = "x-ms-version"
	xMsVersionValue      = "2018-03-28"
	storageResourceName  = "https://storage.azure.com/"
)

var azureBlobDomains = map[string]interface{}{ // golang doesn't have builtin hash sets, so this is a workaround for that
	".blob.core.":       nil,
	".blob.azurestack.": nil,
}

type blobWithMsiToken struct {
	url         string
	msiProvider MsiProvider
}

type MsiProvider func() (msi.Msi, error)

func (self *blobWithMsiToken) GetRequest() (*http.Request, error) {
	msi, err := self.msiProvider()
	if err != nil {
		return nil, err
	}
	if msi.AccessToken == "" {
		return nil, errors.New("MSI token is empty")
	}

	request, err := http.NewRequest(http.MethodGet, self.url, nil)
	if err != nil {
		return nil, err
	}

	if IsAzureStorageBlobUri(self.url) {
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", msi.AccessToken))
		request.Header.Set(xMsVersionHeaderName, xMsVersionValue)
	}
	return request, nil
}

func NewBlobWithMsiDownload(url string, msiProvider MsiProvider) Downloader {
	return &blobWithMsiToken{url, msiProvider}
}

func GetMsiProviderForStorageAccountsImplicitly(blobUri string) MsiProvider {
	msiProvider := msi.NewMsiProvider(httputil.NewSecureHttpClient(httputil.DefaultRetryBehavior))
	return func() (msi.Msi, error) {
		msi, err := msiProvider.GetMsiForResource(GetResourceNameFromBlobUri(blobUri))
		if err != nil {
			return msi, fmt.Errorf("Unable to get managed identity. " +
				"Please make sure that system assigned managed identity is enabled on the VM " +
				"or user assigned identity is added to the system.")
		}
		return msi, nil
	}
}

func GetMsiProviderForStorageAccountsWithClientId(blobUri, clientId string) MsiProvider {
	msiProvider := msi.NewMsiProvider(httputil.NewSecureHttpClient(httputil.DefaultRetryBehavior))
	return func() (msi.Msi, error) {
		msi, err := msiProvider.GetMsiUsingClientId(clientId, GetResourceNameFromBlobUri(blobUri))
		if err != nil {
			return msi, fmt.Errorf("Unable to get managed identity with client id %s. "+
				"Please make sure that the user assigned managed identity is added to the VM ", clientId)
		}
		return msi, nil
	}
}

func GetMsiProviderForStorageAccountsWithObjectId(blobUri, objectId string) MsiProvider {
	msiProvider := msi.NewMsiProvider(httputil.NewSecureHttpClient(httputil.DefaultRetryBehavior))
	return func() (msi.Msi, error) {
		msi, err := msiProvider.GetMsiUsingObjectId(objectId, GetResourceNameFromBlobUri(blobUri))
		if err != nil {
			return msi, fmt.Errorf("Unable to get managed identity with object id %s. "+
				"Please make sure that the user assigned managed identity is added to the VM ", objectId)
		}
		return msi, nil
	}
}

func GetResourceNameFromBlobUri(uri string) string {
	// TODO: update this function as sovereign cloud blob resource strings become available
	// resource string for getting MSI for azure storage is still https://storage.azure.com/ for sovereign regions but it is expected to change
	return storageResourceName
}

func IsAzureStorageBlobUri(url string) bool {
	parsedUrl, err := url2.Parse(url)
	if err != nil {
		return false
	}

	host := parsedUrl.Host

	for validBlobDomain := range azureBlobDomains {
		if strings.Contains(host, validBlobDomain) {
			return true
		}
	}

	return false
}
