package download

import (
	"fmt"
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
func Download(d Downloader) (io.ReadCloser, error) {
	req, err := d.GetRequest()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create the request")
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "http request failed")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: got=%d expected=%d", resp.StatusCode, http.StatusOK)
	}
	return resp.Body, nil
}
