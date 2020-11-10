package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/pkg/errors"
)

// Exec runs the given cmd in /bin/sh, saves its stdout/stderr streams to
// the specified files. It waits until the execution terminates.
//
// On error, an exit code may be returned if it is an exit code error.
// Given stdout and stderr will be closed upon returning.
func Exec(cmd, workdir string, stdout, stderr io.WriteCloser) (int, error) {
	defer stdout.Close()
	defer stderr.Close()

	c := exec.Command("/bin/sh", "-c", cmd)
	c.Dir = workdir
	c.Stdout = stdout
	c.Stderr = stderr

	err := c.Run()
	exitErr, ok := err.(*exec.ExitError)
	if ok {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			code := status.ExitStatus()
			return code, fmt.Errorf("command terminated with exit status=%d", code)
		}
	}
	return 0, errors.Wrapf(err, "failed to execute command")
}

// ExecCmdInDir executes the given command in given directory and saves output
// to ./stdout and ./stderr files (truncates files if exists, creates them if not
// with 0600/-rw------- permissions).
//
// Ideally, we execute commands only once per sequence number in run-command-handler,
// and save their output under /var/lib/waagent/<dir>/download/<seqnum>/*.
func ExecCmdInDir(cmd, workdir string) error {
	outFn, errFn := logPaths(workdir)

	outF, err := os.OpenFile(outFn, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return errors.Wrapf(err, "failed to open stdout file")
	}
	errF, err := os.OpenFile(errFn, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return errors.Wrapf(err, "failed to open stderr file")
	}

	_, err = Exec(cmd, workdir, outF, errF)
	return err
}

// logPaths returns stdout and stderr file paths for the specified output
// directory. It does not create the files.
func logPaths(dir string) (stdout string, stderr string) {
	return filepath.Join(dir, "stdout"), filepath.Join(dir, "stderr")
}
