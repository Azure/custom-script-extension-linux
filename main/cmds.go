package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
)

const (
	maxScriptSize = 256 * 1024
)

type cmdFunc func(ctx *log.Context, hEnv HandlerEnvironment, report *RunCommandInstanceView, extName string, seqNum int) (stdout string, stderr string, err error)
type preFunc func(ctx *log.Context, seqNum int) error

type cmd struct {
	invoke             cmdFunc // associated function
	name               string  // human readable string
	shouldReportStatus bool    // determines if running this should log to a .status file
	pre                preFunc // executed before any status is reported
	failExitCode       int     // exitCode to use when commands fail
}

const (
	fullName                = "Microsoft.Compute.CPlat.Core.RunCommandLinux"
	maxTailLen              = 4 * 1024 // length of max stdout/stderr to be transmitted in .status file
	maxTelemetryTailLen int = 1800
)

var (
	telemetry = sendTelemetry(newTelemetryEventSender(), fullName, Version)

	cmdInstall   = cmd{install, "Install", false, nil, 52}
	cmdEnable    = cmd{enable, "Enable", true, enablePre, 3}
	cmdDisable   = cmd{disable, "Disable", true, nil, 3}
	cmdUpdate    = cmd{update, "Update", true, nil, 3}
	cmdUninstall = cmd{uninstall, "Uninstall", false, nil, 3}

	cmds = map[string]cmd{
		"install":   cmdInstall,
		"enable":    cmdEnable,
		"disable":   cmdDisable,
		"update":    cmdUpdate,
		"uninstall": cmdUninstall,
	}
)

func update(ctx *log.Context, h HandlerEnvironment, report *RunCommandInstanceView, extName string, seqNum int) (string, string, error) {
	ctx.Log("event", "update")
	return "", "", nil
}

func disable(ctx *log.Context, h HandlerEnvironment, report *RunCommandInstanceView, extName string, seqNum int) (string, string, error) {
	ctx.Log("event", "disable")
	KillPreviousExtension(ctx, pidFilePath)
	return "", "", nil
}

func install(ctx *log.Context, h HandlerEnvironment, report *RunCommandInstanceView, extName string, seqNum int) (string, string, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return "", "", errors.Wrap(err, "failed to create data dir")
	}

	// If the file mrseq does not exists
	//  the extension has never been installed on this VMs before
	// if _, err := os.Stat(mostRecentSequence); os.IsNotExist(err) {
	// 	migrateToMostRecentSequence(ctx, h, seqNum)
	// }

	ctx.Log("event", "created data dir", "path", dataDir)
	ctx.Log("event", "installed")
	return "", "", nil
}

func migrateToMostRecentSequence(ctx *log.Context, h HandlerEnvironment, seqNum int) {
	// The status folder is used instead of the settings because the settings file is written
	// by the agent before install is called.  As a result, the extension cannot determine if this
	// is a new install or an upgrade.
	//
	// If this is an upgrade there will be a status file. The agent will re-write the last status
	// file to indicate that the upgrade happened successfully. The extension uses the last status
	// sequence number to determine the last settings file that was executed.
	//
	// The agent helpfully copies mrseq every time an extension is upgraded thereby preserving the
	// most recent executed sequence. If extensions use mrseq they benefit from this mechanism, and
	// do not have invent another method.  The CustomScript extension should have been using this
	// from the beginning, but it was not.
	//
	computedSeqNum, err := FindSeqNumStatus(h.HandlerEnvironment.StatusFolder)
	if err != nil {
		// If there was an error, the sequence number is zero.
		ctx.Log("event", "migrate to mrseq", "error", err)
		return
	}

	fout, err := os.Create(mostRecentSequence)
	if err != nil {
		ctx.Log("event", "migrate to mrseq", "error", err)
		return
	}
	defer fout.Close()

	ctx.Log("event", "migrate to mrseq", "message", fmt.Sprintf("migrated mrseq to %v", computedSeqNum))
	fout.WriteString(fmt.Sprintf("%v", computedSeqNum))
}

