// Command gotmojify formats and lints Conventional Commits with devmoji emojis.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/DO3B/gotmojify/internal/cli"
	"github.com/DO3B/gotmojify/internal/style"
)

// Build-time metadata populated via -ldflags.
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	os.Exit(run())
}

func run() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	info := cli.BuildInfo{
		Version: version,
		Commit:  commit,
		Date:    date,
	}

	if err := cli.ExecuteContext(ctx, info); err != nil {
		if exitErr, ok := errors.AsType[*cli.ExitError](err); ok {
			if !exitErr.IsSilent {
				printError(err)
			}

			return exitErr.Code
		}

		printError(err)

		return 1
	}

	return 0
}

func printError(err error) {
	_, noColor := os.LookupEnv("NO_COLOR")
	fi, statErr := os.Stderr.Stat()
	isTTY := statErr == nil && fi.Mode()&os.ModeCharDevice != 0
	st := style.New(isTTY && !noColor)

	_, _ = fmt.Fprintf(os.Stderr, "%s %v\n", st.Red("error"), err)
}
