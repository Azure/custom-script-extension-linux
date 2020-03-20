package download_test

import (
	"github.com/Azure/azure-extension-foundation/msi"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Azure/custom-script-extension-linux/pkg/download"
	"github.com/ahmetalpbalkan/go-httpbin"
	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/require"
)

var (
	// how much we sleep between retries
	sleepSchedule = []time.Duration{
		3 * time.Second,
		6 * time.Second,
		12 * time.Second,
		24 * time.Second,
		48 * time.Second,
		96 * time.Second}
)

func TestActualSleep_actuallySleeps(t *testing.T) {
	s := time.Now()
	download.ActualSleep(time.Second)
	e := time.Since(s)
	require.InEpsilon(t, 1.0, e.Seconds(), 0.01, "took=%fs", e.Seconds())
}

func TestWithRetries_noRetries(t *testing.T) {
	srv := httptest.NewServer(httpbin.GetMux())
	defer srv.Close()

	d := download.NewURLDownload(srv.URL + "/status/200")

	sr := new(sleepRecorder)
	resp, err := download.WithRetries(nopLog(), []download.Downloader{d}, sr.Sleep)
	require.Nil(t, err, "should not fail")
	require.NotNil(t, resp, "response body exists")
	require.Equal(t, []time.Duration(nil), []time.Duration(*sr), "sleep should not be called")
}

func TestWithRetries_failing_validateNumberOfCalls(t *testing.T) {
	srv := httptest.NewServer(httpbin.GetMux())
	defer srv.Close()

	bd := new(badDownloader)
	_, err := download.WithRetries(nopLog(), []download.Downloader{bd}, new(sleepRecorder).Sleep)
	require.Contains(t, err.Error(), "expected error", "error is preserved")
	require.EqualValues(t, 7, bd.calls, "calls exactly expRetryN times")
}

func TestWithRetries_failingBadStatusCode_validateSleeps(t *testing.T) {
	srv := httptest.NewServer(httpbin.GetMux())
	defer srv.Close()

	d := download.NewURLDownload(srv.URL + "/status/429")

	sr := new(sleepRecorder)
	_, err := download.WithRetries(nopLog(), []download.Downloader{d}, sr.Sleep)
	require.EqualError(t, err, "unexpected status code: actual=429 expected=200")

	require.Equal(t, sleepSchedule, []time.Duration(*sr))
}

func TestWithRetries_healingServer(t *testing.T) {
	srv := httptest.NewServer(new(healingServer))
	defer srv.Close()

	d := download.NewURLDownload(srv.URL)
	sr := new(sleepRecorder)
	resp, err := download.WithRetries(nopLog(), []download.Downloader{d}, sr.Sleep)
	require.Nil(t, err, "should eventually succeed")
	require.NotNil(t, resp, "response body exists")

	require.Equal(t, sleepSchedule[:3], []time.Duration(*sr))
}

func TestRetriesWith_SwitchDownloaderOn404(t *testing.T) {
	svr := httptest.NewServer(httpbin.GetMux())
	hSvr := httptest.NewServer(new(healingServer))
	defer svr.Close()
	d404 := mockDownloader{0, svr.URL + "/status/404"}
	d200 := mockDownloader{0, hSvr.URL}
	resp, err := download.WithRetries(nopLog(), []download.Downloader{&d404, &d200}, func(d time.Duration) { return })
	require.Nil(t, err, "should eventually succeed")
	require.NotNil(t, resp, "response body exists")
	require.Equal(t, d404.timesCalled, 1)
	require.Equal(t, d200.timesCalled, 4)
}

func TestRetriesWith_SwitchDownloaderThenFailWithCorretErrorMessage(t *testing.T) {
	svr := httptest.NewServer(httpbin.GetMux())
	defer svr.Close()
	var mockMsiProvider download.MsiProvider = func() (msi.Msi, error) {
		return msi.Msi{AccessToken:"fakeAccessToken"}, nil
	}

	d404 := mockDownloader{0, svr.URL + "/status/404"}
	msiDownloader403 := download.NewBlobWithMsiDownload(svr.URL + "/status/403", mockMsiProvider)
	resp, err := download.WithRetries(nopLog(), []download.Downloader{&d404, msiDownloader403}, func(d time.Duration) { return })
	require.NotNil(t, err, "download with retries should fail")
	require.Nil(t, resp, "response body should be null for failed download with retries")
	require.Equal(t, d404.timesCalled, 1)
	require.True(t, strings.Contains(err.Error(), download.MsiDownload403ErrorString), "error string doesn't contains the correctMessage")

	d404 = mockDownloader{0, svr.URL + "/status/404"}
	msiDownloader404 := download.NewBlobWithMsiDownload(svr.URL + "/status/404", mockMsiProvider)
	resp, err = download.WithRetries(nopLog(), []download.Downloader{&d404, msiDownloader404}, func(d time.Duration) { return })
	require.NotNil(t, err, "download with retries should fail")
	require.Nil(t, resp, "response body should be null for failed download with retries")
	require.Equal(t, d404.timesCalled, 1)
	require.True(t, strings.Contains(err.Error(), download.MsiDownload404ErrorString), "error string doesn't contains the correctMessage")
}

// Test Utilities:

type mockDownloader struct {
	timesCalled int
	url         string
}

func (self *mockDownloader) GetRequest() (*http.Request, error) {
	self.timesCalled++
	return http.NewRequest("GET", self.url, nil)
}

// sleepRecorder keeps track of the durations of Sleep calls
type sleepRecorder []time.Duration

// Sleep does not actually sleep. It records the duration and returns.
func (s *sleepRecorder) Sleep(d time.Duration) {
	*s = append(*s, d)
}

// healingServer returns HTTP 500 until 4th call, then HTTP 200 afterwards
type healingServer int

func (h *healingServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	*h++
	if *h < 4 {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func nopLog() *log.Context {
	return log.NewContext(log.NewNopLogger())
}
