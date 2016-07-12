package main

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExec_success(t *testing.T) {
	v := new(mockFile)
	ec, err := Exec("date", "/", v, v)
	require.Nil(t, err, "err: %v -- out: %s", err, v.b.Bytes())
	require.EqualValues(t, 0, ec)
}

func TestExec_success_redirectsStdStreams_closesFds(t *testing.T) {
	o, e := new(mockFile), new(mockFile)
	require.False(t, o.closed, "stdout open")
	require.False(t, e.closed, "stderr open")

	_, err := Exec("/bin/echo 'I am stdout!'>&1; /bin/echo 'I am stderr!'>&2", "/", o, e)
	require.Nil(t, err, "err: %v -- stderr: %s", err, e.b.Bytes())
	require.Equal(t, "I am stdout!\n", string(o.b.Bytes()))
	require.Equal(t, "I am stderr!\n", string(e.b.Bytes()))
	require.True(t, o.closed, "stdout closed")
	require.True(t, e.closed, "stderr closed")
}

func TestExec_failure_exitError(t *testing.T) {
	ec, err := Exec("exit 12", "/", new(mockFile), new(mockFile))
	require.NotNil(t, err)
	require.EqualError(t, err, "command terminated with exit status=12") // error is customized
	require.EqualValues(t, 12, ec)
}

func TestExec_failure_genericError(t *testing.T) {
	_, err := Exec("date", "/non-existing-path", new(mockFile), new(mockFile))
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "failed to execute command:") // error is wrapped
}

func TestExec_failure_redirectsStdStreams_closesFds(t *testing.T) {
	o, e := new(mockFile), new(mockFile)
	require.False(t, o.closed, "stdout open")
	require.False(t, e.closed, "stderr open")

	_, err := Exec(`/bin/echo 'I am stdout!'>&1; /bin/echo 'I am stderr!'>&2; exit 12`, "/", o, e)
	require.NotNil(t, err)
	require.Equal(t, "I am stdout!\n", string(o.b.Bytes()))
	require.Equal(t, "I am stderr!\n", string(e.b.Bytes()))
	require.True(t, o.closed, "stdout closed")
	require.True(t, e.closed, "stderr closed")
}

// Test utilities

type mockFile struct {
	b      bytes.Buffer
	closed bool
}

func (m *mockFile) Write(p []byte) (n int, err error) {
	if m.closed {
		return 0, errors.New("file closed")
	}
	return m.b.Write(p)
}

func (m *mockFile) Close() error {
	m.closed = true
	return nil
}