func uninstall(ctx *log.Context, h HandlerEnvironment, report *RunCommandInstanceView, extName string, seqNum int) (string, string, error) {
	{ // a new context scope with path
		ctx = ctx.With("path", dataDir)
		ctx.Log("event", "removing data dir", "path", dataDir)
		if err := os.RemoveAll(dataDir); err != nil {
			return "", "", errors.Wrap(err, "failed to delete data directory")
		}
		ctx.Log("event", "removed data dir")
	}
	ctx.Log("event", "uninstalled")
	return "", "", nil
}

func enablePre(ctx *log.Context, seqNum int) error {
	// exit if this sequence number (a snapshot of the configuration) is already
	// processed. if not, save this sequence number before proceeding.
	if shouldExit, err := checkAndSaveSeqNum(ctx, seqNum, mostRecentSequence); err != nil {
		return errors.Wrap(err, "failed to process sequence number")
	} else if shouldExit {
		ctx.Log("event", "exit", "message", "the script configuration has already been processed, will not run again")
		os.Exit(0)
	}
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func enable(ctx *log.Context, h HandlerEnvironment, report *RunCommandInstanceView, extName string, seqNum int) (string, string, error) {
	// parse the extension handler settings (not available prior to 'enable')
	configFile := fmt.Sprintf("%d.settings", seqNum)
	if extName != "" {
		configFile = extName + "." + configFile
	}
	configPath := filepath.Join(h.HandlerEnvironment.ConfigFolder, configFile)
	cfg, err := parseAndValidateSettings(ctx, configPath)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to get configuration")
	}

	dir := filepath.Join(dataDir, downloadDir, fmt.Sprintf("%d", seqNum))
	if err := downloadFiles(ctx, dir, &cfg); err != nil {
		return "", "", errors.Wrap(err, "processing file downloads failed")
	}

	// AsyncExecution requested by customer means the extension should report successful extension deployment to complete the provisioning state
	// Later the full extension output will be reported
	if cfg.AsyncExecution {
		ctx.Log("message", "anycExecution is true - report success")
		reportInstanceView(ctx, h, extName, seqNum, StatusSuccess, cmd{nil, "Enable", true, nil, 3}, report)
	}

	// execute the command, save its error
	runErr := runCmd(ctx, dir, &cfg)

	// collect the logs if available
	stdoutF, stderrF := logPaths(dir)
	stdoutTail, err := tailFile(stdoutF, maxTailLen)
	if err != nil {
		ctx.Log("message", "error tailing stdout logs", "error", err)
	}
	stderrTail, err := tailFile(stderrF, maxTailLen)
	if err != nil {
		ctx.Log("message", "error tailing stderr logs", "error", err)
	}

	isSuccess := runErr == nil
	telemetry("Output", "-- stdout/stderr omitted from telemetry pipeline --", isSuccess, 0)

	if isSuccess {
		ctx.Log("event", "enabled")
	} else {
		ctx.Log("event", "enable script failed")
	}

	// Always report nil for error because extension should not fail if script throws error
	// Execution error still will be reported in the error stream
	return string(stdoutTail), string(stderrTail), nil
}

// checkAndSaveSeqNum checks if the given seqNum is already processed
// according to the specified seqNumFile and if so, returns true,
// otherwise saves the given seqNum into seqNumFile returns false.
func checkAndSaveSeqNum(ctx log.Logger, seq int, mrseqPath string) (shouldExit bool, _ error) {
	ctx.Log("event", "comparing seqnum", "path", mrseqPath)
	smaller, err := IsSmallerThan(mrseqPath, seq)
	if err != nil {
		return false, errors.Wrap(err, "failed to check sequence number")
	}
	if !smaller {
		// stored sequence number is equals or greater than the current
		// sequence number.
		return true, nil
	}
	if err := SaveSeqNum(mrseqPath, seq); err != nil {
		return false, errors.Wrap(err, "failed to save sequence number")
	}
	ctx.Log("event", "seqnum saved", "path", mrseqPath)
	return false, nil
}

