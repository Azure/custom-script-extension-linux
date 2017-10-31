package main

import (
	"bytes"
	"encoding/json"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_newTelemetryEvent(t *testing.T) {
	duration, _ := time.ParseDuration("2m30s")
	testSubject := newTelemetryEvent("--Name--", "--Version--", "--Operation--", "--Message--", true, duration)

	require.Equal(t, "Name", testSubject.Parameters[0].(telemetryParameterString).Name, "name")
	require.Equal(t, "Version", testSubject.Parameters[1].(telemetryParameterString).Name, "version")
	require.Equal(t, "Operation", testSubject.Parameters[2].(telemetryParameterString).Name, "operation")
	require.Equal(t, "OperationSuccess", testSubject.Parameters[3].(telemetryParameterBool).Name, "operationSuccess")
	require.Equal(t, "Message", testSubject.Parameters[4].(telemetryParameterString).Name, "message")
	require.Equal(t, "Duration", testSubject.Parameters[5].(telemetryParameterLong).Name, "duration")

	require.Equal(t, "--Name--", testSubject.Parameters[0].(telemetryParameterString).Value, "name")
	require.Equal(t, "--Version--", testSubject.Parameters[1].(telemetryParameterString).Value, "version")
	require.Equal(t, "--Operation--", testSubject.Parameters[2].(telemetryParameterString).Value, "operation")
	require.Equal(t, true, testSubject.Parameters[3].(telemetryParameterBool).Value, "operationSuccess")
	require.Equal(t, "--Message--", testSubject.Parameters[4].(telemetryParameterString).Value, "message")
	require.Equal(t, int64(150)*1000, testSubject.Parameters[5].(telemetryParameterLong).Value)
}

func Test_serializeTelemetryEvent(t *testing.T) {
	duration, _ := time.ParseDuration("2m30s")
	testSubject := newTelemetryEvent("--Name--", "--Version--", "--Operation--", "--Message--", true, duration)

	bs, err := json.Marshal(testSubject)
	require.NoError(t, err)

	json := `{
    "eventId": 1,
    "providerId": "69B669B9-4AF8-4C50-BDC4-6006FA76E975",
    "parameters": [
        {
            "name": "Name",
            "value": "--Name--"
        },
        {
            "name": "Version",
            "value": "--Version--"
        },
        {
            "name": "Operation",
            "value": "--Operation--"
        },
        {
            "name": "OperationSuccess",
            "value": true
        },
        {
            "name": "Message",
            "value": "--Message--"
        },
        {
            "name": "Duration",
            "value": 150000
        }
    ]
}`
	require.JSONEq(t, json, string(bs))
}

type mockWriteCloser struct {
	isClosed bool
	buf      *bytes.Buffer
}

func (s *mockWriteCloser) Write(bs []byte) (int, error) {
	return s.buf.Write(bs)
}

func (s *mockWriteCloser) Close() error {
	s.isClosed = true
	return nil
}

func Test_telemetryEventSender(t *testing.T) {
	writeCloser := &mockWriteCloser{
		isClosed: false,
		buf:      bytes.NewBufferString(""),
	}

	duration, _ := time.ParseDuration("2m30s")
	event := newTelemetryEvent("--Name--", "--Version--", "--Operation--", "--Message--", true, duration)

	testSubject := newTelemetryEventSenderWithWriteCloser(writeCloser)
	testSubject.send(event)

	require.Equal(t, true, writeCloser.isClosed, "expected writeCloser to be closed")

	json := `{
    "eventId": 1,
    "providerId": "69B669B9-4AF8-4C50-BDC4-6006FA76E975",
    "parameters": [
        {
            "name": "Name",
            "value": "--Name--"
        },
        {
            "name": "Version",
            "value": "--Version--"
        },
        {
            "name": "Operation",
            "value": "--Operation--"
        },
        {
            "name": "OperationSuccess",
            "value": true
        },
        {
            "name": "Message",
            "value": "--Message--"
        },
        {
            "name": "Duration",
            "value": 150000
        }
    ]
}`
	require.JSONEq(t, writeCloser.buf.String(), json)
}

func Test_getTelemetryFileName(t *testing.T) {
	testSubject := getTelemetryFileName()
	require.True(t, regexp.MustCompile("^/var/lib/waagent/events/\\d{19}\\.tld$").Match([]byte(testSubject)), testSubject)
}
