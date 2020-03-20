package download_test

import (
	"errors"
	"fmt"
	"github.com/Azure/azure-extension-foundation/msi"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Azure/custom-script-extension-linux/pkg/download"
	"github.com/ahmetalpbalkan/go-httpbin"
	"github.com/stretchr/testify/require"
)

type badDownloader struct{ calls int }

func (b *badDownloader) GetRequest() (*http.Request, error) {
	b.calls++
	return nil, errors.New("expected error")
}

func TestDownload_wrapsGetRequestError(t *testing.T) {
	_, _, err := download.Download(new(badDownloader))
	require.NotNil(t, err)
	require.EqualError(t, err, "failed to create http request: expected error")
}

func TestDownload_wrapsHTTPError(t *testing.T) {
	_, _, err := download.Download(download.NewURLDownload("bad url"))
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
		_, _, err := download.Download(download.NewURLDownload(fmt.Sprintf("%s/status/%d", srv.URL, code)))
		require.NotNil(t, err, "not failed for code:%d", code)
		require.Contains(t, err.Error(), "unexpected status code", "wrong message for code %d", code)
	}
}

func TestDownload_statusOKSucceeds(t *testing.T) {
	srv := httptest.NewServer(httpbin.GetMux())
	defer srv.Close()

	_, body, err := download.Download(download.NewURLDownload(srv.URL + "/status/200"))
	require.Nil(t, err)
	defer body.Close()
	require.NotNil(t, body)
}

func TestDowload_msiDownloaderErrorMessage(t *testing.T) {
	var mockMsiProvider download.MsiProvider = func() (msi.Msi, error) {
		return msi.Msi{AccessToken: "fakeAccessToken"}, nil
	}
	srv := httptest.NewServer(httpbin.GetMux())
	defer srv.Close()

	msiDownloader404 := download.NewBlobWithMsiDownload(srv.URL+"/status/404", mockMsiProvider)

	returnCode, body, err := download.Download(msiDownloader404)
	require.True(t, strings.Contains(err.Error(), download.MsiDownload404ErrorString), "error string doesn't contains the correctMessage")
	require.Nil(t, body, "body is not nil for failed download")
	require.Equal(t, 404, returnCode, "return code was not 404")

	msiDownloader403 := download.NewBlobWithMsiDownload(srv.URL+"/status/403", mockMsiProvider)
	returnCode, body, err = download.Download(msiDownloader403)
	require.True(t, strings.Contains(err.Error(), download.MsiDownload403ErrorString), "error string doesn't contains the correctMessage")
	require.Nil(t, body, "body is not nil for failed download")
	require.Equal(t, 403, returnCode, "return code was not 403")

}

func TestDownload_retrievesBody(t *testing.T) {
	srv := httptest.NewServer(httpbin.GetMux())
	defer srv.Close()

	_, body, err := download.Download(download.NewURLDownload(srv.URL + "/bytes/65536"))
	require.Nil(t, err)
	defer body.Close()
	b, err := ioutil.ReadAll(body)
	require.Nil(t, err)
	require.EqualValues(t, 65536, len(b))
}

func TestDownload_bodyClosesWithoutError(t *testing.T) {
	srv := httptest.NewServer(httpbin.GetMux())
	defer srv.Close()

	_, body, err := download.Download(download.NewURLDownload(srv.URL + "/get"))
	require.Nil(t, err)
	require.Nil(t, body.Close(), "body should close fine")
}
