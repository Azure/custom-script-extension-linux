package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

const (
	// chmod is used to set the mode bits for new seqnum files.
	chmod = os.FileMode(0600)
)

// FindSeqNumConfig gets the laster seq no from config files
func FindSeqNumConfig(path string) (int, error) {
	return FindSeqNum(path, ".settings")
}

// FindSeqNumStatus gets the laster seq no from status files
func FindSeqNumStatus(path string) (int, error) {
	return FindSeqNum(path, ".status")
}

// FindSeqNum finds the file with the highest number under configFolder
// named like 0.settings, 1.settings so on.
func FindSeqNum(path, ext string) (int, error) {
	g, err := filepath.Glob(filepath.Join(path, fmt.Sprintf("*%s", ext)))
	if err != nil {
		return 0, err
	}
	seqs := make([]int, len(g))
	for _, v := range g {
		f := filepath.Base(v)
		i, err := strconv.Atoi(strings.TrimSuffix(f, filepath.Ext(f)))
		if err != nil {
			return 0, fmt.Errorf("Can't parse int from filename: %s", f)
		}
		seqs = append(seqs, i)
	}
	if len(seqs) == 0 {
		return 0, fmt.Errorf("Can't find out seqnum from %s, not enough files", path)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(seqs)))
	return seqs[0], nil
}

// SaveSeqNum replaces the stored sequence number in file, or creates a new file at
// path if it does not exist.
func SaveSeqNum(path string, num int) error {
	b := []byte(fmt.Sprintf("%v", num))
	return errors.Wrap(ioutil.WriteFile(path, b, chmod), "seqnum: failed to write")
}

// IsSmallerThan returns true if the sequence number stored at path is smaller
// than the provided num. If no number is stored, returns true and no
// error.
func IsSmallerThan(path string, num int) (bool, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, errors.Wrap(err, "seqnum: failed to read")
	}
	stored, err := strconv.Atoi(string(b))
	return stored < num, errors.Wrapf(err, "seqnum: cannot parse number %q", b)
}
