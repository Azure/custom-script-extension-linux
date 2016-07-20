package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		printUsage()
		fmt.Println("Incorrect usage.")
		os.Exit(1)
	}
	op := os.Args[1]
	cmd, ok := cmds[op]
	if !ok {
		printUsage()
		fmt.Printf("Incorrect command: %q\n", op)
		os.Exit(1)
	}

	fmt.Printf("operation: %s\n", cmd.name)
}

func printUsage() {
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
}
