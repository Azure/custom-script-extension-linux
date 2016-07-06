package download

import (
	"net/http"
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
	return http.NewRequest("GET", u.url, nil)
}
