package download_test

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Azure/azure-extension-foundation/msi"
	"github.com/go-kit/kit/log"

	"github.com/Azure/custom-script-extension-linux/pkg/download"
	"github.com/Azure/custom-script-extension-linux/pkg/errorutil"
	"github.com/ahmetalpbalkan/go-httpbin"
	"github.com/stretchr/testify/require"
)

type badDownloader struct{ calls int }

var (
	testctx = log.NewContext(log.NewNopLogger())
)

func (b *badDownloader) GetRequest() (*http.Request, error) {
	b.calls++
	return nil, errors.New("expected error")
}

func TestDownload_wrapsGetRequestError(t *testing.T) {
	_, _, errCode, err := download.Download(testctx, new(badDownloader))
	require.Equal(t, errCode, errorutil.FileDownload_genericError)
	require.NotNil(t, err)
	require.EqualError(t, err, "failed to create http request: expected error")
}

func TestDownload_wrapsHTTPError(t *testing.T) {
	_, _, errCode, err := download.Download(testctx, download.NewURLDownload("bad url"))
	require.Equal(t, errCode, errorutil.FileDownload_unknownError)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "http request failed:")
}

// This test is only to make sure that formatting of error messages for specific codes is correct
func TestDownload_wrapsCommonErrorCodes(t *testing.T) {
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
		respCode, _, errCode, err := download.Download(testctx, download.NewURLDownload(fmt.Sprintf("%s/status/%d", srv.URL, code)))
		require.NotNil(t, err, "not failed for code:%d", code)
		require.Equal(t, code, respCode)
		switch respCode {
		case http.StatusNotFound:
			require.Equal(t, errCode, errorutil.FileDownload_doesNotExist)
			require.Contains(t, err.Error(), "because it does not exist")
		case http.StatusForbidden:
			require.Equal(t, errCode, errorutil.FileDownload_networkingError)
			require.Contains(t, err.Error(), "Please verify the machine has network connectivity")
		case http.StatusInternalServerError:
			require.Equal(t, errCode, errorutil.Storage_internalServerError)
			require.Contains(t, err.Error(), "due to an issue with storage")
		case http.StatusBadRequest:
			require.Equal(t, errCode, errorutil.FileDownload_badRequest)
			require.Contains(t, err.Error(), "because parts of the request were incorrectly formatted, missing, and/or invalid")
		case http.StatusUnauthorized:
			require.Equal(t, errCode, errorutil.FileDownload_accessDenied)
			require.Contains(t, err.Error(), "because access was denied")
		}
	}
}

func TestDownload_statusOKSucceeds(t *testing.T) {
	srv := httptest.NewServer(httpbin.GetMux())
	defer srv.Close()

	_, body, errCode, err := download.Download(testctx, download.NewURLDownload(srv.URL+"/status/200"))
	require.Equal(t, errCode, errorutil.NoError)
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

	returnCode, body, errCode, err := download.Download(testctx, msiDownloader404)
	require.Equal(t, errCode, errorutil.Msi_notFound)
	require.True(t, strings.Contains(err.Error(), download.MsiDownload404ErrorString), "error string doesn't contain the correct message")
	require.Nil(t, body, "body is not nil for failed download")
	require.Equal(t, 404, returnCode, "return code was not 404")

	msiDownloader403 := download.NewBlobWithMsiDownload(srv.URL+"/status/403", mockMsiProvider)
	returnCode, body, errCode, err = download.Download(testctx, msiDownloader403)
	require.Equal(t, errCode, errorutil.Msi_doesNotHaveRightPermissions)
	require.True(t, strings.Contains(err.Error(), download.MsiDownload403ErrorString), "error string doesn't contain the correct message")
	require.Nil(t, body, "body is not nil for failed download")
	require.Equal(t, 403, returnCode, "return code was not 403")

}

func TestDownload_retrievesBody(t *testing.T) {
	srv := httptest.NewServer(httpbin.GetMux())
	defer srv.Close()

	_, body, errCode, err := download.Download(testctx, download.NewURLDownload(srv.URL+"/bytes/65536"))
	require.Equal(t, errCode, errorutil.NoError)
	require.Nil(t, err)
	defer body.Close()
	b, err := io.ReadAll(body)
	require.Nil(t, err)
	require.EqualValues(t, 65536, len(b))
}

func TestDownload_bodyClosesWithoutError(t *testing.T) {
	srv := httptest.NewServer(httpbin.GetMux())
	defer srv.Close()

	_, body, errCode, err := download.Download(testctx, download.NewURLDownload(srv.URL+"/get"))
	require.Equal(t, errCode, errorutil.NoError)
	require.Nil(t, err)
	require.Nil(t, body.Close(), "body should close fine")
}
