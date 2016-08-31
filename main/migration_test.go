package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/require"
)

func Test_dirExists(t *testing.T) {
	ok, err := dirExists("/non-existing")
	require.Nil(t, err)
	require.False(t, ok)

	d := tempDir(t)
	defer os.RemoveAll(d)
	ok, err = dirExists(d)
	require.Nil(t, err)
	require.True(t, ok)
}

func tempDir(t *testing.T) string {
	d, err := ioutil.TempDir("", "")
	require.Nil(t, err, "failed to create test dir")
	return d
}

func Test_migrateDataDir_noPriorData(t *testing.T) {
	err := migrateDataDir(log.NewNopLogger(), "/non-existing", "/tmp/foo")
	require.Nil(t, err)
}

func Test_migrateDataDir(t *testing.T) {
	d1, err := ioutil.TempDir("", "old")
	defer os.RemoveAll(d1)

	d2, err := ioutil.TempDir("", "new")
	defer os.RemoveAll(d2)
	require.Nil(t, ioutil.WriteFile(filepath.Join(d1, "hello.txt"), []byte("hello"), 0644))

	require.Nil(t, migrateDataDir(log.NewNopLogger(), d1, d2))

	f, err := os.Stat(filepath.Join(d1, "hello.txt"))
	require.NotNil(t, err, "old file should have moved: %#v", f)

	_, err = os.Stat(filepath.Join(d2, "hello.txt"))
	require.Nil(t, err, "new file should be there")

	ok, err := dirExists(d1)
	require.Nil(t, err)
	require.False(t, ok, "old directory must be gone")
}
