package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
)

// migrateDataDir moves oldDir to newDir, if oldDir exists by shelling out to
// 'mv -rf'.
func migrateDataDir(ctx log.Logger, oldDir, newDir string) error {
	ok, err := dirExists(oldDir)
	if err != nil {
		return errors.Wrap(err, "could not check old directory")
	}
	if !ok { // no need for migration
		ctx.Log("message", "no old state found to migrate")
		return nil
	}
	ctx.Log("message", "migrating old state")

	var b bytes.Buffer
	var bc = bufferCloser{&b}
	stdout, stderr := bc, bc
	if exitCode, err := Exec(fmt.Sprintf(`mv -f '%s'/* '%s'`, oldDir, newDir), "", stdout, stderr); err != nil {
		output := string(b.Bytes())
		return errors.Wrapf(err, "failed to migrate with mv, exit status: %d, output: %q", exitCode, output)
	}
	if err := os.RemoveAll(oldDir); err != nil {
		return errors.Wrapf(err, "failed to delete old state directory %q", oldDir)
	}
	ctx.Log("message", "migrated old state with mv")
	return nil
}

func dirExists(path string) (bool, error) {
	s, err := os.Stat(path)
	if err == nil {
		return s.IsDir(), nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, errors.Wrap(err, "cannot check if directory exists")
}

type bufferCloser struct {
	*bytes.Buffer
}

func (b bufferCloser) Close() error { return nil }
