package download

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_urlDownload_GetRequest_badURL(t *testing.T) {
	// http.NewRequest will not fail with most URLs, such as
	// containing spaces, relative URLs by design. So testing a
	// misencoded URL here.

	u := "http://[fe80::1%en0]/a.txt"
	d := NewURLDownload(u)
	r, err := d.GetRequest()
	require.NotNil(t, err, u)
	require.Contains(t, err.Error(), "invalid URL", u)
	require.Nil(t, r, u)
}

func Test_urlDownload_GetRequest_goodURL(t *testing.T) {
	u := "http://example.com/a.txt"
	d := NewURLDownload(u)
	r, err := d.GetRequest()
	require.Nil(t, err, u)
	require.NotNil(t, r, u)
}
