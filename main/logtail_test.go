package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_tailFile_notFound(t *testing.T) {
	b, err := tailFile("/non/existing/path", 1024)
	require.Nil(t, err)
	require.Len(t, b, 0)
}

func Test_tailFile_openError(t *testing.T) {
	tf := tempFile(t)
	defer os.RemoveAll(tf)

	require.Nil(t, os.Chmod(tf, 0333)) // no read
	_, err := tailFile(tf, 1024)
	require.NotNil(t, err)
	require.Regexp(t, `^error opening file:`, err.Error())
}

func Test_tailFile(t *testing.T) {
	tf := tempFile(t)
	defer os.RemoveAll(tf)

	in := bytes.Repeat([]byte("0123456789"), 10)
	require.Nil(t, ioutil.WriteFile(tf, in, 0666))

	// max=0
	b, err := tailFile(tf, 0)
	require.Nil(t, err)
	require.Len(t, b, 0)

	// max < size
	b, err = tailFile(tf, 5)
	require.Nil(t, err)
	require.EqualValues(t, []byte("56789"), b)

	// max==size
	b, err = tailFile(tf, int64(len(in)))
	require.Nil(t, err)
	require.EqualValues(t, in, b)

	// max>=size
	b, err = tailFile(tf, int64(len(in)+1000))
	require.Nil(t, err)
	require.EqualValues(t, in, b)
}

func tempFile(t *testing.T) string {
	f, err := ioutil.TempFile("", "")
	require.Nil(t, err, "error creating test file")
	defer f.Close()
	return f.Name()
}
