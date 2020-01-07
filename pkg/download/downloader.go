package download

import (
	"fmt"
	"github.com/Azure/custom-script-extension-linux/pkg/urlutil"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

// Downloader describes a method to download files.
type Downloader interface {
	// GetRequest returns a new GET request for the resource.
	GetRequest() (*http.Request, error)
}

var (
	// httpClient is the default client to be used in downloading files from
	// Internet. http.Get() uses a client without timeouts (http.DefaultClient)
	// so it is dangerous to use it for downloading files from the Internet.
	httpClient = &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			Proxy:                 http.ProxyFromEnvironment,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 20 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}}
)

// Download retrieves a response body and checks the response status code to see
// if it is 200 OK and then returns the response body. It issues a new request
// every time called. It is caller's responsibility to close the response body.
func Download(d Downloader) (int, io.ReadCloser, error) {
	req, err := d.GetRequest()
	if err != nil {
		return -1, nil, errors.Wrapf(err, "failed to create http request")
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		err = urlutil.RemoveUrlFromErr(err)
		return -1, nil, errors.Wrapf(err, "http request failed")
	}

	if resp.StatusCode == http.StatusOK {
		return resp.StatusCode, resp.Body, nil
	}

	err = fmt.Errorf("unexpected status code: actual=%d expected=%d", resp.StatusCode, http.StatusOK)
	switch d.(type) {
	case *blobWithMsiToken:
		if resp.StatusCode == http.StatusNotFound {
			return resp.StatusCode, nil, errors.Wrapf(err, "please ensure that the blob location in the fileUri setting exists and the specified Managed Identity has read permissions to the storage blob")
		}
	}
	return resp.StatusCode, nil, err
}
