package download

import (
	"os"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
)

// SaveTo uses given downloader to fetch the resource with retries and saves the
// given file. Directory of dst is not created by this function. If a file at
// dst exists, it will be truncated. If a new file is created, mode is used to
// set the permission bits. Written number of bytes are returned on success.
func SaveTo(ctx *log.Context, d []Downloader, dst string, mode os.FileMode) (int64, error) {
	f, err := os.OpenFile(dst, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, mode)
	if err != nil {
		return 0, errors.Wrap(err, "failed to open file for writing")
	}
	defer f.Close()

	n, err := WithRetries(ctx, f, d, ActualSleep)
	if err != nil {
		return n, errors.Wrapf(err, "failed to download response and write to file: %s", dst)
	}

	return n, nil
}
