package seqnum_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/Azure/custom-script-extension-linux/seqnum"
	"github.com/stretchr/testify/require"
)

func TestSet_nonExistingDir(t *testing.T) {
	err := seqnum.Set("/non/existing/path", 1)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "seqnum: failed to write")
}

func TestSet_writeFail(t *testing.T) {
	fp := testFile(t, 0500) // remove read permissions // remove write permissions
	defer os.RemoveAll(fp)

	err := seqnum.Set(fp, 0)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "seqnum: failed to write")
}

func TestSet_newFile(t *testing.T) {
	fp := testFile(t, 0600)
	require.Nil(t, os.RemoveAll(fp)) // remove test file first

	require.Nil(t, seqnum.Set(fp, 1))

	// validate contents
	b, err := ioutil.ReadFile(fp)
	require.Nil(t, err)
	require.Equal(t, "1", string(b))

	// validate chmod
	fi, err := os.Stat(fp)
	require.Nil(t, err)
	require.EqualValues(t, os.FileMode(0600).String(), fi.Mode().String())
}

func TestSet_truncates(t *testing.T) {
	fp := testFile(t, 0600)
	defer os.RemoveAll(fp)

	require.Nil(t, seqnum.Set(fp, 1))
	require.Nil(t, seqnum.Set(fp, 2))

	b, err := ioutil.ReadFile(fp)
	require.Nil(t, err)
	require.Equal(t, "2", string(b))
}

func TestIsSmallerThan_nonExistingFile(t *testing.T) {
	b, err := seqnum.IsSmallerThan("/non/existing/path", -1)
	require.Nil(t, err)
	require.True(t, b, "non-existing file is always smaller than specified seqnum")
}

func TestIsSmallerThan_readFailure(t *testing.T) {
	fp := testFile(t, 0100) // remove read permissions
	defer os.RemoveAll(fp)

	_, err := seqnum.IsSmallerThan(fp, 0)
	require.NotNil(t, err, "read should have failed")
	require.Contains(t, err.Error(), "seqnum: failed to read")
}

func TestIsSmallerThan_parseError(t *testing.T) {
	fp := testFile(t, 0600)
	defer os.RemoveAll(fp)

	require.Nil(t, ioutil.WriteFile(fp, []byte{'a'}, 0700))

	_, err := seqnum.IsSmallerThan(fp, 0)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "seqnum: cannot parse number \"a\"")
}

func TestIsSmallerThan(t *testing.T) {
	fp := testFile(t, 0600)
	defer os.RemoveAll(fp)

	require.Nil(t, seqnum.Set(fp, 0)) // SET 0

	b, err := seqnum.IsSmallerThan(fp, 0)
	require.Nil(t, err)
	require.False(t, b, "stored=0 ≮ given=0")

	b, err = seqnum.IsSmallerThan(fp, 1)
	require.Nil(t, err)
	require.True(t, b, "stored=0 < given=1")

	require.Nil(t, seqnum.Set(fp, 1)) // SET 1

	b, err = seqnum.IsSmallerThan(fp, 0)
	require.Nil(t, err)
	require.False(t, b, "stored=1 ≮ given=0")

	b, err = seqnum.IsSmallerThan(fp, 1)
	require.Nil(t, err)
	require.False(t, b, "stored=1 ≮ given=1")

	b, err = seqnum.IsSmallerThan(fp, 2)
	require.Nil(t, err)
	require.True(t, b, "stored=1 < given=2")
}

func testFile(t *testing.T, mode os.FileMode) string {
	f, err := ioutil.TempFile("", "")
	require.Nil(t, err, "creating test file failed")
	require.Nil(t, f.Chmod(mode), "chmod test file failed")
	require.Nil(t, f.Close(), "failed to close test file")
	return f.Name()
}
