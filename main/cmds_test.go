package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/Azure/azure-extension-platform/pkg/extensionpolicysettings"
	"github.com/Azure/custom-script-extension-linux/pkg/errorutil"
	"github.com/ahmetalpbalkan/go-httpbin"
	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/require"
)

func Test_commandsExist(t *testing.T) {
	// we expect these subcommands to be handled
	expect := []string{"install", "enable", "disable", "uninstall", "update"}
	for _, c := range expect {
		_, ok := cmds[c]
		if !ok {
			t.Fatalf("cmd '%s' is not handled", c)
		}
	}
}

func Test_commands_shouldReportStatus(t *testing.T) {
	// - certain extension invocations are supposed to write 'N.status' files and some do not.

	// these subcommands should NOT report status
	require.False(t, cmds["install"].shouldReportStatus, "install should not report status")
	require.False(t, cmds["uninstall"].shouldReportStatus, "uninstall should not report status")

	// these subcommands SHOULD report status
	require.True(t, cmds["enable"].shouldReportStatus, "enable should report status")
	require.True(t, cmds["disable"].shouldReportStatus, "disable should report status")
	require.True(t, cmds["update"].shouldReportStatus, "update should report status")
}

func Test_checkAndSaveSeqNum_fails(t *testing.T) {
	// pass in invalid seqnum format
	_, err := checkAndSaveSeqNum(log.NewNopLogger(), 0, "/non/existing/dir")
	require.NotNil(t, err)
	require.Contains(t, err.Error(), `failed to save sequence number`)
}

func Test_checkAndSaveSeqNum(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	fp := filepath.Join(dir, "seqnum")
	defer os.RemoveAll(dir)

	nop := log.NewNopLogger()

	// no sequence number, 0 comes in.
	shouldExit, err := checkAndSaveSeqNum(nop, 0, fp)
	require.Nil(t, err)
	require.False(t, shouldExit)

	// file=0, seq=0 comes in. (should exit)
	shouldExit, err = checkAndSaveSeqNum(nop, 0, fp)
	require.Nil(t, err)
	require.True(t, shouldExit)

	// file=0, seq=1 comes in.
	shouldExit, err = checkAndSaveSeqNum(nop, 1, fp)
	require.Nil(t, err)
	require.False(t, shouldExit)

	// file=1, seq=1 comes in. (should exit)
	shouldExit, err = checkAndSaveSeqNum(nop, 1, fp)
	require.Nil(t, err)
	require.True(t, shouldExit)

	// file=1, seq=0 comes in. (should exit)
	shouldExit, err = checkAndSaveSeqNum(nop, 1, fp)
	require.Nil(t, err)
	require.True(t, shouldExit)
}

const policyTestDir = "./testdata"
const policyTestFile = "extensionPolicySettingsTestConfig.json"
const policyTestPath = policyTestDir + "/" + policyTestFile

func Test_LoadExtensionPolicySettings_PolicyFileExistsValid(t *testing.T) {
	// Set up test parameters
	require.Nil(t, setupPolicyDir(policyTestDir))
	defer cleanupPolicyFile(policyTestPath)
	// maybe clean up policy test directory too?
	require.Nil(t, loadTestPolicy("valid, basic", nil))

	// Replicate the logic in cmd.go enable()
	ExtensionPolicyManagerPtr, err := extensionpolicysettings.NewExtensionPolicySettingsManager[CSEExtensionPolicySettings](policyTestPath)
	require.NoError(t, err, "should be able to create extension policy settings manager")
	err = ExtensionPolicyManagerPtr.LoadExtensionPolicySettings()
	require.NoError(t, err, "should be able to load extension policy settings")
	// Check that settings are loaded correctly
	require.NoError(t, err)
	settings, err := ExtensionPolicyManagerPtr.GetSettings()
	require.NoError(t, err, "should be able to get extension policy settings")
	require.NotNil(t, settings, "settings should not be nil")
	require.Equal(t, false, settings.RequireSigning)
	require.Empty(t, settings.AllowedScripts)
}

