package cli

import (
	"cmp"
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// BuildInfo carries build-time version metadata injected via ldflags.
type BuildInfo struct {
	Version string
	Commit  string
	Date    string
}

func normalizeBuildInfo(info BuildInfo) BuildInfo {
	return BuildInfo{
		Version: cmp.Or(info.Version, "dev"),
		Commit:  cmp.Or(info.Commit, "unknown"),
		Date:    cmp.Or(info.Date, "unknown"),
	}
}

func newVersionCmd(info BuildInfo) *cobra.Command {
	normalized := normalizeBuildInfo(info)

	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := fmt.Fprintf(
				cmd.OutOrStdout(),
				"gotmojify %s (commit: %s, built: %s)\n",
				normalized.Version,
				normalized.Commit,
				normalized.Date,
			)
			if err != nil {
				return fmt.Errorf("writing version: %w", err)
			}

			if bi, ok := debug.ReadBuildInfo(); ok {
				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "go: %s\n", bi.GoVersion); err != nil {
					return fmt.Errorf("writing go version: %w", err)
				}
			}

			return nil
		},
	}
}
