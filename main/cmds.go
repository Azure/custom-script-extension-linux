package main

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/Azure/azure-docker-extension/pkg/vmextension"
	"github.com/Azure/custom-script-extension-linux/pkg/seqnum"
	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
)

type cmdFunc func(*log.Context, vmextension.HandlerEnvironment) error

type cmd struct {
	f                  cmdFunc // associated function
	name               string  // human readable string
	shouldReportStatus bool    // determines if running this should log to a .status file
}

var (
	cmdInstall   = cmd{install, "Install", false}
	cmdEnable    = cmd{enable, "Enable", true}
	cmdUninstall = cmd{uninstall, "Uninstall", false}

	cmds = map[string]cmd{
		"install":   cmdInstall,
		"uninstall": cmdUninstall,
		"enable":    cmdEnable,
		"update":    {noop, "Update", true},
		"disable":   {noop, "Disable", true},
	}
)

func noop(ctx *log.Context, h vmextension.HandlerEnvironment) error {
	ctx.Log("event", "noop")
	return nil
}

func install(ctx *log.Context, h vmextension.HandlerEnvironment) error {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return errors.Wrap(err, "failed to create data dir")
	}
	ctx.Log("event", "created data dir", "path", dataDir)
	ctx.Log("event", "installed")
	return nil
}

func uninstall(ctx *log.Context, h vmextension.HandlerEnvironment) error {
	{ // a new context scope with path
		ctx = ctx.With("path", dataDir)
		ctx.Log("event", "removing data dir", "path", dataDir)
		if err := os.RemoveAll(dataDir); err != nil {
			return errors.Wrap(err, "failed to delete data dir")
		}
		ctx.Log("event", "removed data dir")
	}
	ctx.Log("event", "uninstalled")
	return nil
}

func enable(ctx *log.Context, h vmextension.HandlerEnvironment) error {
	// parse the extension handler settings (not available prior to 'enable')
	cfg, err := parseAndValidateSettings(ctx, h.HandlerEnvironment.ConfigFolder)
	if err != nil {
		return errors.Wrap(err, "failed to get configuration")
	}

	// exit if this sequence number (a snapshot of the configuration) is alrady
	// processed. if not, save this sequence number before proceeding.
	seqNumPath := filepath.Join(dataDir, seqNumFile)
	if shouldExit, err := checkAndSaveSeqNum(ctx, h.SeqNo, seqNumPath); err != nil {
		return errors.Wrap(err, "failed to process ")
	} else if shouldExit {
		ctx.Log("event", "exit", "message", "this script configuration is already processed, will not run again")
		os.Exit(0) // exit immediately to prevent reporting '.status'
	}

	dir := filepath.Join(dataDir, downloadDir, h.SeqNo)
	if err := downloadFiles(ctx, dir, cfg); err != nil {
		return errors.Wrap(err, "processing file downloads failed")
	}

	if err := runCmd(ctx, dir, cfg); err != nil {
		return err
	}

	ctx.Log("event", "enabled")
	return nil
}

// checkAndSaveSeqNum checks if the given seqNum is already processed
// according to the specified seqNumFile and if so, returns true,
// otherwise saves the given seqNum into seqNumFile returns false.
func checkAndSaveSeqNum(ctx log.Logger, seqNum, seqNumFile string) (shouldExit bool, _ error) {
	ctx.Log("event", "comparing seqnum", "path", seqNumFile)
	seq, err := strconv.Atoi(seqNum)
	if err != nil {
		return false, errors.Wrapf(err, "failed to parse seqnum: %q", seqNum)
	}

	smaller, err := seqnum.IsSmallerThan(seqNumFile, seq)
	if err != nil {
		return false, errors.Wrap(err, "failed to check sequence number")
	}
	if !smaller {
		// stored sequence number is equals or greater than the current
		// sequence number.
		return true, nil
	}
	if err := seqnum.Set(seqNumFile, seq); err != nil {
		return false, errors.Wrap(err, "failed to save the sequence number")
	}
	ctx.Log("event", "seqnum saved", "path", seqNumFile)
	return false, nil
}

// downloadFiles downloads the files specified in cfg into dir (creates if does
// not exist) and takes storage credentials specified in cfg into account.
func downloadFiles(ctx *log.Context, dir string, cfg handlerSettings) error {
	// - prepare the output directory for files and the command output
	// - create the directory if missing
	ctx.Log("event", "creating output directory", "path", dir)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return errors.Wrap(err, "failed to prepare output directory")
	}
	ctx.Log("event", "created output directory")

	// - download files
	ctx.Log("files", len(cfg.FileURLs))
	for i, f := range cfg.FileURLs {
		ctx := ctx.With("file", i)
		ctx.Log("event", "download start")
		if err := downloadAndProcessURL(ctx, f, dir, cfg.StorageAccountName, cfg.StorageAccountKey); err != nil {
			ctx.Log("event", "download failed", "error", err)
			return errors.Wrapf(err, "failed to download file[%d]", i)
		}
		ctx.Log("event", "download complete", "output", dir)
	}
	return nil
}

// runCmd runs the command (extracted from cfg) in the given dir (assumed to exist).
func runCmd(ctx log.Logger, dir string, cfg handlerSettings) error {
	ctx.Log("event", "executing command", "output", dir)
	cmd := cfg.publicSettings.CommandToExecute
	if cmd == "" {
		cmd = cfg.protectedSettings.CommandToExecute
	}
	if err := ExecCmdInDir(cmd, dir); err != nil {
		ctx.Log("event", "failed to execute command", "error", err, "output", dir)
		return errors.Wrap(err, "failed to execute command")
	}
	ctx.Log("event", "executed command", "output", dir)
	return nil
}
