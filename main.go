package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Azure/azure-docker-extension/pkg/vmextension"
	"github.com/Azure/azure-docker-extension/pkg/vmextension/status"
	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
)

var (
	// dataDir is where we store the downloaded files, logs and state for
	// the extension handler
	dataDir = "/var/lib/azure/custom-script"

	// seqNumFile holds the processed highest sequence number to make
	// sure we do not run the command more than once for the same sequence
	// number. Stored under dataDir.
	seqNumFile = "seqnum"

	// downloadDir is where we store the downloaded files in the "{downloadDir}/{seqnum}/file"
	// format and the logs as "{downloadDir}/{seqnum}/std(out|err)". Stored under dataDir
	downloadDir = "download"
)

func main() {
	ctx := log.NewContext(log.NewSyncLogger(log.NewLogfmtLogger(
		os.Stdout))).With("time", log.DefaultTimestamp).With("version", VersionString())

	// parse command line arguments
	cmd := parseCmd(os.Args)
	ctx = ctx.With("operation", strings.ToLower(cmd.name))

	// parse extension environment
	hEnv, err := vmextension.GetHandlerEnv()
	if err != nil {
		ctx.Log("message", "failed to parse handlerenv", "error", err)
		os.Exit(1)
	}
	ctx = ctx.With("seq", hEnv.SeqNo)

	// execute the subcommand
	ctx.Log("event", "start")
	reportStatus(ctx, hEnv, status.StatusTransitioning, cmd, "")
	if err := cmd.f(ctx, hEnv); err != nil {
		ctx.Log("event", "failed to handle", "error", err)
		reportStatus(ctx, hEnv, status.StatusError, cmd, err.Error())
		os.Exit(1)
	}
	reportStatus(ctx, hEnv, status.StatusSuccess, cmd, "")
	ctx.Log("event", "end")
}

// parseCmd looks at os.Args and parses the subcommand. If it is invalid,
// it prints the usage string and an error message and exits with code 0.
func parseCmd(args []string) cmd {
	if len(os.Args) != 2 {
		printUsage(args)
		fmt.Println("Incorrect usage.")
		os.Exit(1)
	}
	op := os.Args[1]
	cmd, ok := cmds[op]
	if !ok {
		printUsage(args)
		fmt.Printf("Incorrect command: %q\n", op)
		os.Exit(1)
	}
	return cmd
}

// printUsage prints the help string and version of the program to stdout with a
// trailing new line.
func printUsage(args []string) {
	fmt.Printf("Usage: %s ", os.Args[0])
	i := 0
	for k := range cmds {
		fmt.Printf(k)
		if i != len(cmds)-1 {
			fmt.Printf("|")
		}
		i++
	}
	fmt.Println()
	fmt.Println(DetailedVersionString())
}

// reportStatus saves operation status to the status file for the extension
// handler with the optional given message, if the given cmd requires reporting
// status.
//
// If an error occurs reporting the status, it will be logged and returned.
func reportStatus(ctx *log.Context, hEnv vmextension.HandlerEnvironment, t status.Type, c cmd, msg string) error {
	if !c.shouldReportStatus {
		ctx.Log("status", "not reported for operation (by design)")
		return nil
	}
	s := status.NewStatus(t, c.name, statusMsg(c, t, msg))
	seq, _ := strconv.Atoi(hEnv.SeqNo)
	if err := s.Save(hEnv.HandlerEnvironment.StatusFolder, seq); err != nil {
		ctx.Log("event", "failed to save handler status", "error", err)
		return errors.Wrap(err, "failed to save handler status")
	}
	return nil
}

// statusMsg creates the reported status message based on the provided operation
// type and the given message string.
//
// A message will be generated for empty string. For error status, pass the
// error message.
func statusMsg(c cmd, t status.Type, msg string) string {
	s := c.name
	switch t {
	case status.StatusSuccess:
		s += " succeeded"
	case status.StatusTransitioning:
		s += " in progress"
	case status.StatusError:
		s += " failed"
	}

	if msg != "" {
		// append the original
		s += ": " + msg
	}
	return s
}
