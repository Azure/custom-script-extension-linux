// Package seqnum provides utilities to persistently store sequence number in a
// file and compare against the stored value to ensure a sequence number is
// processed only once in the extension handler.
package seqnum

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/pkg/errors"
)

const (
	// chmod is used to set the mode bits for new seqnum files.
	chmod = os.FileMode(0600)
)

// Set replaces the stored sequence number in file, or creates a new file at
// path if it does not exist.
func Set(path string, num int) error {
	b := []byte(fmt.Sprintf("%v", num))
	return errors.Wrap(ioutil.WriteFile(path, b, chmod), "seqnum: failed to write")
}

// IsSmallerOrEqualThan returns true if the sequence number stored at path is smaller
// or equal than the provided num. If no number is stored, returns false and no
// error.
func IsSmallerOrEqualThan(path string, num int) (bool, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, errors.Wrap(err, "seqnum: failed to read")
	}
	stored, err := strconv.Atoi(string(b))
	return stored <= num, errors.Wrapf(err, "seqnum: cannot parse number %q", b)
}
