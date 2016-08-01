package preprocess

import (
	"bytes"
)

var (
	dosLineEndings  = []byte{'\r', '\n'}
	unixLineEndings = []byte{'\n'}
)

// Dos2Unix converts given DOS-line endings to UNIX-line endings
func Dos2Unix(b []byte) []byte {
	return bytes.Replace(b, dosLineEndings, unixLineEndings, -1)
}
