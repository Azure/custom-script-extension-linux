package download

import (
	"net/http"

	"github.com/google/uuid"
)

const (
	xMsClientRequestIdHeaderName  = "x-ms-client-request-id"
	xMsServiceRequestIdHeaderName = "x-ms-request-id"
)

// urlDownload describes a URL to download.
type urlDownload struct {
	url string
}

// NewURLDownload creates a new  downloader with the provided URL
func NewURLDownload(url string) Downloader {
	return urlDownload{url}
}

// GetRequest returns a new request to download the URL
func (u urlDownload) GetRequest() (*http.Request, error) {
	req, err := http.NewRequest("GET", u.url, nil)
	if req != nil {
		req.Header.Add(xMsClientRequestIdHeaderName, uuid.New().String())
	}
	return req, err
}
