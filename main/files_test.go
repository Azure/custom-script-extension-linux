package main

import (
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/ahmetalpbalkan/go-httpbin"
	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/require"
)

func Test_getDownloader_azureBlob(t *testing.T) {
	// error condition
	_, err := getDownloader("http://acct.blob.core.windows.net/", "acct", "key")
	require.NotNil(t, err)

	// valid input
	d, err := getDownloader("http://acct.blob.core.windows.net/container/blob", "acct", "key")
	require.Nil(t, err)
	require.NotNil(t, d)
	require.Equal(t, "download.blobDownload", fmt.Sprintf("%T", d), "got wrong type")
}

func Test_getDownloader_externalUrl(t *testing.T) {
	d, err := getDownloader("http://acct.blob.core.windows.net/", "", "")
	require.Nil(t, err)
	require.NotNil(t, d)
	require.Equal(t, "download.urlDownload", fmt.Sprintf("%T", d), "got wrong type")

	d, err = getDownloader("http://acct.blob.core.windows.net/", "foo", "")
	require.Nil(t, err)
	require.NotNil(t, d)
	require.Equal(t, "download.urlDownload", fmt.Sprintf("%T", d), "got wrong type")

	d, err = getDownloader("http://acct.blob.core.windows.net/", "", "bar")
	require.Nil(t, err)
	require.NotNil(t, d)
	require.Equal(t, "download.urlDownload", fmt.Sprintf("%T", d), "got wrong type")
}

func Test_urlToFileName_badURL(t *testing.T) {
	_, err := urlToFileName("http://192.168.0.%31/")
	require.NotNil(t, err)
	require.Contains(t, err.Error(), `unable to parse URL: "http://192.168.0.%31/"`)
}

func Test_urlToFileName_noFileName(t *testing.T) {
	cases := []string{
		"http://example.com",
		"http://example.com",
		"http://example.com/",
		"http://example.com/#foo",
		"http://example.com?bar",
		"http://example.com/bar/",  // empty after last slash
		"http://example.com/bar//", // empty after last slash
		"http://example.com/?bar",
		"http://example.com/?bar#quux",
	}

	for _, c := range cases {
		_, err := urlToFileName(c)
		require.NotNil(t, err, "not failed: %s", "url=%s", c)
		require.Contains(t, err.Error(), "cannot extract file name from URL", "url=%s", c)
	}
}

func Test_urlToFileName(t *testing.T) {
	cases := []struct{ in, out string }{
		{"http://example.com/1", "1"},
		{"http://example.com/1/2", "2"},
		{"http://example.com/1///2", "2"},
		{"http://example.com/1/2?3=4", "2"},
		{"http://example.com/1/2?3#", "2"},
	}
	for _, c := range cases {
		fn, err := urlToFileName(c.in)
		require.Nil(t, err, "url=%s")
		require.Equal(t, c.out, fn, "url=%s", c)
	}
}

func Test_postProcessFile_fail(t *testing.T) {
	require.NotNil(t, postProcessFile("/non/existing/path"))
}

func Test_postProcessFile(t *testing.T) {
	f, err := ioutil.TempFile("", "")
	require.Nil(t, err)
	defer os.RemoveAll(f.Name())
	_, err = fmt.Fprintf(f, "#!/bin/sh\r\necho 'Hello, world!'\n")
	require.Nil(t, err)
	f.Close()

	require.Nil(t, postProcessFile(f.Name()))

	b, err := ioutil.ReadFile(f.Name())
	require.Nil(t, err)
	require.Equal(t, []byte("#!/bin/sh\necho 'Hello, world!'\n"), b)
}

func Test_downloadAndProcessURL(t *testing.T) {
	srv := httptest.NewServer(httpbin.GetMux())
	defer srv.Close()

	tmpDir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	defer os.RemoveAll(tmpDir)

	cfg := handlerSettings{publicSettings{}, protectedSettings{StorageAccountName: "", StorageAccountKey: ""}}
	err = downloadAndProcessURL(log.NewContext(log.NewNopLogger()), srv.URL+"/bytes/256", tmpDir, &cfg)
	require.Nil(t, err)

	fp := filepath.Join(tmpDir, "256")
	fi, err := os.Stat(fp)
	require.Nil(t, err)
	require.EqualValues(t, 256, fi.Size())
	require.Equal(t, os.FileMode(0500).String(), fi.Mode().String())
}
