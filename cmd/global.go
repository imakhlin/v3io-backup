package main

import (
	"context"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"os"
	"runtime"
	"strings"
	"syscall"
)

var version = "0.0.1"

// GlobalOptions hold all global options for v3io-backup tool.
type GlobalOptions struct {
	Repo     string
	Quiet    bool
	LogLevel string

	ctx    context.Context
	stdout io.Writer
	stderr io.Writer

	Options []string
}

var globalOptions = GlobalOptions{
	stdout: os.Stdout,
	stderr: os.Stderr,
}

func init() {
	var cancel context.CancelFunc
	globalOptions.ctx, cancel = context.WithCancel(context.Background())
	AddCleanupHandler(func() error {
		cancel()
		return nil
	})

	f := cmdRoot.GetCmd().PersistentFlags()
	f.StringVarP(&globalOptions.Repo, "repo", "r", os.Getenv("V3IO_REPOSITORY"), "repository to backup to or restore from (default: $V3IO_REPOSITORY)")
	f.BoolVarP(&globalOptions.Quiet, "quiet", "q", false, "do not output comprehensive progress report")
	f.StringVar(&globalOptions.LogLevel, "log-level", "v", "set log level. Valid values error|warn|info|debug.")
	f.StringSliceVarP(&globalOptions.Options, "option", "o", []string{}, "set extended option (`key=value`, can be specified multiple times)")

	restoreTerminal()
}

// checkErrno returns nil when err is set to syscall.Errno(0), since this is no
// error condition.
func checkErrno(err error) error {
	e, ok := err.(syscall.Errno)
	if !ok {
		return err
	}

	if e == 0 {
		return nil
	}

	return err
}

func stdinIsTerminal() bool {
	return terminal.IsTerminal(int(os.Stdin.Fd()))
}

func stdoutIsTerminal() bool {
	return terminal.IsTerminal(int(os.Stdout.Fd()))
}

func stdoutTerminalWidth() int {
	w, _, err := terminal.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 0
	}
	return w
}

// restoreTerminal installs a cleanup handler that restores the previous
// terminal state on exit.
func restoreTerminal() {
	if !stdoutIsTerminal() {
		return
	}

	fd := int(os.Stdout.Fd())
	state, err := terminal.GetState(fd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to get terminal state: %v\n", err)
		return
	}

	AddCleanupHandler(func() error {
		err := checkErrno(terminal.Restore(fd, state))
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to get restore terminal state: %#+v\n", err)
		}
		return err
	})
}

// ClearLine creates a platform dependent string to clear the current
// line, so it can be overwritten. ANSI sequences are not supported on
// current windows commands shell.
func ClearLine() string {
	if runtime.GOOS == "windows" {
		if w := stdoutTerminalWidth(); w > 0 {
			return strings.Repeat(" ", w-1) + "\r"
		}
		return ""
	}
	return "\x1b[2K"
}

// Printf writes the message to the configured stdout stream.
func Printf(format string, args ...interface{}) {
	_, err := fmt.Fprintf(globalOptions.stdout, format, args...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to write to stdout: %v\n", err)
	}
}

// PrintProgress wraps fmt.Printf to handle the difference in writing progress
// information to terminals and non-terminal stdout
func PrintProgress(format string, args ...interface{}) {
	var (
		message         string
		carriageControl string
	)
	message = fmt.Sprintf(format, args...)

	if !(strings.HasSuffix(message, "\r") || strings.HasSuffix(message, "\n")) {
		if stdoutIsTerminal() {
			carriageControl = "\r"
		} else {
			carriageControl = "\n"
		}
		message = fmt.Sprintf("%s%s", message, carriageControl)
	}

	if stdoutIsTerminal() {
		message = fmt.Sprintf("%s%s", ClearLine(), message)
	}

	fmt.Print(message)
}

// Warnf writes the message to the configured stderr stream.
func Warnf(format string, args ...interface{}) {
	_, err := fmt.Fprintf(globalOptions.stderr, format, args...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to write to stderr: %v\n", err)
	}
}

// Exitf uses Warnf to write the message and then terminates the process with
// the given exit code.
func Exitf(exitcode int, format string, args ...interface{}) {
	if format[len(format)-1] != '\n' {
		format += "\n"
	}

	Warnf(format, args...)
	Exit(exitcode)
}
