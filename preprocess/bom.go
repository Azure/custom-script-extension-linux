package preprocess

import (
	"bytes"

	"golang.org/x/text/encoding/unicode"
)

var bomSequences = [][]byte{
	{'\xef', '\xbb', '\xbf'}, // python: codecs.BOM_UTF8
	{'\xff', '\xfe'},         // python: codecs.BOM, codecs.BOM_LE, codecs.BOM_UTF16_LE
	{'\xfe', '\xff'},         // python: codecs.BOM_BE, codecs.BOM_UTF16_BE
}

// RemoveBOM trims the BOM prefix from provided the data and converts
// the text to UTF-8  if it was encoded as UTF-16 with BOM.
func RemoveBOM(b []byte) []byte {
	b = encodeToUTF8(b)
	for _, bs := range bomSequences {
		if bytes.HasPrefix(b, bs) {
			return b[len(bs):]
		}
	}
	return b
}

// encodeUTF8 detects and converts utf16 to utf8 and returns a new slice.
// If the encoding is already correct or given utf16 content is without
// BOM, the provided slice is returned.
func encodeToUTF8(b []byte) []byte {
	var e unicode.Endianness // unused as we'll ExpectBOM

	// if b is not utf16 with bom, decoding will terminate early
	// with unicode.ErrMissingBOM.
	utf16Encoding := unicode.UTF16(e, unicode.ExpectBOM)
	utf8Bytes, err := utf16Encoding.NewDecoder().Bytes(b)
	if err != nil {
		// if we got err == unicode.ErrMissingBOM input is already utf8 but in
		// case of other errors, we return the given slice anyway as this is
		// best-effort.
		return b
	}
	return utf8Bytes // decoded from utf16
}
