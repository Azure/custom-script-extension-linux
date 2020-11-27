package main

import (
	"encoding/json"
	"fmt"

	"github.com/go-kit/kit/log"
)

// ExecutionState represents script current execution state
type ExecutionState string

const (
	// Unknown state (default value)
	Unknown ExecutionState = "Unknown"

	// Pending script execution
	Pending ExecutionState = "Pending"

	// Running script state
	Running ExecutionState = "Running"

	// Failed to execute script
	Failed = "Failed"

	// Succeeded state when successfully completed the script execution
	Succeeded = "Succeeded"

	// TimedOut state when time timit is reached and scrip has not completed yet
	TimedOut = "TimedOut"

	// Canceled state when customer canceled the script execution
	Canceled = "Canceled"
)

// RunCommandInstanceView reports script execution status
type RunCommandInstanceView struct {
	ExecutionState   ExecutionState `json:"executionState"`
	ExecutionMessage string         `json:"executionMessage"`
	Output           string         `json:"output"`
	Error            string         `json:"error"`
	ExitCode         int            `json:"exitCode"`
	StartTime        string         `json:"startTime"`
	EndTime          string         `json:"endTime"`
}

func (instanceView RunCommandInstanceView) marshal() ([]byte, error) {
	return json.Marshal(instanceView)
}

// reportInstanceView saves operation status to the status file for the extension
// handler with the optional given message, if the given cmd requires reporting
// status.
//
// If an error occurs reporting the status, it will be logged and returned.
func reportInstanceView(ctx *log.Context, hEnv HandlerEnvironment, extName string, seqNum int, t StatusType, c cmd, instanceview *RunCommandInstanceView) error {
	if !c.shouldReportStatus {
		ctx.Log("status", "not reported for operation (by design)")
		return nil
	}
	msg, err := serializeInstanceView(instanceview)
	if err != nil {
		return err
	}
	return reportStatus(ctx, hEnv, extName, seqNum, t, c, msg)
}

func serializeInstanceView(instanceview *RunCommandInstanceView) (string, error) {
	bytes, err := instanceview.marshal()
	if err != nil {
		return "", fmt.Errorf("status: failed to marshal into json: %v", err)
	}
	return string(bytes), err
}
