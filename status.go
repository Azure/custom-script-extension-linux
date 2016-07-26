package main

import (
	"strconv"

	"github.com/Azure/azure-docker-extension/pkg/vmextension"
	"github.com/Azure/azure-docker-extension/pkg/vmextension/status"
	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
)

// reportStatus saves operation status to the status file for the extension
// handler with the optional given message, if the given cmd requires reporting
// status.
//
// If an error occurs reporting the status, it will be logged and returned.
func reportStatus(ctx *log.Context, hEnv vmextension.HandlerEnvironment, t status.Type, c cmd, msg string) error {
	if !c.shouldReportStatus {
		ctx.Log("status", "not reported for operation (by design)")
		return nil
	}
	s := status.NewStatus(t, c.name, statusMsg(c, t, msg))
	seq, _ := strconv.Atoi(hEnv.SeqNo)
	if err := s.Save(hEnv.HandlerEnvironment.StatusFolder, seq); err != nil {
		ctx.Log("event", "failed to save handler status", "error", err)
		return errors.Wrap(err, "failed to save handler status")
	}
	return nil
}

// statusMsg creates the reported status message based on the provided operation
// type and the given message string.
//
// A message will be generated for empty string. For error status, pass the
// error message.
func statusMsg(c cmd, t status.Type, msg string) string {
	s := c.name
	switch t {
	case status.StatusSuccess:
		s += " succeeded"
	case status.StatusTransitioning:
		s += " in progress"
	case status.StatusError:
		s += " failed"
	}

	if msg != "" {
		// append the original
		s += ": " + msg
	}
	return s
}
