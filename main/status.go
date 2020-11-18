package main

import (
	"github.com/Azure/azure-docker-extension/pkg/vmextension"
	"github.com/Azure/azure-extension-platform/pkg/status"
	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
)

// reportStatus saves operation status to the status file for the extension
// handler with the optional given message, if the given cmd requires reporting
// status.
//
// If an error occurs reporting the status, it will be logged and returned.
func reportStatus(ctx *log.Context, hEnv vmextension.HandlerEnvironment, seqNum uint, t status.StatusType, c cmd, msg string) error {
	if !c.shouldReportStatus {
		ctx.Log("status", "not reported for operation (by design)")
		return nil
	}
	s := status.New(t, c.name, status.StatusMsg(c.name, t, msg))
	if err := s.Save(hEnv.HandlerEnvironment.StatusFolder, seqNum); err != nil {
		ctx.Log("event", "failed to save handler status", "error", err)
		return errors.Wrap(err, "failed to save handler status")
	}
	return nil
}