// downloadFiles downloads the files specified in cfg into dir (creates if does
// not exist) and takes storage credentials specified in cfg into account.
func downloadFiles(ctx *log.Context, dir string, cfg *handlerSettings) error {
	// - prepare the output directory for files and the command output
	// - create the directory if missing
	ctx.Log("event", "creating output directory", "path", dir)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return errors.Wrap(err, "failed to prepare output directory")
	}
	ctx.Log("event", "created output directory")

	dos2unix := 1

	// - download scriptURI
	scriptURI := cfg.scriptURI()
	ctx.Log("scriptUri", scriptURI)
	if scriptURI != "" {
		telemetry("scenario", fmt.Sprintf("source.scriptUri;dos2unix=%d", dos2unix), true, 0*time.Millisecond)
		ctx.Log("event", "download start")
		if err := downloadAndProcessURL(ctx, scriptURI, dir, cfg); err != nil {
			ctx.Log("event", "download failed", "error", err)
			return errors.Wrapf(err, "failed to download file %s", scriptURI)
		}
		ctx.Log("event", "download complete", "output", dir)
	}
	return nil
}

// runCmd runs the command (extracted from cfg) in the given dir (assumed to exist).
func runCmd(ctx *log.Context, dir string, cfg *handlerSettings) (err error) {
	ctx.Log("event", "executing command", "output", dir)
	var cmd string
	var scenario string
	var scenarioInfo string

	// If script is specified - use it directly for command
	if cfg.script() != "" {
		cmd = cfg.script()
		scenario = fmt.Sprintf("public-script;%s", scenarioInfo)
	} else if cfg.scriptURI() != "" {
		// If scriptUri is specified then cmd should start it
		cmd = fmt.Sprintf("%s/script.sh", dir)
		scenario = "public-scriptUri"
	}

	ctx.Log("event", "prepare command", "cmd", cmd)

	// We need to kill previous extension process if exists before staring a new one.
	KillPreviousExtension(ctx, pidFilePath)

	// Store the active process id and start time in case its a long running process that needs to be killed later
	// If process exited successfully the pid file is deleted
	SaveCurrentPidAndStartTime(pidFilePath)
	defer DeleteCurrentPidAndStartTime(pidFilePath)

	begin := time.Now()
	err = ExecCmdInDir(ctx, cmd, dir, cfg)
	elapsed := time.Now().Sub(begin)
	isSuccess := err == nil

	telemetry("scenario", scenario, isSuccess, elapsed)

	if err != nil {
		ctx.Log("event", "failed to execute command", "error", err, "output", dir)
		return errors.Wrap(err, "failed to execute command")
	}
	ctx.Log("event", "executed command", "output", dir)
	return nil
}

func writeTempScript(script, dir string) (string, string, error) {
	if len(script) > maxScriptSize {
		return "", "", fmt.Errorf("The script's length (%d) exceeded the maximum allowed length of %d", len(script), maxScriptSize)
	}

	s, info, err := decodeScript(script)
	if err != nil {
		return "", "", err
	}

	cmd := fmt.Sprintf("%s/script.sh", dir)
	f, err := os.OpenFile(cmd, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0500)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to write script.sh")
	}

	f.WriteString(s)
	f.Close()

	dos2unix := 1
	err = postProcessFile(cmd)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to post-process script.sh")
	}
	dos2unix = 0
	return cmd, fmt.Sprintf("%s;dos2unix=%d", info, dos2unix), nil
}

// base64 decode and optionally GZip decompress a script
func decodeScript(script string) (string, string, error) {
	// scripts must be base64 encoded
	s, err := base64.StdEncoding.DecodeString(script)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to decode script")
	}

	// scripts may be gzip'ed
	r, err := gzip.NewReader(bytes.NewReader(s))
	if err != nil {
		return string(s), fmt.Sprintf("%d;%d;gzip=0", len(script), len(s)), nil
	}

	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	n, err := io.Copy(w, r)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to decompress script")
	}

	w.Flush()
	return buf.String(), fmt.Sprintf("%d;%d;gzip=1", len(script), n), nil
}
