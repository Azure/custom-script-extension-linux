package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_SaveAndReadPid(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	defer os.RemoveAll(tmpDir)

	// Verify Save pid operation
	path := filepath.Join(tmpDir, "extName.pid")
	require.Nil(t, SaveCurrentPidAndStartTime(path))

	pid, date, err := ReadPidAndStartTime(path)
	require.Nil(t, err, "ReadPidAndStartTime failed")

	expectedPid := os.Getpid()
	pidString := fmt.Sprintf("%d", pid)
	expectedStartTime, err := exec.Command("bash", "-c", "ps -o lstart= -p "+pidString).Output()
	require.Equal(t, expectedPid, pid)
	require.Equal(t, string(expectedStartTime), date)
}

func Test_IsExtensionStillRunning(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	defer os.RemoveAll(tmpDir)

	path := filepath.Join(tmpDir, "extName.pid")
	require.Nil(t, SaveCurrentPidAndStartTime(path))

	running := IsExtensionStillRunning(path)
	require.Equal(t, true, running)

	running = IsExtensionStillRunning(path + "notexist")
	require.Equal(t, false, running)
}

func Test_GetProcessStartTime(t *testing.T) {
	startTime, err := GetProcessStartTime(os.Getpid())
	require.NotEmpty(t, startTime)
	require.Nil(t, err)

	startTime, err = GetProcessStartTime(123456)
	require.Empty(t, startTime)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "failed to execute bash ps command")
}
