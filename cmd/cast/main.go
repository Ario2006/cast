package main

import (
	"os"

	"github.com/aryankumar/cast/internal/cli"
)

// version is overridden by release builds with -ldflags.
var version = "dev"

func main() {
	cmd := cli.NewRootCommand(cli.Options{Version: version, Out: os.Stdout, ErrOut: os.Stderr})
	if err := cmd.Execute(); err != nil {
		cli.PrintError(os.Stderr, err)
		os.Exit(1)
	}
}
