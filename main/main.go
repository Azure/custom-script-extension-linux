package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/Azure/azure-docker-extension/pkg/vmextension"
	"github.com/Azure/azure-docker-extension/pkg/vmextension/status"
	"github.com/go-kit/kit/log"
)

var (
	// dataDir is where we store the downloaded files, logs and state for
	// the extension handler
	dataDir = "/var/lib/waagent/custom-script"

	// seqNumFile holds the processed highest sequence number to make
	// sure we do not run the command more than once for the same sequence
	// number. Stored under dataDir.
	seqNumFile = "seqnum"

	// most recent sequence, which was previously traced by seqNumFile. This was
	// incorrect. The correct way is mrseq.  This file is auto-preserved by the agent.
	mostRecentSequence = "mrseq"

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
		os.Exit(cmd.failExitCode)
	}
	seqNum, err := vmextension.FindSeqNumConfig(hEnv.HandlerEnvironment.ConfigFolder)
	if err != nil {
		ctx.Log("messsage", "failed to find sequence number", "error", err)
	}
	ctx = ctx.With("seq", seqNum)

	// check sub-command preconditions, if any, before executing
	ctx.Log("event", "start")
	if cmd.pre != nil {
		ctx.Log("event", "pre-check")
		if err := cmd.pre(ctx, seqNum); err != nil {
			ctx.Log("event", "pre-check failed", "error", err)
			os.Exit(cmd.failExitCode)
		}
	}
	// execute the subcommand
	reportStatus(ctx, hEnv, seqNum, status.StatusTransitioning, cmd, "")
	msg, err := cmd.f(ctx, hEnv, seqNum)
	if err != nil {
		ctx.Log("event", "failed to handle", "error", err)
		reportStatus(ctx, hEnv, seqNum, status.StatusError, cmd, err.Error()+msg)
		os.Exit(cmd.failExitCode)
	}
	reportStatus(ctx, hEnv, seqNum, status.StatusSuccess, cmd, msg)
	ctx.Log("event", "end")
}

// parseCmd looks at os.Args and parses the subcommand. If it is invalid,
// it prints the usage string and an error message and exits with code 0.
func parseCmd(args []string) cmd {
	if len(os.Args) != 2 {
		printUsage(args)
		fmt.Println("Incorrect usage.")
		os.Exit(2)
	}
	op := os.Args[1]
	cmd, ok := cmds[op]
	if !ok {
		printUsage(args)
		fmt.Printf("Incorrect command: %q\n", op)
		os.Exit(2)
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
