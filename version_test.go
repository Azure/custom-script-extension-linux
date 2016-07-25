package main

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVersionString_clean(t *testing.T) {
	defer resetStrings()

	Version = "1.0.0"
	State = ""
	GitCommit = "03669cef"
	require.Equal(t, "v1.0.0/git@03669cef", VersionString())
}

func TestVersionString_dirty(t *testing.T) {
	defer resetStrings()

	Version = "1.0.0"
	State = "dirty"
	GitCommit = "03669cef"
	require.Equal(t, "v1.0.0/git@03669cef-dirty", VersionString())
}

func TestDetailedVersionString(t *testing.T) {
	defer resetStrings()

	Version = "1.0.0"
	State = ""
	GitCommit = "03669cef"
	BuildDate = "DATE"
	goVersion := runtime.Version()
	require.Equal(t, "v1.0.0 git:03669cef-clean build:DATE "+goVersion, DetailedVersionString())

	State = "dirty"
	require.Equal(t, "v1.0.0 git:03669cef-dirty build:DATE "+goVersion, DetailedVersionString())
}

func resetStrings() { Version, GitCommit, BuildDate, State = "", "", "", "" }
