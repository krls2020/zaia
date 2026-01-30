package main

import (
	"os"

	"github.com/zeropsio/zaia/internal/commands"
)

func main() {
	rootCmd := commands.NewRoot()
	if err := commands.Execute(rootCmd); err != nil {
		exitCode := 1
		if exitCoder, ok := err.(interface{ ExitCode() int }); ok {
			exitCode = exitCoder.ExitCode()
		}
		os.Exit(exitCode)
	}
}
