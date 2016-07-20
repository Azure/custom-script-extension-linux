package executil

import (
	"bytes"
	"fmt"
	"io"
	lg "log"
	"os"
	osexec "os/exec"
)

var (
	out io.Writer // default output stream for ExecPipe
	log *lg.Logger
)

func init() {
	SetOut(os.Stderr)
}

func SetOut(o io.Writer) {
	out = o
	log = lg.New(out, "[executil] ", lg.LstdFlags)
}

type Fds struct{ Out, Err io.Writer }

// ExecPipe is a convenience method to run programs with
// arguments and return their combined stdout/stderr
// output while printing them both to calling process'
// stdout.
func ExecPipe(program string, args ...string) error {
	return ExecPipeToFds(Fds{out, out}, program, args...)
}

// ExecPipeToFds runs the program with specified args and given
// out/err descriptiors. Non-specified (nil) descriptors will be
// replaced with default out stream.
func ExecPipeToFds(fds Fds, program string, args ...string) error {
	log.Printf("+++ invoke: %s %v", program, args)
	defer log.Printf("--- invoke end")
	cmd := osexec.Command(program, args...)

	// replace nil streams with default
	if fds.Out == nil {
		fds.Out = out
	}
	if fds.Err == nil {
		fds.Err = out
	}

	cmd.Stdout, cmd.Stderr = fds.Out, fds.Err
	err := cmd.Run()
	if err != nil {
		err = fmt.Errorf("executing %s %v failed: %v", program, args, err)
	}
	return err
}

// Exec is a convenience method to run programs with
// arguments and return their combined stdout/stderr
// output as bytes.
func Exec(program string, args ...string) ([]byte, error) {
	var b bytes.Buffer
	cmd := osexec.Command(program, args...)
	cmd.Stdout = &b
	cmd.Stderr = &b
	err := cmd.Run()
	if err != nil {
		err = fmt.Errorf("executing %s failed: %v", program, err)
	}
	return b.Bytes(), err
}

// ExecWithStdin pipes given ReadCloser's contents to the stdin of executed
// command and returns stdout as bytes and redirects stderr of executed command
// stderr of executing process.
func ExecWithStdin(in io.ReadCloser, program string, args ...string) ([]byte, error) {
	var b bytes.Buffer
	cmd := osexec.Command(program, args...)
	cmd.Stdin = in
	cmd.Stdout = &b
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		err = fmt.Errorf("executing %s failed: %v", program, err)
	}
	return b.Bytes(), err
}
