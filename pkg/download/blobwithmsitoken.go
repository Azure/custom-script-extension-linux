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
	azureBlobDomainName  = ".blob.core.windows.net"
	storageResourceName  = "https://storage.azure.com/"
)

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

func GetMsiProviderForStorageAccountsImplicitly() MsiProvider {
	msiProvider := msi.NewMsiProvider(httputil.NewSecureHttpClient(httputil.DefaultRetryBehavior))
	return func() (msi.Msi, error) { return msiProvider.GetMsiForResource(storageResourceName) }
}

func GetMsiProviderForStorageAccountsWithClientId(clientId string) MsiProvider {
	msiProvider := msi.NewMsiProvider(httputil.NewSecureHttpClient(httputil.DefaultRetryBehavior))
	return func() (msi.Msi, error) { return msiProvider.GetMsiUsingClientId(clientId, storageResourceName) }
}

func GetMsiProviderForStorageAccountsWithObjectId(objectId string) MsiProvider {
	msiProvider := msi.NewMsiProvider(httputil.NewSecureHttpClient(httputil.DefaultRetryBehavior))
	return func() (msi.Msi, error) { return msiProvider.GetMsiUsingClientId(objectId, storageResourceName) }
}

func IsAzureStorageBlobUri(url string) bool {
	// TODO update this function for sovereign regions
	parsedUrl, err := url2.Parse(url)
	if err != nil {
		return false
	}
	return strings.HasSuffix(parsedUrl.Host, azureBlobDomainName)
}
