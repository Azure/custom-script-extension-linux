package main

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/Azure/azure-docker-extension/pkg/vmextension"
	"github.com/Azure/custom-script-extension-linux/seqnum"
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

func enable(ctx *log.Context, h vmextension.HandlerEnvironment) error {
	// parse the extension handler settings (not available prior to 'enable')
	cfg, err := parseAndValidateSettings(ctx, h.HandlerEnvironment.ConfigFolder)
	if err != nil {
		return errors.Wrap(err, "failed to get configuration")
	}

	seq, err := strconv.Atoi(h.SeqNo)
	if err != nil {
		return errors.Wrapf(err, "failed to parse seqnum: %q", h.SeqNo)
	}

	// exit if this sequence number (a snapshot of the configuration) is
	// processed. if not, save this sequence number before proceeding.
	sf := filepath.Join(dataDir, seqNumFile)
	ctx.Log("event", "comparing seqnum")
	if b, err := seqnum.IsSmallerOrEqualThan(sf, seq); err != nil {
		return errors.Wrap(err, "failed to check sequence number")
	} else if b {
		ctx.Log("event", "exit", "message", "this configuration is already processed, will not run again")
		os.Exit(0) // exit immediately to prevent reporting '.status'
	}
	if err := seqnum.Set(sf, seq); err != nil {
		return errors.Wrap(err, "failed to save sequence number")
	}
	ctx.Log("event", "seqnum saved", "path", sf)

	// - prepare the output directory for files and the command output
	dir := filepath.Join(dataDir, downloadDir, h.SeqNo)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return errors.Wrap(err, "failed to prepare output directory")
	}
	ctx.Log("event", "created output directory", "path", dir)

	// - download files
	ctx.Log("files", len(cfg.FileURLs))
	for i, f := range cfg.FileURLs {
		ctx := ctx.With("file", i)
		ctx.Log("event", "download start")
		if err := downloadAndProcessURL(f, dir, cfg.StorageAccountName, cfg.StorageAccountKey); err != nil {
			ctx.Log("event", "download failed", "error", err)
			return errors.Wrapf(err, "failed to download file[%d]", i)
		}
		ctx.Log("event", "download complete", "output", dir)
	}

	// - run the script
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
	ctx.Log("event", "enabled")
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
