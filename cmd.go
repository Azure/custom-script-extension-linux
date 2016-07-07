package main

import (
	"github.com/google/shlex"
	"github.com/pkg/errors"
)

// parseCmd is a wrapper function for the underlying shell lexer implementation,
// it splits a commands string into argv slice based upon shell quoting and escaping
// rules.
func parseCmd(cmd string) ([]string, error) {
	v, err := shlex.Split(cmd)
	return v, errors.Wrap(err, "shlex: failed to parse command")
}
