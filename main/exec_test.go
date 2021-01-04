package main

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/require"
)

var (
	testHandlerSettings handlerSettings = handlerSettings{publicSettings{}, protectedSettings{}}
	testContext                         = log.NewContext(log.NewNopLogger())
)

func TestExec_success(t *testing.T) {
	v := new(mockFile)
	ec, err := Exec(testContext, "date", "/", v, v, &testHandlerSettings)
	require.Nil(t, err, "err: %v -- out: %s", err, v.b.Bytes())
	require.EqualValues(t, 0, ec)
}

func TestExec_success_redirectsStdStreams_closesFds(t *testing.T) {
	o, e := new(mockFile), new(mockFile)
	require.False(t, o.closed, "stdout open")
	require.False(t, e.closed, "stderr open")

	_, err := Exec(testContext, "/bin/echo 'I am stdout!'>&1; /bin/echo 'I am stderr!'>&2", "/", o, e, &testHandlerSettings)
	require.Nil(t, err, "err: %v -- stderr: %s", err, e.b.Bytes())
	require.Equal(t, "I am stdout!\n", string(o.b.Bytes()))
	require.Equal(t, "I am stderr!\n", string(e.b.Bytes()))
	require.True(t, o.closed, "stdout closed")
	require.True(t, e.closed, "stderr closed")
}

func TestExec_failure_exitError(t *testing.T) {
	ec, err := Exec(testContext, "exit 12", "/", new(mockFile), new(mockFile), &testHandlerSettings)
	require.NotNil(t, err)
	require.EqualError(t, err, "command terminated with exit status=12") // error is customized
	require.EqualValues(t, 12, ec)
}

func TestExec_failure_timeout(t *testing.T) {
	testHandlerSettings.publicSettings.TimeoutInSeconds = 1
	ec, err := Exec(testContext, "sleep 20", "/", new(mockFile), new(mockFile), &testHandlerSettings)
	testHandlerSettings.publicSettings.TimeoutInSeconds = 0
	require.NotNil(t, err)
	require.EqualError(t, err, "command terminated with exit status=-1") // error is customized
	require.EqualValues(t, -1, ec)
}

// func TestExec_runasuser(t *testing.T) {
// 	if os.Geteuid() != 0 {
// 		fmt.Println("SKIP: Should be run under root. Use sudo.")
// 		return
// 	}

// 	o, e := new(mockFile), new(mockFile)
// 	testHandlerSettings.publicSettings.RunAsUser = "runcommand"
// 	exitCode, err := Exec(testContext, "whoami", "/", o, e, &testHandlerSettings)
// 	testHandlerSettings.publicSettings.RunAsUser = ""
// 	require.Nil(t, err)
// 	require.EqualValues(t, 0, exitCode)
// 	require.Equal(t, "runcommand\n", string(o.b.Bytes()))
// }

func TestExec_failure_genericError(t *testing.T) {
	_, err := Exec(testContext, "date", "/non-existing-path", new(mockFile), new(mockFile), &testHandlerSettings)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "failed to execute command:") // error is wrapped
}

func TestExec_failure_fdClosed(t *testing.T) {
	out := new(mockFile)
	require.Nil(t, out.Close())

	_, err := Exec(testContext, "date", "/", out, out, &testHandlerSettings)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "file closed") // error is wrapped
}

func TestExec_failure_redirectsStdStreams_closesFds(t *testing.T) {
	o, e := new(mockFile), new(mockFile)
	require.False(t, o.closed, "stdout open")
	require.False(t, e.closed, "stderr open")

	_, err := Exec(testContext, `/bin/echo 'I am stdout!'>&1; /bin/echo 'I am stderr!'>&2; exit 12`, "/", o, e, &testHandlerSettings)
	require.NotNil(t, err)
	require.Equal(t, "I am stdout!\n", string(o.b.Bytes()))
	require.Equal(t, "I am stderr!\n", string(e.b.Bytes()))
	require.True(t, o.closed, "stdout closed")
	require.True(t, e.closed, "stderr closed")
}

func TestExecCmdInDir(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	defer os.RemoveAll(dir)

	err = ExecCmdInDir(testContext, "/bin/echo 'Hello world'", dir, &testHandlerSettings)
	require.Nil(t, err)
	require.True(t, fileExists(t, filepath.Join(dir, "stdout")), "stdout file should be created")
	require.True(t, fileExists(t, filepath.Join(dir, "stderr")), "stderr file should be created")

	b, err := ioutil.ReadFile(filepath.Join(dir, "stdout"))
	require.Nil(t, err)
	require.Equal(t, "Hello world\n", string(b))

	b, err = ioutil.ReadFile(filepath.Join(dir, "stderr"))
	require.Nil(t, err)
	require.EqualValues(t, 0, len(b), "stderr file must be empty")
}

func TestExecCmdInDir_cantOpenError(t *testing.T) {
	err := ExecCmdInDir(testContext, "/bin/echo 'Hello world'", "/non-existing-dir", &testHandlerSettings)
	require.Contains(t, err.Error(), "failed to open stdout file")
}

func TestExecCmdInDir_truncates(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	defer os.RemoveAll(dir)

	require.Nil(t, ExecCmdInDir(testContext, "/bin/echo '1:out'; /bin/echo '1:err'>&2", dir, &testHandlerSettings))
	require.Nil(t, ExecCmdInDir(testContext, "/bin/echo '2:out'; /bin/echo '2:err'>&2", dir, &testHandlerSettings))

	b, err := ioutil.ReadFile(filepath.Join(dir, "stdout"))
	require.Nil(t, err)
	require.Equal(t, "2:out\n", string(b), "stdout did not truncate")

	b, err = ioutil.ReadFile(filepath.Join(dir, "stderr"))
	require.Nil(t, err)
	require.Equal(t, "2:err\n", string(b), "stderr did not truncate")
}

func Test_logPaths(t *testing.T) {
	stdout, stderr := logPaths("/tmp")
	require.Equal(t, "/tmp/stdout", stdout)
	require.Equal(t, "/tmp/stderr", stderr)
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

func fileExists(t *testing.T, path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	t.Fatalf("failed to check if %s exists: %v", path, err)
	return false
}
