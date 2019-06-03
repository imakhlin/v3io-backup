package main

import (
	"fmt"
	"os"

	//"github.org/imakhlin/v3io-backup/pkg/commands"
	"v3io-backup/pkg/commands"
)

var cmdRoot, err = commands.NewCmdRoot()

func main() {
	if err := Run(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func Run() error {
	if err != nil || cmdRoot == nil {
		fmt.Fprintf(os.Stderr, "ERROR: Unable to initialize command line interface.\nDetails: %v", err)
		return err
	}

	defer tearDown(cmdRoot)
	return cmdRoot.Execute()
}

func tearDown(cmd *commands.CmdRoot) {
	if cmd.Reporter != nil { // could be nil if has failed on initialisation
		cmd.Reporter.Stop()
	}
}
