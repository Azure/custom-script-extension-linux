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
	f, err := ioutil.TempFile("", "")
	require.Nil(t, err)
	f.Close()
	fp := f.Name()
	require.Nil(t, os.Remove(fp)) // delete file on purpose
	defer os.RemoveAll(fp)

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
	f, err := ioutil.TempFile("", "")
	require.Nil(t, err)
	f.Close()
	fp := f.Name()
	defer os.RemoveAll(fp)

	require.Nil(t, seqnum.Set(fp, 1))
	require.Nil(t, seqnum.Set(fp, 2))

	b, err := ioutil.ReadFile(fp)
	require.Nil(t, err)
	require.Equal(t, "2", string(b))
}

func TestIsSmallerOrEqualThan(t *testing.T) {
	fp := testFile(t, 0600)
	defer os.RemoveAll(fp)

	require.Nil(t, seqnum.Set(fp, 0)) // SET 0

	b, err := seqnum.IsSmallerOrEqualThan(fp, 0)
	require.Nil(t, err)
	require.True(t, b, "0≤0")

	b, err = seqnum.IsSmallerOrEqualThan(fp, 1)
	require.Nil(t, err)
	require.True(t, b, "0≤1")

	require.Nil(t, seqnum.Set(fp, 1)) // SET 1

	b, err = seqnum.IsSmallerOrEqualThan(fp, 0)
	require.Nil(t, err)
	require.False(t, b, "1≰0")

	b, err = seqnum.IsSmallerOrEqualThan(fp, 1)
	require.Nil(t, err)
	require.True(t, b, "1≤1")

	b, err = seqnum.IsSmallerOrEqualThan(fp, 2)
	require.Nil(t, err)
	require.True(t, b, "1≤2")
}

func TestIsSmallerOrEqualThan_nonExistingFile(t *testing.T) {
	b, err := seqnum.IsSmallerOrEqualThan("/non/existing/path", 0)
	require.Nil(t, err)
	require.False(t, b, "non-existing file is always smaller than specified seqnum")
}

func TestIsSmallerOrEqualThan_readFailure(t *testing.T) {
	fp := testFile(t, 0100) // remove read permissions
	defer os.RemoveAll(fp)

	_, err := seqnum.IsSmallerOrEqualThan(fp, 0)
	require.NotNil(t, err, "read should have failed")
	require.Contains(t, err.Error(), "seqnum: failed to read")
}

func TestIsSmallerOrEqualThan_parseError(t *testing.T) {
	fp := testFile(t, 0600)
	defer os.RemoveAll(fp)

	require.Nil(t, ioutil.WriteFile(fp, []byte{'a'}, 0700))

	_, err := seqnum.IsSmallerOrEqualThan(fp, 0)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "seqnum: cannot parse number \"a\"")
}

func testFile(t *testing.T, mode os.FileMode) string {
	f, err := ioutil.TempFile("", "")
	require.Nil(t, err, "creating test file failed")
	require.Nil(t, f.Chmod(mode), "chmod test file failed")
	require.Nil(t, f.Close(), "failed to close test file")
	return f.Name()
}
