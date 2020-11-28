package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/require"
)

func Test_serializeInstanceView(t *testing.T) {
	instanceView := RunCommandInstanceView{
		ExecutionState:   Running,
		ExecutionMessage: "Completed",
		Output:           "Script output stream with \\ \n \t \"  ",
		Error:            "Script error stream",
		StartTime:        time.Date(2000, 2, 1, 12, 30, 0, 0, time.UTC).Format(time.RFC3339),
		EndTime:          time.Date(2000, 2, 1, 12, 35, 0, 0, time.UTC).Format(time.RFC3339),
	}
	msg, err := serializeInstanceView(&instanceView)
	require.Nil(t, err)
	require.NotNil(t, msg)
	expectedMsg := "{\"executionState\":\"Running\",\"executionMessage\":\"Completed\",\"output\":\"Script output stream with \\\\ \\n \\t \\\"  \",\"error\":\"Script error stream\",\"exitCode\":0,\"startTime\":\"2000-02-01T12:30:00Z\",\"endTime\":\"2000-02-01T12:35:00Z\"}"
	require.Equal(t, expectedMsg, msg)

	var iv RunCommandInstanceView
	json.Unmarshal([]byte(msg), &iv)
	require.Equal(t, instanceView, iv)
}

func Test_reportInstanceView(t *testing.T) {
	instanceView := RunCommandInstanceView{
		ExecutionState:   Running,
		ExecutionMessage: "Completed",
		Output:           "Script output stream with \\ \n \t \"  ",
		Error:            "Script error stream",
		StartTime:        time.Date(2000, 2, 1, 12, 30, 0, 0, time.UTC).Format(time.RFC3339),
		EndTime:          time.Date(2000, 2, 1, 12, 35, 0, 0, time.UTC).Format(time.RFC3339),
	}
	tmpDir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	defer os.RemoveAll(tmpDir)

	extName := "first"
	fakeEnv := HandlerEnvironment{}
	fakeEnv.HandlerEnvironment.StatusFolder = tmpDir

	require.Nil(t, reportInstanceView(log.NewContext(log.NewNopLogger()), fakeEnv, extName, 1, StatusSuccess, cmdEnable, &instanceView))

	path := filepath.Join(tmpDir, extName+"."+"1.status")
	b, err := ioutil.ReadFile(path)
	require.Nil(t, err, ".status file exists")
	require.NotEqual(t, 0, len(b), ".status file not empty")

	var r StatusReport
	json.Unmarshal(b, &r)
	require.Equal(t, 1, len(r))
	require.Equal(t, StatusSuccess, r[0].Status.Status)
	require.Equal(t, cmdEnable.name, r[0].Status.Operation)

	msg, _ := serializeInstanceView(&instanceView)
	require.Equal(t, msg, r[0].Status.FormattedMessage.Message)
}
