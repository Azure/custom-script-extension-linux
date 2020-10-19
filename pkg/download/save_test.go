package download_test

import (
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/ahmetalpbalkan/go-httpbin"
	"github.com/koralski/run-command-extension-linux/pkg/download"
	"github.com/stretchr/testify/require"
)

func TestSaveTo_invalidDir(t *testing.T) {
	srv := httptest.NewServer(httpbin.GetMux())
	defer srv.Close()

	d := download.NewURLDownload(srv.URL + "/bytes/65536")

	_, err := download.SaveTo(nopLog(), []download.Downloader{d}, "/nonexistent-dir/dst", 0600)
	require.Contains(t, err.Error(), "failed to open file for writing")
}

func TestSave(t *testing.T) {
	srv := httptest.NewServer(httpbin.GetMux())
	defer srv.Close()

	dir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	defer os.RemoveAll(dir)

	d := download.NewURLDownload(srv.URL + "/bytes/65536")
	path := filepath.Join(dir, "test-file")
	n, err := download.SaveTo(nopLog(), []download.Downloader{d}, path, 0600)
	require.Nil(t, err)
	require.EqualValues(t, 65536, n)

	fi, err := os.Stat(path)
	require.Nil(t, err)
	require.EqualValues(t, 65536, fi.Size())
	require.Equal(t, os.FileMode(0600).String(), fi.Mode().String(), "not chmod'ed")
}

func TestSave_truncates(t *testing.T) {
	srv := httptest.NewServer(httpbin.GetMux())
	defer srv.Close()

	dir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "test-file")
	_, err = download.SaveTo(nopLog(), []download.Downloader{download.NewURLDownload(srv.URL + "/bytes/65536")}, path, 0600)
	require.Nil(t, err)
	_, err = download.SaveTo(nopLog(), []download.Downloader{download.NewURLDownload(srv.URL + "/bytes/128")}, path, 0777)
	require.Nil(t, err)

	fi, err := os.Stat(path)
	require.Nil(t, err)
	require.EqualValues(t, 128, fi.Size())
	require.Equal(t, os.FileMode(0600).String(), fi.Mode().String(), "mode should not be changed")
}

func TestSave_largeFile(t *testing.T) {
	srv := httptest.NewServer(httpbin.GetMux())
	defer srv.Close()

	dir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	defer os.RemoveAll(dir)

	size := 1024 * 1024 * 128 // 128 mb

	path := filepath.Join(dir, "large-file")
	n, err := download.SaveTo(nopLog(), []download.Downloader{download.NewURLDownload(srv.URL + "/bytes/" + fmt.Sprintf("%d", size))}, path, 0600)
	require.Nil(t, err)
	require.EqualValues(t, size, n)

	fi, err := os.Stat(path)
	require.Nil(t, err)
	require.EqualValues(t, size, fi.Size())
}
