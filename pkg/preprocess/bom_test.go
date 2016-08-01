package preprocess

import (
	"bytes"
	"crypto/md5"
	"io/ioutil"
	"testing"

	"path/filepath"

	"github.com/stretchr/testify/require"
)

import "fmt"

const testDataDir = "testdata"

// BOM test files and their checksums in case of loss of BOM bytes
// during moving files around
var bomTestFiles = map[string]string{
	"dos_with_bom.py":                 "c3965bff1e08e3239988c9c5b87b1104",
	"dos_with_bom.sh":                 "ecf2012e5463f40bb06a1104c067b9f5",
	"dos_without_bom.py":              "36c74535cfa0799706524da36d055b7a",
	"dos_without_bom.sh":              "32c36bb357ea9c91e2ef06c29d34c5aa",
	"utf16_big_endian_with_bom.py":    "b72663e152c071eaa64a1f73d4c1e9fc",
	"utf16_big_endian_with_bom.sh":    "a750a43ba017954e1bf624aa89f68ab9",
	"utf16_little_endian_with_bom.py": "10dbd37d2d72c6a27603c7bca6a36d21",
	"utf16_little_endian_with_bom.sh": "1e9701e1e6bdb065372bf0dadced56cf",
	"utf8_with_bom.py":                "8d2e8dcd3fe9133a635d211454fc5ee7",
	"utf8_with_bom.sh":                "05fbe34a82ef8ad16d31686a1fd74d2a",
	"utf8_without_bom.py":             "8d9f6213cfd7cdd5a2c629002256087f",
	"utf8_without_bom.sh":             "ce0abf0a472c7c1ad6f5c4fd2f26d05e",
}

func TestBOM_CheckTestDataIntegrity(t *testing.T) {
	for f, sum := range bomTestFiles {
		fp := filepath.Join(testDataDir, f)
		b, err := ioutil.ReadFile(fp)
		require.Nil(t, err, "error reading %s", fp)

		hash := fmt.Sprintf("%x", md5.Sum(b))
		require.Equal(t, sum, hash, "test file checksum mismatch: %s", fp)
	}
}

func TestRemoveBOM(t *testing.T) {
	for fn := range bomTestFiles {
		fp := filepath.Join(testDataDir, fn)
		b, err := ioutil.ReadFile(fp)
		require.Nil(t, err, "error reading %s", fp)

		n := RemoveBOM(b)
		for _, bv := range bomSequences {
			require.False(t, bytes.HasPrefix(n, bv), "%s still has BOM sequence (%v): %v", fp, bv, n)
		}

		// check if the new file starts with #!
		require.True(t, bytes.HasPrefix(n, []byte("#!")), "%s does not start with shebang: %#v", fn, n)
	}
}

func Test_encodeToUTF8(t *testing.T) {
	// a utf16 string with little endian BOM
	s := []byte("\xff\xfe\x68\x00\x65\x00\x6c\x00\x6c\x00\x6f\x00")
	b := encodeToUTF8(s)
	require.Equal(t, []byte("hello"), b)
}

func Test_encodeToUTF8_returnsSameSliceOnFailure(t *testing.T) {
	// a utf16 string with no BOM
	s := []byte("\x68\x00\x65\x00\x6c\x00\x6c\x00\x6f\x00")
	b := encodeToUTF8(s)
	require.Equal(t, s, b)
}
