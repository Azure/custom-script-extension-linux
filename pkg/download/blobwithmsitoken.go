package download

import (
	"fmt"
	"github.com/Azure/azure-extension-foundation/httputil"
	"github.com/Azure/azure-extension-foundation/msi"
	"github.com/pkg/errors"
	"net/http"
	url2 "net/url"
	"strings"
)

const (
	xMsVersionHeaderName = "x-ms-version"
	xMsVersionValue      = "2018-03-28"
	storageResourceName  = "https://storage.azure.com/"
)

var azureBlobDomains = map[string]interface{}{ // golang doesn't have builtin hash sets, so this is a workaround for that
	"blob.core.windows.net":       nil,
	"blob.core.chinacloudapi.cn":  nil,
	"blob.core.usgovcloudapi.net": nil,
	"blob.core.couldapi.de":       nil,
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
		return nil, errors.New("MSI token was empty")
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
			return msi, errors.Wrapf(err, "Unable to get managed identity. "+
				"Please make sure that system assigned managed identity is enabled on the VM"+
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
			return msi, errors.Wrapf(err, "Unable to get managed identity with client id %s. "+
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
			return msi, errors.Wrapf(err, "Unable to get managed identity with object id %s. "+
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
	// TODO update this function for sovereign regions
	parsedUrl, err := url2.Parse(url)
	if err != nil {
		return false
	}
	s := strings.Split(parsedUrl.Hostname(), ".")
	if len(s) < 2 {
		return false
	}

	domainName := strings.Join(s[1:], ".")
	_, foundDomain := azureBlobDomains[domainName]
	return foundDomain

}
