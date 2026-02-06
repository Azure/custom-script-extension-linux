package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	status "github.com/Azure/azure-extension-platform/pkg/status"
	vmextension "github.com/Azure/azure-extension-platform/vmextension"
	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
)

type StatusReport []StatusItem

type StatusItem struct {
	Version      float64 `json:"version"`
	TimestampUTC string  `json:"timestampUTC"`
	Status       Status  `json:"status"`
}

type Type string

const (
	StatusTransitioning Type = "transitioning"
	StatusError         Type = "error"
	StatusSuccess       Type = "success"
)

type Status struct {
	Operation        string           `json:"operation"`
	Status           Type             `json:"status"`
	FormattedMessage FormattedMessage `json:"formattedMessage"`
}
type FormattedMessage struct {
	Lang    string `json:"lang"`
	Message string `json:"message"`
}

func NewStatus(t Type, operation, message string) StatusReport {
	return []StatusItem{
		{
			Version:      1.0,
			TimestampUTC: time.Now().UTC().Format(time.RFC3339),
			Status: Status{
				Operation: operation,
				Status:    t,
				FormattedMessage: FormattedMessage{
					Lang:    "en",
					Message: message},
			},
		},
	}
}

func (r StatusReport) marshal() ([]byte, error) {
	return json.MarshalIndent(r, "", "\t")
}

// Save persists the status message to the specified status folder using the
// sequence number. The operation consists of writing to a temporary file in the
// same folder and moving it to the final destination for atomicity.
func (r StatusReport) Save(statusFolder string, seqNum int) error {
	fn := fmt.Sprintf("%d.status", seqNum)
	path := filepath.Join(statusFolder, fn)
	tmpFile, err := ioutil.TempFile(statusFolder, fn)
	if err != nil {
		return fmt.Errorf("status: failed to create temporary file: %v", err)
	}
	tmpFile.Close()

	b, err := r.marshal()
	if err != nil {
		return fmt.Errorf("status: failed to marshal into json: %v", err)
	}
	if err := ioutil.WriteFile(tmpFile.Name(), b, 0644); err != nil {
		return fmt.Errorf("status: failed to path=%s error=%v", tmpFile.Name(), err)
	}

	if err := os.Rename(tmpFile.Name(), path); err != nil {
		return fmt.Errorf("status: failed to move to path=%s error=%v", path, err)
	}
	return nil
}

// reportStatus saves operation status to the status file for the extension
// handler with the optional given message, if the given cmd requires reporting
// status.
//
// If an error occurs reporting the status, it will be logged and returned.
func reportStatus(ctx *log.Context, hEnv HandlerEnvironment, seqNum int, t Type, c cmd, msg string) error {
	if !c.shouldReportStatus {
		ctx.Log("status", "not reported for operation (by design)")
		return nil
	}
	s := NewStatus(t, c.name, statusMsg(c, t, msg))
	if err := s.Save(hEnv.HandlerEnvironment.StatusFolder, seqNum); err != nil {
		ctx.Log("event", "failed to save handler status", "error", err)
		return errors.Wrap(err, "failed to save handler status")
	}
	return nil
}

// reportErrorStatus saves the error(s) that occurred during the operation
// to the status file for the extension handler with clarification messages and codes,
// if the given cmd requires reporting status.
//
// If an error occurs reporting the status, it will be logged and returned.
func reportErrorStatus(ctx *log.Context, hEnv HandlerEnvironment, seqNum int, t Type, c cmd, ewc *vmextension.ErrorWithClarification) error {
	if !c.shouldReportStatus {
		ctx.Log("status", "not reported for operation (by design)")
		return nil
	}
	var err error
	if ewc == nil {
		s := NewStatus(t, c.name, statusMsg(c, t, ewc.Err.Error()))
		err = s.Save(hEnv.HandlerEnvironment.StatusFolder, seqNum)
	} else {
		s := status.NewError(c.name, status.ErrorClarification{Code: ewc.ErrorCode, Message: ewc.Error()})
		err = s.Save(hEnv.HandlerEnvironment.StatusFolder, uint(seqNum))
	}
	if err != nil {
		ctx.Log("event", "failed to save handler status", "error", err)
		return errors.Wrap(err, "failed to save handler status")
	}
	return nil
}

// readStatus loads current status file in StatusReport
func readStatus(ctx *log.Context, hEnv HandlerEnvironment, seqNum int) (Type, error) {
	fileName := fmt.Sprintf("%d.status", seqNum)
	path := filepath.Join(hEnv.HandlerEnvironment.StatusFolder, fileName)
	buffer, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("Error reading status file %s: %v", path, err)
	}

	var statusReport StatusReport
	if err := json.Unmarshal(buffer, &statusReport); err != nil {
		return "", fmt.Errorf("error parsing json: %v", err)
	}

	if len(statusReport) != 1 {
		return "", fmt.Errorf("wrong statusReport count. expected:1, got:%d", len(statusReport))
	}
	return statusReport[0].Status.Status, nil
}

// statusMsg creates the reported status message based on the provided operation
// type and the given message string.
//
// A message will be generated for empty string. For error status, pass the
// error message.
func statusMsg(c cmd, t Type, msg string) string {
	s := c.name
	switch t {
	case StatusSuccess:
		s += " succeeded"
	case StatusTransitioning:
		s += " in progress"
	case StatusError:
		s += " failed"
	}

	if msg != "" {
		// append the original
		s += ": " + msg
	}
	return s
}
