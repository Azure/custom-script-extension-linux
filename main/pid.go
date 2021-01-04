package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
)

// GetProcessStartTime returns the start time of the active process if still active
func GetProcessStartTime(pid int) (string, error) {
	pidString := fmt.Sprintf("%d", pid)
	startTime, err := exec.Command("bash", "-c", "ps -o lstart= -p "+pidString).Output()
	if err != nil {
		return "", errors.Wrap(err, "failed to execute bash ps command")
	}
	return string(startTime), nil
}

// SaveCurrentPidAndStartTime stores current process id with start date in file extName.pid
// Example: 325	Tue Dec  8 15:54:04 2020
func SaveCurrentPidAndStartTime(path string) error {
	pid := os.Getpid()
	pidString := fmt.Sprintf("%d", pid)
	startTime, err := GetProcessStartTime(pid)
	if err != nil {
		return errors.Wrap(err, "failed to execute bash ps command")
	}

	b := []byte(fmt.Sprintf("%s\t%s", pidString, startTime))
	return errors.Wrap(ioutil.WriteFile(path, b, chmod), "extName.pid: failed to write")
}

// DeleteCurrentPidAndStartTime delete the file created by SaveCurrentPidAndStartTime
func DeleteCurrentPidAndStartTime(path string) error {
	return errors.Wrap(os.Remove(path), "failed to delete "+path)
}

// ReadPidAndStartTime reads the stored pid and process start time from a file extName.pid
// Returns 0 and "" if path not found
func ReadPidAndStartTime(path string) (int, string, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, "", nil
		}
		return 0, "", errors.Wrap(err, "extName.pid: failed to read:"+path)
	}
	data := strings.Split(string(b), "\t")
	if len(data) != 2 {
		return 0, "", errors.Wrap(err, "unexpected format in extName.pid:"+string(b))
	}

	pid, err := strconv.Atoi(data[0])
	if err != nil {
		return 0, "", errors.Wrap(err, "failed to convert pid:"+data[0])
	}
	return pid, data[1], nil
}

// IsExtensionStillRunning checks if there is active process for the same extension name
func IsExtensionStillRunning(path string) bool {
	// Check if we have a file record for previous process
	previousPid, previousStartTime, err := ReadPidAndStartTime(path)
	if err != nil || previousPid == 0 || previousStartTime == "" {
		return false
	}

	// Try to get previous process start time
	startTime, err := GetProcessStartTime(previousPid)
	if err != nil || startTime == "" {
		return false
	}

	return startTime == previousStartTime
}

// KillPreviousExtension handles the case where a process for the same extension name is still active from previous execution.
// We need to kill it before staring a new one.
func KillPreviousExtension(ctx *log.Context, path string) {
	if IsExtensionStillRunning(pidFilePath) {
		previousPid, _, _ := ReadPidAndStartTime(pidFilePath)
		if ctx != nil {
			ctx.Log("event", "check process", "Active previous execution found. Killing pid ", previousPid)
		}
		syscall.Kill(-previousPid, syscall.SIGKILL) // Negative pid means kill the whole process group
		DeleteCurrentPidAndStartTime(path)
	}
}
