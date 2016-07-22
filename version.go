package main

import (
	"fmt"
	"runtime"
)

// These fields are set by the compiler using the linker flags upon build via
// Makefile.
var (
	Version   string
	GitCommit string
	BuildDate string
	State     string
)

// VersionString builds a compact version string in format:
// vVERSION/git@GitCommit[-State].
func VersionString() string {
	s := fmt.Sprintf("v%s/git@%s", Version, GitCommit)
	if State != "" {
		s += "-" + State
	}
	return s
}

// DetailedVersionString returns a detailed version string including version
// number, git commit, build date, source tree state and the go runtime version.
func DetailedVersionString() string {
	state := State
	if state == "" {
		state = "clean"
	}

	// e.g. v2.2.0 git:03669cef-clean build:2016-07-22T16:22:26.556103000+00:00 go:go1.6.2
	return fmt.Sprintf("v%s git:%s-%s build:%s %s", Version, GitCommit, state, BuildDate, runtime.Version())
}
