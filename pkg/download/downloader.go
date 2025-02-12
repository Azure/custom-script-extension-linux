package download

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
	"net/url"

	"github.com/Azure/azure-extension-platform/vmextension"
	"github.com/Azure/custom-script-extension-linux/pkg/errorutil"
	"github.com/Azure/custom-script-extension-linux/pkg/urlutil"
	"github.com/go-kit/kit/log"

	"github.com/pkg/errors"
)

// Downloader describes a method to download files.
type Downloader interface {
	// GetRequest returns a new GET request for the resource.
	GetRequest() (*http.Request, error)
}

const (
	MsiDownload404ErrorString = "please ensure that the blob location in the fileUri setting exists, and the specified Managed Identity has read permissions to the storage blob"
	MsiDownload403ErrorString = "please ensure that the specified Managed Identity has read permissions to the storage blob"
	MsiDownloadGenericErrorString = "unable to download the MSI. This may be due to firewall rules or a networking error"
	MsiDownload500ErrorString = "the IMDS service returned a00 upon requesting the MSI"
)

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
func Download(ctx *log.Context, d Downloader) (int, io.ReadCloser, vmextension.ErrorWithClarification) {
	req, err := d.GetRequest()
	if err != nil {
		return -1, nil, vmextension.NewErrorWithClarification(errorutil.FileDownload_genericError, errors.Wrapf(err, "failed to create http request"))
	}
	requestID := req.Header.Get(xMsClientRequestIdHeaderName)
	if len(requestID) > 0 {
		ctx.Log("info", fmt.Sprintf("starting download with client request ID %s", requestID))
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		if ((err.(*url.Error)).Timeout()) {
			err = urlutil.RemoveUrlFromErr(err)
			return -1, nil, vmextension.NewErrorWithClarification(errorutil.FileDownload_exceededTimeout, errors.Wrapf(err, "http request timed out"))
		}
		err = urlutil.RemoveUrlFromErr(err)
		return -1, nil, vmextension.NewErrorWithClarification(errorutil.FileDownload_unknownError, errors.Wrapf(err, "http request failed"))
	}

	if resp.StatusCode == http.StatusOK {
		// We're setting the errorCode to MaxInt because we're only checking whether the internal error is nil
		return resp.StatusCode, resp.Body, vmextension.NewErrorWithClarification(errorutil.NoError, nil)
	}

	errString := ""
	errClarificationCode := 0
	requestId := resp.Header.Get(xMsServiceRequestIdHeaderName)
	switch d.(type) {
	case *blobWithMsiToken:
		switch resp.StatusCode {
		case http.StatusNotFound:
			errString = MsiDownload404ErrorString
			errClarificationCode = errorutil.Msi_notFound
		case http.StatusForbidden:
			errString = MsiDownload403ErrorString
			errClarificationCode = errorutil.Msi_doesNotHaveRightPermissions
		case http.StatusInternalServerError:
			errString = MsiDownload500ErrorString
			errClarificationCode = errorutil.Imds_internalMsiError
		default:
			errString = MsiDownloadGenericErrorString
			errClarificationCode = errorutil.Msi_GenericRetrievalError
		}
		break
	default:
		hostname := req.URL.Host //"string"
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			errString = fmt.Sprintf("CustomScript failed to download the file from %s because access was denied. Please fix the blob permissions and try again, the response code and message returned were: %q",
				hostname,
				resp.Status)
			errClarificationCode = errorutil.FileDownload_accessDenied
		case http.StatusNotFound:
			errString = fmt.Sprintf("CustomScript failed to download the file from %s because it does not exist. Please create the blob and try again, the response code and message returned were: %q",
				hostname,
				resp.Status)
			errClarificationCode = errorutil.FileDownload_doesNotExist

		case http.StatusBadRequest:
			errString = fmt.Sprintf("CustomScript failed to download the file from %s because parts of the request were incorrectly formatted, missing, and/or invalid. The response code and message returned were: %q",
				hostname,
				resp.Status)
			errClarificationCode = errorutil.FileDownload_badRequest

		case http.StatusInternalServerError:
			errString = fmt.Sprintf("CustomScript failed to download the file from %s due to an issue with storage. The response code and message returned were: %q",
				hostname,
				resp.Status)
			errClarificationCode = errorutil.Storage_internalServerError
		default:
			errString = fmt.Sprintf("CustomScript failed to download the file from %s because the server returned a response code and message of %q Please verify the machine has network connectivity.",
				hostname,
				resp.Status)
			errClarificationCode = errorutil.FileDownload_networkingError
		}
	}
	if len(requestId) > 0 {
		errString += fmt.Sprintf(" (Service request ID: %s)", requestId)
	}
	return resp.StatusCode, nil, vmextension.NewErrorWithClarification(errClarificationCode, fmt.Errorf(errString))
}
