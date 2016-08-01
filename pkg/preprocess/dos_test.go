package preprocess

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDos2Unix(t *testing.T) {
	in := []byte("\r\nLine1\nLine2\r\nLine3\r\n\r\n.")
	out := []byte("\nLine1\nLine2\nLine3\n\n.")

	require.Equal(t, out, Dos2Unix(in))
}

func TestDos2UnixFiles(t *testing.T) {
	testFiles := []string{
		"dos_with_bom.py",
		"dos_with_bom.sh",
		"dos_without_bom.py",
		"dos_without_bom.sh"}

	for _, fn := range testFiles {
		fp := filepath.Join(testDataDir, fn)
		b, err := ioutil.ReadFile(fp)
		require.NoError(t, err, "can't read %s", fp)
		require.True(t, bytes.Contains(b, dosLineEndings), "input test file %s does not contain DOS line endings", fn)

		n := Dos2Unix(b)
		require.False(t, bytes.Contains(n, dosLineEndings), "output of %s still contains DOS line endings: %v ", fn, n)
	}
}
