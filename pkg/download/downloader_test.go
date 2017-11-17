package download_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ahmetalpbalkan/go-httpbin"
	"github.com/koralski/run-command-extension-linux/pkg/download"
	"github.com/stretchr/testify/require"
)

type badDownloader struct{ calls int }

func (b *badDownloader) GetRequest() (*http.Request, error) {
	b.calls++
	return nil, errors.New("expected error")
}

func TestDownload_wrapsGetRequestError(t *testing.T) {
	_, err := download.Download(new(badDownloader))
	require.NotNil(t, err)
	require.EqualError(t, err, "failed to create the request: expected error")
}

func TestDownload_wrapsHTTPError(t *testing.T) {
	_, err := download.Download(download.NewURLDownload("bad url"))
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "http request failed:")
}

func TestDownload_badStatusCodeFails(t *testing.T) {
	srv := httptest.NewServer(httpbin.GetMux())
	defer srv.Close()

	for _, code := range []int{
		http.StatusNotFound,
		http.StatusForbidden,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusBadRequest,
		http.StatusUnauthorized,
	} {
		_, err := download.Download(download.NewURLDownload(fmt.Sprintf("%s/status/%d", srv.URL, code)))
		require.NotNil(t, err, "not failed for code:%d", code)
		require.Contains(t, err.Error(), "unexpected status code", "wrong message for code %d", code)
	}
}

func TestDownload_statusOKSucceeds(t *testing.T) {
	srv := httptest.NewServer(httpbin.GetMux())
	defer srv.Close()

	body, err := download.Download(download.NewURLDownload(srv.URL + "/status/200"))
	require.Nil(t, err)
	defer body.Close()
	require.NotNil(t, body)
}

func TestDownload_retrievesBody(t *testing.T) {
	srv := httptest.NewServer(httpbin.GetMux())
	defer srv.Close()

	body, err := download.Download(download.NewURLDownload(srv.URL + "/bytes/65536"))
	require.Nil(t, err)
	defer body.Close()
	b, err := ioutil.ReadAll(body)
	require.Nil(t, err)
	require.EqualValues(t, 65536, len(b))
}

func TestDownload_bodyClosesWithoutError(t *testing.T) {
	srv := httptest.NewServer(httpbin.GetMux())
	defer srv.Close()

	body, err := download.Download(download.NewURLDownload(srv.URL + "/get"))
	require.Nil(t, err)
	require.Nil(t, body.Close(), "body should close fine")
}