func Test_LoadExtensionPolicySettings_PolicyFileMissing(t *testing.T) {
	// Replicate the logic in cmd.go enable()
	missingPolicyFilePath := filepath.Join(policyTestDir, "missingPolicyFile.json")
	ExtensionPolicyManagerPtr, err := extensionpolicysettings.NewExtensionPolicySettingsManager[CSEExtensionPolicySettings](missingPolicyFilePath)
	require.NoError(t, err, "should be able to create extension policy settings manager")

	_, err = os.Stat(missingPolicyFilePath)
	require.True(t, os.IsNotExist(err), "policy file should not exist")
	err = ExtensionPolicyManagerPtr.LoadExtensionPolicySettings()
	require.Error(t, err, "should not be able to load extension policy settings")
}

func Test_runCmd_success(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	defer os.RemoveAll(dir)

	require.Nil(t, runCmd(log.NewNopLogger(), dir, handlerSettings{
		publicSettings: publicSettings{CommandToExecute: "date"},
	}), "command should run successfully")
}

func Test_runCmd_fail(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	defer os.RemoveAll(dir)

	ewc := runCmd(log.NewNopLogger(), dir, handlerSettings{
		publicSettings: publicSettings{CommandToExecute: "non-existing-cmd"},
	})
	require.Equal(t, errorutil.CommandExecution_failureExitCode, ewc.ErrorCode)
	require.NotNil(t, ewc.Err, "command terminated with exit status")
	require.Contains(t, ewc.Err.Error(), "failed to execute command")
}

func Test_downloadFiles(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	defer os.RemoveAll(dir)

	srv := httptest.NewServer(httpbin.GetMux())
	defer srv.Close()

	ewc := downloadFiles(log.NewContext(log.NewNopLogger()),
		dir,
		handlerSettings{
			publicSettings: publicSettings{
				FileURLs: []string{
					srv.URL + "/bytes/10",
					srv.URL + "/bytes/100",
					srv.URL + "/bytes/1000",
				}},
		}, nil)
	require.Nil(t, ewc)

	// check the files
	f := []string{"10", "100", "1000"}
	for _, fn := range f {
		fp := filepath.Join(dir, fn)
		_, err := os.Stat(fp)
		data, err := os.ReadFile(fp)
		fmt.Println("File Content:")
		fmt.Println(string(data))

		require.Nil(t, err, "%s is missing from download dir", fp)
	}
}

