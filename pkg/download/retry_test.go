package download_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Azure/azure-extension-foundation/msi"
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
	dir, file := CreateTestFile(t)
	defer file.Close()
	defer os.RemoveAll(dir)

	srv := httptest.NewServer(httpbin.GetMux())
	defer srv.Close()

	d := download.NewURLDownload(srv.URL + "/status/200")
	sr := new(sleepRecorder)

	n, err := download.WithRetries(nopLog(), file, []download.Downloader{d}, sr.Sleep)
	require.Nil(t, err, "should not fail")
	require.EqualValues(t, 0, n, "downloaded number of bytes should be zero")
	require.Equal(t, []time.Duration(nil), []time.Duration(*sr), "sleep should not be called")
}

func TestWithRetries_failing_validateNumberOfCalls(t *testing.T) {
	dir, file := CreateTestFile(t)
	defer file.Close()
	defer os.RemoveAll(dir)

	srv := httptest.NewServer(httpbin.GetMux())
	defer srv.Close()

	bd := new(badDownloader)
	sr := new(sleepRecorder)

	n, err := download.WithRetries(nopLog(), file, []download.Downloader{bd}, sr.Sleep)
	require.Contains(t, err.Error(), "expected error", "error is preserved")
	require.EqualValues(t, 0, n, "downloaded number of bytes should be zero")
	require.EqualValues(t, 7, bd.calls, "calls exactly expRetryN times")
}

func TestWithRetries_failingBadStatusCode_validateSleeps(t *testing.T) {
	dir, file := CreateTestFile(t)
	defer file.Close()
	defer os.Remove(dir)

	srv := httptest.NewServer(httpbin.GetMux())
	defer srv.Close()

	d := download.NewURLDownload(srv.URL + "/status/429")
	sr := new(sleepRecorder)

	n, err := download.WithRetries(nopLog(), file, []download.Downloader{d}, sr.Sleep)
	require.Contains(t, err.Error(), "429 Too Many Requests")
	require.Contains(t, err.Error(), "Please verify the machine has network connectivity")
	require.EqualValues(t, 0, n, "downloaded number of bytes should be zero")
	require.Equal(t, sleepSchedule, []time.Duration(*sr))
}

func TestWithRetries_healingServer(t *testing.T) {
	dir, file := CreateTestFile(t)
	defer file.Close()
	defer os.RemoveAll(dir)

	srv := httptest.NewServer(new(healingServer))
	defer srv.Close()

	d := download.NewURLDownload(srv.URL)
	sr := new(sleepRecorder)

	n, err := download.WithRetries(nopLog(), file, []download.Downloader{d}, sr.Sleep)
	require.Nil(t, err, "should eventually succeed")
	require.EqualValues(t, 0, n, "downloaded number of bytes should be zero")
	require.Equal(t, sleepSchedule[:3], []time.Duration(*sr))
}

func TestRetriesWith_SwitchDownloaderOn404(t *testing.T) {
	dir, file := CreateTestFile(t)
	defer file.Close()
	defer os.RemoveAll(dir)

	svr := httptest.NewServer(httpbin.GetMux())
	hSvr := httptest.NewServer(new(healingServer))
	defer svr.Close()

	d404 := mockDownloader{0, svr.URL + "/status/404"}
	d200 := mockDownloader{0, hSvr.URL}

	n, err := download.WithRetries(nopLog(), file, []download.Downloader{&d404, &d200}, func(d time.Duration) { return })
	require.Nil(t, err, "should eventually succeed")
	require.EqualValues(t, 0, n, "downloaded number of bytes should be zero")
	require.Equal(t, d404.timesCalled, 1)
	require.Equal(t, d200.timesCalled, 4)
}

func TestRetriesWith_SwitchDownloaderThenFailWithCorretErrorMessage(t *testing.T) {
	dir, file := CreateTestFile(t)
	defer file.Close()
	defer os.RemoveAll(dir)

	svr := httptest.NewServer(httpbin.GetMux())
	defer svr.Close()

	var mockMsiProvider download.MsiProvider = func() (msi.Msi, error) {
		return msi.Msi{AccessToken: "fakeAccessToken"}, nil
	}

	d404 := mockDownloader{0, svr.URL + "/status/404"}
	msiDownloader403 := download.NewBlobWithMsiDownload(svr.URL+"/status/403", mockMsiProvider)

	n, err := download.WithRetries(nopLog(), file, []download.Downloader{&d404, msiDownloader403}, func(d time.Duration) { return })
	require.NotNil(t, err, "download with retries should fail")
	require.EqualValues(t, 0, n, "downloaded number of bytes should be zero")
	require.Equal(t, d404.timesCalled, 1)
	require.True(t, strings.Contains(err.Error(), download.MsiDownload403ErrorString), "error string doesn't contain the correct message")

	d404 = mockDownloader{0, svr.URL + "/status/404"}
	msiDownloader404 := download.NewBlobWithMsiDownload(svr.URL+"/status/404", mockMsiProvider)

	n, err = download.WithRetries(nopLog(), file, []download.Downloader{&d404, msiDownloader404}, func(d time.Duration) { return })
	require.NotNil(t, err, "download with retries should fail")
	require.EqualValues(t, 0, n, "downloaded number of bytes should be zero")
	require.Equal(t, d404.timesCalled, 1)
	require.True(t, strings.Contains(err.Error(), download.MsiDownload404ErrorString), "error string doesn't contain the correct message")
}

func TestRetriesWith_LargeFileThatTimesOutWhileDownloading(t *testing.T) {
	dir, file := CreateTestFile(t)
	defer file.Close()
	defer os.RemoveAll(dir)

	srv := httptest.NewServer(httpbin.GetMux())
	srv.Config.WriteTimeout = 100 * time.Millisecond
	defer srv.Close()

	size := 1024 * 1024 * 256 // 256 MB
	largeFileDownloader := mockDownloader{0, srv.URL + "/bytes/" + fmt.Sprintf("%d", size)}
	sr := new(sleepRecorder)

	n, err := download.WithRetries(nopLog(), file, []download.Downloader{&largeFileDownloader}, sr.Sleep)
	require.NotNil(t, err, "download with retries should fail because of server timeout")
	require.EqualValues(t, 0, n, "downloaded number of bytes should be zero")

	fi, err := file.Stat()
	require.Nil(t, err)
	require.EqualValues(t, 0, fi.Size())
}

func CreateTestFile(t *testing.T) (string, *os.File) {
	dir, err := ioutil.TempDir("", "")
	require.Nil(t, err)

	path := filepath.Join(dir, "test-file")

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0600)
	require.Nil(t, err)

	return dir, file
}

// Test Utilities:

type mockDownloader struct {
	timesCalled int
	url         string
}

func (self *mockDownloader) GetRequest() (*http.Request, error) {
	self.timesCalled++
	req, err := http.NewRequest("GET", self.url, nil)
	return req, err
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
