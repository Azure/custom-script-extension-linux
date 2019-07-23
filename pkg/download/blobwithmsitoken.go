package download

import (
	"github.com/Azure/azure-extension-foundation/msi"
	"net/http"
	"github.com/pkg/errors"
	"fmt"
)

const (
	xMsVersionHeaderName = "x-ms-version"
	xMsVersionValue = "2018-03-28"
)

type blobWithMsiToken struct{
	url string
	msiProvider msi.MsiProvider
}

func (self *blobWithMsiToken) GetRequest() (*http.Request, error){
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
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", msi.AccessToken))
	request.Header.Set(xMsVersionHeaderName, xMsVersionValue)
	return request, nil
}


func NewBlobWithMsiDownload(url string, msiProvider msi.MsiProvider) Downloader{
	return &blobWithMsiToken{url,msiProvider}
}