func Test_downloadFiles_goodAllowlist_SHA256(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	defer os.RemoveAll(dir)

	file1Content := "echo hello"
	file2Content := "echo world"
	file3Content := "echo !"

	// Create a custom HTTP server with preset content
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { // r.URL.Path is our key string to fetch file content below.
		// Define preset file contents
		files := map[string]string{
			"/file1": file1Content,
			"/file2": file2Content,
			"/file3": file3Content,
		}

		if content, ok := files[r.URL.Path]; ok { // 'content' is value corresponding to the key r.URL.Path. 'ok' is true if content exists.
			w.Header().Set("Content-Type", "application/octet-stream") // 'content' is going to be the response body from the test server.
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, content)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	// Compute SHA256 hashes (you can use the extensionpolicysettings.ComputeFileHash function)
	file1Hash, _ := extensionpolicysettings.ComputeFileHash(file1Content, extensionpolicysettings.HashTypeSHA256)
	file2Hash, _ := extensionpolicysettings.ComputeFileHash(file2Content, extensionpolicysettings.HashTypeSHA256)
	file3Hash, _ := extensionpolicysettings.ComputeFileHash(file3Content, extensionpolicysettings.HashTypeSHA256)

	// Create a hash list.
	al := []string{file1Hash, file2Hash, file3Hash}

	// Write the policy file (GA behavior)
	require.Nil(t, loadTestPolicy("valid, allowlist present", al))
	defer cleanupPolicyFile(policyTestPath)

	// Load the policy into manager (enable() behavior)
	ExtensionPolicyManagerPtr, err := extensionpolicysettings.NewExtensionPolicySettingsManager[CSEExtensionPolicySettings](policyTestPath)
	require.NoError(t, err, "should be able to create extension policy settings manager")
	err = ExtensionPolicyManagerPtr.LoadExtensionPolicySettings()
	require.NoError(t, err, "should be able to load extension policy settings")

	ewc := downloadFiles(log.NewContext(log.NewNopLogger()),
		dir,
		handlerSettings{
			publicSettings: publicSettings{
				FileURLs: []string{
					srv.URL + "/file1",
					srv.URL + "/file2",
					srv.URL + "/file3",
				}},
		}, ExtensionPolicyManagerPtr)
	require.Nil(t, ewc)

	// check the files. All files should have passed.
	f := []string{"file1", "file2", "file3"}
	for _, fn := range f {
		fp := filepath.Join(dir, fn)
		_, err := os.Stat(fp)
		require.Nil(t, err, "%s is missing from download dir", fp)
	}
}

func Test_downloadFiles_badAllowlist(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	defer os.RemoveAll(dir)

	// Generate hashes of the preset content for the allowlist
	file1Content := "echo hello"
	file2Content := "echo world"
	file3Content := "echo !"
	file4Content := "echo bad"
	file5Content := "echo also bad"

	// Create a custom HTTP server with preset content
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Define preset file contents
		files := map[string]string{
			"/file1": file1Content,
			"/file2": file2Content,
			"/file3": file3Content,
			"/file4": file4Content,
			"/file5": file5Content,
		}

		if content, ok := files[r.URL.Path]; ok {
			w.Header().Set("Content-Type", "application/octet-stream")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, content)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	// Compute SHA256 hashes (you can use the extensionpolicysettings.ComputeFileHash function)
	file1Hash, _ := extensionpolicysettings.ComputeFileHash(file1Content, extensionpolicysettings.HashTypeSHA256)
	file2Hash, _ := extensionpolicysettings.ComputeFileHash(file2Content, extensionpolicysettings.HashTypeSHA256)
	file3Hash, _ := extensionpolicysettings.ComputeFileHash(file3Content, extensionpolicysettings.HashTypeSHA256)

	// Create a hash list.
	al := []string{file1Hash, file2Hash, file3Hash}

	// Write the policy file (GA behavior)
	require.Nil(t, loadTestPolicy("valid, allowlist present", al))
	defer cleanupPolicyFile(policyTestPath)

	// Load the policy into manager (enable() behavior)
	ExtensionPolicyManagerPtr, err := extensionpolicysettings.NewExtensionPolicySettingsManager[CSEExtensionPolicySettings](policyTestPath)
	require.NoError(t, err, "should be able to create extension policy settings manager")
	err = ExtensionPolicyManagerPtr.LoadExtensionPolicySettings()
	require.NoError(t, err, "should be able to load extension policy settings")

	ewc := downloadFiles(log.NewContext(log.NewNopLogger()),
		dir,
		handlerSettings{
			publicSettings: publicSettings{
				FileURLs: []string{
					srv.URL + "/file1",
					srv.URL + "/file2",
					srv.URL + "/file3",
					srv.URL + "/file4", // this file is not in the allowlist. Extension should exit gracefully.
				}},
		}, ExtensionPolicyManagerPtr)
	require.NotNil(t, ewc)
	require.Contains(t, ewc.Err.Error(), "Validation of script 'file4' against policy-allowlist failed", "error should indicate that file4 failed validation")

	// Check the files. At most file1, file2, and file3 should be present. file4 and file5 should not both be present,
	// but one of them will be present because they are validated after being downloaded.
	// The extension should exit gracefully. If downloaded out of order, it's possible that not file1, file2, and file3 are all
	// present because the extension will exit immediately after downloading the first file that is not in the allowlist.

	// The check below assumes the files are downloaded in order.
	f := []string{"file1", "file2", "file3", "file4", "file5"}
	for _, fn := range f {
		fp := filepath.Join(dir, fn)
		_, err := os.Stat(fp)
		if fn == "file5" {
			require.NotNil(t, err, "%s should not be downloaded because it's not in the allowlist", fp)
			continue
		} else {
			require.Nil(t, err, "%s is missing from download dir", fp)
		}
	}
}

func Test_downloadFiles_emptyAllowlist(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	defer os.RemoveAll(dir)

	srv := httptest.NewServer(httpbin.GetMux())
	defer srv.Close()

	// Create an EMPTY list.
	al := []string{}

	// Write the policy file (GA behavior)
	require.Nil(t, loadTestPolicy("valid, allowlist present", al))
	defer cleanupPolicyFile(policyTestPath)

	// Load the policy into manager (enable() behavior)
	ExtensionPolicyManagerPtr, err := extensionpolicysettings.NewExtensionPolicySettingsManager[CSEExtensionPolicySettings](policyTestPath)
	require.NoError(t, err, "should be able to create extension policy settings manager")
	err = ExtensionPolicyManagerPtr.LoadExtensionPolicySettings()
	require.NoError(t, err, "should be able to load extension policy settings")

	ewc := downloadFiles(log.NewContext(log.NewNopLogger()),
		dir,
		handlerSettings{
			publicSettings: publicSettings{
				FileURLs: []string{
					srv.URL + "/bytes/10",
					srv.URL + "/bytes/100",
					srv.URL + "/bytes/1000",
				}},
		}, ExtensionPolicyManagerPtr)
	require.Nil(t, ewc)

	// Check the files. All files should have passed.
	f := []string{"10", "100", "1000"}
	for _, fn := range f {
		fp := filepath.Join(dir, fn)
		_, err := os.Stat(fp)
		require.Nil(t, err, "%s is missing from download dir", fp)
	}
}

func Test_decodeScript(t *testing.T) {
	testSubject := "bHMK"
	s, info, err := decodeScript(testSubject)

	require.NoError(t, err)
	require.Equal(t, info, "4;3;gzip=0")
	require.Equal(t, s, "ls\n")
}

func Test_decodeScriptGzip(t *testing.T) {
	testSubject := "H4sIACD731kAA8sp5gIAfShLWgMAAAA="
	s, info, err := decodeScript(testSubject)

	require.NoError(t, err)
	require.Equal(t, info, "32;3;gzip=1")
	require.Equal(t, s, "ls\n")
}

// Helper Methods
func writeToFile(filePath, content string) error {
	err := os.WriteFile(filePath, []byte(content), 0644)
	return err
}

func cleanupPolicyFile(path string) {
	// Do not remove missingPolicyFilePath as it simulates a missing file
	if _, err := os.Stat(path); err == nil {
		os.Remove(path)
	}
}

func setupPolicyDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, 0750)
		return err
	}
	return nil
}

func loadTestPolicy(scenario string, list []string) error {
	var validPolicyContent string
	var allowlistStr string
	if list != nil {
		allowlistStr = "["
		for i, item := range list {
			allowlistStr += `"` + item + `"`
			if i < len(list)-1 {
				allowlistStr += ","
			}
		}
		allowlistStr += "]"
	}

	switch scenario {
	case "valid, basic":
		validPolicyContent = `{
				"requireSigning": false,
				"allowedScripts": []
			}`
	case "valid, allowlist present":
		// Convert list to JSON array string

		validPolicyContent = `{
				"requireSigning": false,
				"allowedScripts": ` + allowlistStr + `
			}`
	default:
		validPolicyContent = `{}`
	}
	return writeToFile(filepath.Join(policyTestDir, policyTestFile), validPolicyContent)
}
