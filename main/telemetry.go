package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/pkg/errors"
)

const (
	telemetryEventsPath = "/var/lib/waagent/events"
)

type telemetryParameterString struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type telemetryParameterLong struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
}

type telemetryParameterBool struct {
	Name  string `json:"name"`
	Value bool   `json:"value"`
}

type telemetryEvent struct {
	EventID    int           `json:"eventId"`
	ProviderID string        `json:"providerId"`
	Parameters []interface{} `json:"parameters"`
}

type telemetryEventWriter struct {
	fh *os.File
}

func (w *telemetryEventWriter) Write(bs []byte) (n int, err error) {
	fn := getTelemetryFileName()
	temp := fn + ".tmp"

	fh, err := os.OpenFile(temp, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0400)
	if err != nil {
		return 0, errors.Wrap(err, "failed to open telemetry file")
	}

	w.fh = fh
	n, err = w.fh.Write(bs)
	err = os.Rename(temp, fn)

	return
}

func (w *telemetryEventWriter) Close() (err error) {
	if w.fh != nil {
		err = w.fh.Close()
		w.fh = nil
	}
	return
}

type telemetryEventSender struct {
	writer io.WriteCloser
}

func newTelemetryEventSender() *telemetryEventSender {
	return newTelemetryEventSenderWithWriteCloser(&telemetryEventWriter{})
}

func sendTelemetry(sender *telemetryEventSender, name, version string) func(operation, message string, isSuccess bool, duration time.Duration) error {
	return func(operation, message string, isSuccess bool, duration time.Duration) error {
		e := newTelemetryEvent(name, version, operation, message, isSuccess, duration)
		return sender.send(e)
	}
}

func newTelemetryEventSenderWithWriteCloser(writer io.WriteCloser) *telemetryEventSender {
	return &telemetryEventSender{writer: writer}
}

func (w *telemetryEventSender) send(e telemetryEvent) error {
	defer w.writer.Close()

	bs, err := json.Marshal(e)
	if err != nil {
		return errors.Wrap(err, "failed to marhsal telemetry event")
	}

	_, err = w.writer.Write(bs)
	if err != nil {
		return errors.Wrap(err, "failed to write telemetry event")
	}

	return nil
}

func getTelemetryFileName() string {
	fn := fmt.Sprintf("%d.tld", time.Now().UnixNano())
	return path.Join(telemetryEventsPath, fn)
}

func newTelemetryEvent(name, version, operation, message string, isSuccess bool, duration time.Duration) telemetryEvent {
	return telemetryEvent{
		EventID:    1,
		ProviderID: "69B669B9-4AF8-4C50-BDC4-6006FA76E975",
		Parameters: []interface{}{
			telemetryParameterString{
				Name:  "Name",
				Value: name,
			},
			telemetryParameterString{
				Name:  "Version",
				Value: version,
			},
			telemetryParameterString{
				Name:  "Operation",
				Value: operation,
			},
			telemetryParameterBool{
				Name:  "OperationSuccess",
				Value: isSuccess,
			},
			telemetryParameterString{
				Name:  "Message",
				Value: message,
			},
			telemetryParameterLong{
				Name:  "Duration",
				Value: duration.Nanoseconds() / 1e6,
			},
		},
	}
}
