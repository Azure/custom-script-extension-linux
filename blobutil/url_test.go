package blobutil_test

import (
	"testing"

	"github.com/Azure/custom-script-extension-linux/blobutil"
	"github.com/stretchr/testify/require"
)

func TestParseBlobURL_badURL(t *testing.T) {
	for _, u := range []string{
		"",
		" http://a.b.c/d/e",
		"http://a.b.c/d/e  ",
	} {
		_, err := blobutil.ParseBlobURL(u)
		require.NotNil(t, err, "invalid URL: %q", u)
	}
}

func TestParseBlobURL_badHost(t *testing.T) {
	for _, v := range []string{ // bad URL formats are already captured by json-schema
		"http://1/c/blob.txt",
		"http://1.blob/c/blob.txt",
		"http://1.blob./c/blob.txt",
		"http://1..blob./c/blob.txt",
		"http://1.notblob.a/c/blob.txt",
	} {
		_, err := blobutil.ParseBlobURL(v)
		require.NotNil(t, err, "bad host: %q", v)
	}
}

func TestParseBlobURL_goodHost_parsedStorageBase(t *testing.T) {
	for _, v := range []struct{ in, out string }{
		{"http://1.blob.local/c/blob.txt", "local"},
		{"http://1.BLOB.local/c/blob.txt", "local"}, // upper
		{"http://1.blob.storage.local:5000/c/blob.txt", "storage.local:5000"},
		{"http://1.blob.core.windows.net/c/blob.txt", "core.windows.net"},
	} {
		o, err := blobutil.ParseBlobURL(v.in)
		require.Nil(t, err, "url: %q", v)
		require.Equal(t, v.out, o.StorageBase, "url: %q", v)
	}
}

func TestParseBlobURL_unsupportedScheme(t *testing.T) {
	for _, v := range []string{
		"",                   // no scheme
		"container/blob.txt", // no scheme
		"1.blob.local/c/blob.txt",
		"tcp://1.blob.storage.local:5000/c/blob.txt",
		"ftp://1.blob.storage.local:5000/c/blob.txt",
		"httpq://1.blob.storage.local:5000/c/blob.txt",
	} {
		_, err := blobutil.ParseBlobURL(v)
		require.NotNil(t, err, "url: %q", v)
		require.Contains(t, err.Error(), "unsupported scheme in URL", "url: %q")
	}
}

func TestParseBlobURL_scheme(t *testing.T) {
	for _, v := range []struct{ in, out string }{
		{"http://1.blob.storage.local:5000/c/blob.txt", "http"},   // explicitly downgraded
		{"https://1.blob.storage.local:5000/c/blob.txt", "https"}, // preserve https
	} {
		o, err := blobutil.ParseBlobURL(v.in)
		require.Nil(t, err, "url: %q", v)
		require.Equal(t, v.out, o.Scheme, "url: %q", v)
	}
}

func TestParseBlobURL_containerMissing(t *testing.T) {
	_, err := blobutil.ParseBlobURL("http://acct.blob.core.windows.net/")
	require.NotNil(t, err, "missing blob/container name")
	require.Contains(t, err.Error(), "cannot extract Azure container/blob name")
}

func TestParseBlobURL_containerName(t *testing.T) {
	for _, v := range []struct{ in, expected string }{
		{"http://acct.blob.core.windows.net/a.txt", "$root"},
		{"http://acct.blob.core.windows.net/$root/a.txt", "$root"},
		{"http://acct.blob.core.windows.net/a/b.txt", "a"},
		{"http://acct.blob.core.windows.net/a/b.txt?c=d", "a"},
		{"http://acct.blob.core.windows.net/a/b/c.txt", "a"},
		{"http://acct.blob.core.windows.net/a/b/c/d", "a"},
		{"http://acct.blob.core.windows.net/a/b//c/d", "a"},
	} {
		o, err := blobutil.ParseBlobURL(v.in)
		require.Nil(t, err, "url: %q", v)
		require.Equal(t, v.expected, o.Container, "url: %q", v.in)
	}
}

func TestParseBlobURL_blobName(t *testing.T) {
	for _, v := range []struct{ in, expected string }{
		{"http://acct.blob.core.windows.net/a.txt", "a.txt"},
		{"http://acct.blob.core.windows.net/$root/a.txt", "a.txt"},
		{"http://acct.blob.core.windows.net/a/b.txt", "b.txt"},
		{"http://acct.blob.core.windows.net/a/b/c.txt", "b/c.txt"},
		{"http://acct.blob.core.windows.net/a/b.txt?c=d", "b.txt"},
		{"http://acct.blob.core.windows.net/a/b/c/d", "b/c/d"},
		{"http://acct.blob.core.windows.net/a/b//c/d", "b//c/d"},
	} {
		o, err := blobutil.ParseBlobURL(v.in)
		require.Nil(t, err, "url: %q", v)
		require.Equal(t, v.expected, o.Blob, "url: %q", v.in)
	}
}
