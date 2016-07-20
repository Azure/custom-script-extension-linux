package preprocess

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHasTextExtension(t *testing.T) {
	require.False(t, hasTextExtension(""), "empty")
	require.True(t, hasTextExtension("a.txt"), "known extension")
	require.True(t, hasTextExtension("a.TXT"), "case insensitive")
	require.False(t, hasTextExtension("a.jpg"), "unknown extension")
	require.True(t, hasTextExtension("a/b/c.py"), "full path")
}

func TestHasShebang(t *testing.T) {
	require.False(t, hasShebang(nil), "empty")

	require.True(t, hasShebang([]byte("#! foo")), "shebang")

	script := "\r\n  \t#!foo"
	require.False(t, hasShebang([]byte(script)), "whitespace")

	bomdata := append(bomSequences[0], []byte("#!booyah")...)
	require.True(t, hasShebang(bomdata), "data=%v", bomdata)
}

func TestIsTextFile_openError(t *testing.T) {
	_, err := IsTextFile("/non/existing/path")
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "failed to open file")
}

func TestIsTextFile(t *testing.T) {
	files := map[string]bool{
		"script_noshebang.py": true,  // file extension, no shebang
		"script_shebang":      true,  // no extension, contains whitespace and shebang
		"utf8_with_bom":       true,  // no extension, has BOM (removing it is ok), contains shebang
		"mslogo.png":          false, // binary
		"whitespace.ws":       false, // a Whitespace script, but can't tell if it is text (https://en.wikipedia.org/wiki/Whitespace_%28programming_language%29)
	}

	for f, exp := range files {
		out, err := IsTextFile(filepath.Join(testDataDir, f))
		require.NoError(t, err)
		require.Equal(t, exp, out, "IsTextFile(%s)", f)
	}
}
