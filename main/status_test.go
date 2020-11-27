package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/require"
)

func Test_reportStatus_fails(t *testing.T) {
	fakeEnv := HandlerEnvironment{}
	fakeEnv.HandlerEnvironment.StatusFolder = "/non-existing/dir/"

	err := reportStatus(log.NewContext(log.NewNopLogger()), fakeEnv, "", 1, StatusSuccess, cmdEnable, "")
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "failed to save handler status")
}

func Test_reportStatus_fileExists(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	defer os.RemoveAll(tmpDir)

	extName := "first"
	fakeEnv := HandlerEnvironment{}
	fakeEnv.HandlerEnvironment.StatusFolder = tmpDir

	require.Nil(t, reportStatus(log.NewContext(log.NewNopLogger()), fakeEnv, extName, 1, StatusError, cmdEnable, "FOO ERROR"))

	path := filepath.Join(tmpDir, "first.1.status")
	b, err := ioutil.ReadFile(path)
	require.Nil(t, err, ".status file exists")
	require.NotEqual(t, 0, len(b), ".status file not empty")
}

func Test_reportStatus_checksIfShouldBeReported(t *testing.T) {
	for _, c := range cmds {
		tmpDir, err := ioutil.TempDir("", "status-"+c.name)
		require.Nil(t, err)
		defer os.RemoveAll(tmpDir)

		extName := "first"
		fakeEnv := HandlerEnvironment{}
		fakeEnv.HandlerEnvironment.StatusFolder = tmpDir
		require.Nil(t, reportStatus(log.NewContext(log.NewNopLogger()), fakeEnv, extName, 2, StatusSuccess, c, ""))

		fp := filepath.Join(tmpDir, "first.2.status")
		_, err = os.Stat(fp) // check if the .status file is there
		if c.shouldReportStatus && err != nil {
			t.Fatalf("cmd=%q should have reported status file=%q err=%v", c.name, fp, err)
		}
		if !c.shouldReportStatus {
			if err == nil {
				t.Fatalf("cmd=%q should not have reported status file. file=%q", c.name, fp)
			} else if !os.IsNotExist(err) {
				t.Fatalf("cmd=%q some other error occurred. file=%q err=%q", c.name, fp, err)
			}
		}
	}
}
