package download

import (
	"fmt"
	"github.com/Azure/azure-extension-foundation/msi"
	"github.com/pkg/errors"
	"net/http"
	url2 "net/url"
	"strings"
)

const (
	xMsVersionHeaderName = "x-ms-version"
	xMsVersionValue      = "2018-03-28"
	azureBlobDomainName  = "blob.core.windows.net"
)

type blobWithMsiToken struct {
	url         string
	msiProvider msi.MsiProvider
}

func (self *blobWithMsiToken) GetRequest() (*http.Request, error) {
	msi, err := self.msiProvider.GetMsi()
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

func NewBlobWithMsiDownload(url string, msiProvider msi.MsiProvider) Downloader {
	return &blobWithMsiToken{url, msiProvider}
}

func IsAzureStorageBlobUri(url string) bool {
	// TODO update this function
	parsedUrl, err := url2.Parse(url)
	if err != nil {
		return false
	}
	return strings.HasSuffix(parsedUrl.Host, azureBlobDomainName)
}
