package main

import (
	"fmt"
	"io"
	"os/exec"
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
