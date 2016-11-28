package main

import (
	"io"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
)

// tailFile returns the last max bytes (or the entire file if the file size is
// smaller than max) from the file at path. If the file does not exist, it
// returns a nil slice and no error.
func tailFile(path string, max int64) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil && os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "error opening file")
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving file info")
	}
	size := fi.Size()

	n := max
	if size < n {
		n = size
	}
	_, err = f.Seek(-n, io.SeekEnd)
	if err != nil {
		return nil, errors.Wrapf(err, "error seeking file: offset=%d whence=%v", n, io.SeekEnd)
	}

	b, err := ioutil.ReadAll(io.LimitReader(f, max))
	return b, errors.Wrap(err, "error reading from file")
}
