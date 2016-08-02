package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/Azure/azure-docker-extension/pkg/vmextension"
	"github.com/Azure/azure-docker-extension/pkg/vmextension/status"
	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/require"
)

func Test_statusMsg(t *testing.T) {
	require.Equal(t, "Enable succeeded", statusMsg(cmdEnable, status.StatusSuccess, ""))
	require.Equal(t, "Enable succeeded: msg", statusMsg(cmdEnable, status.StatusSuccess, "msg"))

	require.Equal(t, "Enable failed", statusMsg(cmdEnable, status.StatusError, ""))
	require.Equal(t, "Enable failed: msg", statusMsg(cmdEnable, status.StatusError, "msg"))

	require.Equal(t, "Enable in progress", statusMsg(cmdEnable, status.StatusTransitioning, ""))
	require.Equal(t, "Enable in progress: msg", statusMsg(cmdEnable, status.StatusTransitioning, "msg"))
}

func Test_reportStatus_fails(t *testing.T) {
	fakeEnv := vmextension.HandlerEnvironment{}
	fakeEnv.HandlerEnvironment.StatusFolder = "/non-existing/dir/"

	err := reportStatus(log.NewContext(log.NewNopLogger()), fakeEnv, 1, status.StatusSuccess, cmdEnable, "")
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "failed to save handler status")
}

func Test_reportStatus_fileExists(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	defer os.RemoveAll(tmpDir)

	fakeEnv := vmextension.HandlerEnvironment{}
	fakeEnv.HandlerEnvironment.StatusFolder = tmpDir

	require.Nil(t, reportStatus(log.NewContext(log.NewNopLogger()), fakeEnv, 1, status.StatusError, cmdEnable, "FOO ERROR"))

	path := filepath.Join(tmpDir, "1.status")
	b, err := ioutil.ReadFile(path)
	require.Nil(t, err, ".status file exists")
	require.NotEqual(t, 0, len(b), ".status file not empty")
}

func Test_reportStatus_checksIfShouldBeReported(t *testing.T) {
	for _, c := range cmds {
		tmpDir, err := ioutil.TempDir("", "status-"+c.name)
		require.Nil(t, err)
		defer os.RemoveAll(tmpDir)

		fakeEnv := vmextension.HandlerEnvironment{}
		fakeEnv.HandlerEnvironment.StatusFolder = tmpDir
		require.Nil(t, reportStatus(log.NewContext(log.NewNopLogger()), fakeEnv, 2, status.StatusSuccess, c, ""))

		fp := filepath.Join(tmpDir, "2.status")
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
