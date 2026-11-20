package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/DO3B/gotmojify/internal/commit"
	"github.com/DO3B/gotmojify/internal/config"
	"github.com/DO3B/gotmojify/internal/devmoji"
	"github.com/DO3B/gotmojify/internal/emoji"
	"github.com/DO3B/gotmojify/internal/style"
)

const maxStdinLineSize = 1024 * 1024

// Option configures optional dependencies on the root command (primarily for isolated tests).
type Option func(*rootOptions)

type rootOptions struct {
	workDir      string
	hasStdinTTY  bool
	isStdinTTY   bool
	hasStdoutTTY bool
	isStdoutTTY  bool
}

// WithWorkDir sets the working directory used for config and .git discovery.
func WithWorkDir(dir string) Option {
	return func(o *rootOptions) {
		o.workDir = dir
	}
}

// WithStdinTTY overrides TTY detection on stdin for deterministic testing.
func WithStdinTTY(isTTY bool) Option {
	return func(o *rootOptions) {
		o.hasStdinTTY = true
		o.isStdinTTY = isTTY
	}
}

// WithStdoutTTY overrides TTY detection on stdout for deterministic testing.
func WithStdoutTTY(isTTY bool) Option {
	return func(o *rootOptions) {
		o.hasStdoutTTY = true
		o.isStdoutTTY = isTTY
	}
}

type appState struct {
	viper      *viper.Viper
	opts       rootOptions
	cfgFile    string
	listFlag   bool
	textFlag   string
	lintFlag   bool
	formatFlag string
	commitFlag bool
	noCommit   bool
	editFlag   bool
	logFlag    bool
	colorFlag  bool
	noColor    bool

	cfg       *config.Config
	converter *devmoji.Converter
	formatter *commit.Formatter
}

// NewRootCmd builds an isolated Cobra command tree for gotmojify.
func NewRootCmd(info BuildInfo, opts ...Option) *cobra.Command {
	normalized := normalizeBuildInfo(info)
	state := &appState{
		viper: viper.New(),
	}

	for _, opt := range opts {
		opt(&state.opts)
	}

	rootCmd := &cobra.Command{
		Use:           "gotmojify",
		Short:         "Add color and gitmoji-inspired emojis to Conventional Commits",
		Long:          "gotmojify formats, lints, and colorizes Conventional Commits and text using devmoji, GitHub shortcodes, and Unicode emojis.",
		Version:       normalized.Version,
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return state.initConfig(cmd)
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return state.run(cmd)
		},
	}

	bindRootFlags(rootCmd, state)
	rootCmd.AddCommand(newVersionCmd(normalized))

	return rootCmd
}

// ExecuteContext runs the gotmojify CLI using the provided context and build metadata.
func ExecuteContext(ctx context.Context, info BuildInfo) error {
	rootCmd := NewRootCmd(info)

	return rootCmd.ExecuteContext(ctx)
}

func bindRootFlags(cmd *cobra.Command, s *appState) {
	pflags := cmd.PersistentFlags()
	pflags.StringVarP(&s.cfgFile, "config", "c", "", "location of config file (.gotmojify.{yaml,json,toml} or devmoji.config.*)")

	flags := cmd.Flags()
	flags.BoolVarP(&s.listFlag, "list", "l", false, "list all known devmojis")
	flags.StringVarP(&s.textFlag, "text", "t", "", "text to format. reads from stdin when omitted")
	flags.BoolVar(&s.lintFlag, "lint", false, "lint the conventional commit. disabled for --log")
	flags.StringVarP(&s.formatFlag, "format", "f", string(devmoji.FormatUnicode), "output format: unicode, shortcode, devmoji, strip")
	flags.BoolVar(&s.commitFlag, "commit", true, "automatically add a devmoji to the conventional commit header")
	flags.BoolVar(&s.noCommit, "no-commit", false, "do not process conventional commit headers")
	flags.BoolVarP(&s.editFlag, "edit", "e", false, "read last commit message from .git/COMMIT_EDITMSG in the git root")
	flags.BoolVar(&s.logFlag, "log", false, "format conventional commits in text similar to git log")
	flags.BoolVar(&s.colorFlag, "color", false, "use colors for formatting (enabled by default for TTY output)")
	flags.BoolVar(&s.noColor, "no-color", false, "don't use colors")

	cmd.MarkFlagsMutuallyExclusive("commit", "no-commit")
	cmd.MarkFlagsMutuallyExclusive("color", "no-color")

	_ = cmd.RegisterFlagCompletionFunc("format", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return devmoji.ValidFormats(), cobra.ShellCompDirectiveNoFileComp
	})

	s.viper.SetEnvPrefix("GOTMOJIFY")
	s.viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	s.viper.AutomaticEnv()

	_ = s.viper.BindPFlag("format", flags.Lookup("format"))
	_ = s.viper.BindPFlag("commit", flags.Lookup("commit"))
	_ = s.viper.BindPFlag("lint", flags.Lookup("lint"))
	_ = s.viper.BindPFlag("log", flags.Lookup("log"))
	_ = s.viper.BindPFlag("color", flags.Lookup("color"))
}

func (s *appState) initConfig(cmd *cobra.Command) error {
	if cmd.Name() == "version" || cmd.Name() == "help" || cmd.Name() == "completion" {
		return nil
	}

	workDir := s.opts.workDir
	if workDir == "" {
		var err error

		workDir, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("getting working directory: %w", err)
		}

		s.opts.workDir = workDir
	}

	cfg, err := config.Load(s.viper, s.cfgFile, workDir)
	if err != nil {
		return err
	}

	s.cfg = cfg
	s.converter = devmoji.New(cfg)
	s.formatter = commit.New(s.converter)

	return nil
}

func (s *appState) run(cmd *cobra.Command) error {
	if s.listFlag {
		return s.runList(cmd)
	}

	format, err := s.resolveFormat(cmd)
	if err != nil {
		return err
	}

	processCommit := s.resolveCommit(cmd)
	processLog := s.logFlag || s.viper.GetBool("log")
	lintEnabled := s.lintFlag || s.viper.GetBool("lint")
	colorEnabled := s.resolveColor(cmd, false)

	if cmd.Flags().Changed("text") {
		out, err := s.formatText(cmd, s.textFlag, format, processCommit, processLog, lintEnabled, colorEnabled)
		if err != nil {
			return err
		}

		_, err = fmt.Fprintln(cmd.OutOrStdout(), out)

		return err
	}

	if s.editFlag {
		return s.runEdit(cmd, format, processCommit, lintEnabled)
	}

	if !s.stdinIsTTY(cmd) {
		return s.runStdin(cmd, format, processCommit, processLog, lintEnabled, colorEnabled)
	}

	if err := cmd.Help(); err != nil {
		return err
	}

	return newSilentExitError(1)
}

func (s *appState) resolveFormat(cmd *cobra.Command) (devmoji.Format, error) {
	raw := s.formatFlag
	if !cmd.Flags().Changed("format") {
		if envVal, ok := os.LookupEnv("GOTMOJIFY_FORMAT"); ok && envVal != "" {
			raw = envVal
		} else if s.cfg.Format != "" {
			raw = s.cfg.Format
		}
	}

	return devmoji.ParseFormat(raw)
}

func (s *appState) resolveCommit(cmd *cobra.Command) bool {
	if s.noCommit {
		return false
	}

	if cmd.Flags().Changed("commit") {
		return s.commitFlag
	}

	if _, ok := os.LookupEnv("GOTMOJIFY_COMMIT"); ok {
		return s.viper.GetBool("commit")
	}

	return s.commitFlag
}

func (s *appState) resolveColor(cmd *cobra.Command, defaultWhenUnset bool) bool {
	if s.noColor {
		return false
	}

	if cmd.Flags().Changed("color") {
		return s.colorFlag
	}

	if _, disabled := os.LookupEnv("NO_COLOR"); disabled {
		return false
	}

	if _, ok := os.LookupEnv("GOTMOJIFY_COLOR"); ok {
		return s.viper.GetBool("color")
	}

	if defaultWhenUnset {
		return true
	}

	return s.stdoutIsTTY(cmd)
}

func (s *appState) stdinIsTTY(cmd *cobra.Command) bool {
	if s.opts.hasStdinTTY {
		return s.opts.isStdinTTY
	}

	if cmd.InOrStdin() != os.Stdin {
		return false
	}

	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}

	return fi.Mode()&os.ModeCharDevice != 0
}

func (s *appState) stdoutIsTTY(cmd *cobra.Command) bool {
	if s.opts.hasStdoutTTY {
		return s.opts.isStdoutTTY
	}

	if cmd.OutOrStdout() != os.Stdout {
		return false
	}

	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}

	return fi.Mode()&os.ModeCharDevice != 0
}

func (s *appState) formatText(
	cmd *cobra.Command,
	text string,
	format devmoji.Format,
	processCommit bool,
	processLog bool,
	lintEnabled bool,
	colorEnabled bool,
) (string, error) {
	if processCommit && lintEnabled && !processLog {
		errs := s.formatter.Lint(text, s.resolveColor(cmd, true))
		if len(errs) > 0 {
			st := style.New(s.resolveColor(cmd, true))
			for _, e := range errs {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "%s %s\n", st.Red("✖"), e)
			}

			return "", newSilentExitError(1)
		}
	}

	switch {
	case processLog:
		text = s.formatter.FormatLog(text, colorEnabled)
	case processCommit:
		text = s.formatter.FormatCommit(text, colorEnabled)
	}

	return s.converter.ApplyFormat(text, format)
}

func (s *appState) runList(cmd *cobra.Command) error {
	st := style.New(s.resolveColor(cmd, true))
	out := cmd.OutOrStdout()

	if _, err := fmt.Fprintln(out, st.BlueDimUnderline("Available Devmoji")); err != nil {
		return err
	}

	for _, item := range s.cfg.Pack.All() {
		var cc string
		if slices.Contains(s.cfg.Types, item.Code) {
			cc = item.Code + ": "
		}

		if prefix, scope, ok := strings.Cut(item.Code, "-"); ok && slices.Contains(s.cfg.Types, prefix) {
			cc = fmt.Sprintf("%s(%s): ", prefix, scope)
		}

		paddedCode := fmt.Sprintf("%-15s", emoji.Wrap(item.Code))
		if _, err := fmt.Fprintf(
			out,
			"%s  %s %s%s\n",
			s.converter.Resolve(item.Emoji),
			st.Blue(paddedCode),
			st.Green(cc),
			item.Description,
		); err != nil {
			return err
		}
	}

	return nil
}

func (s *appState) runEdit(
	cmd *cobra.Command,
	format devmoji.Format,
	processCommit bool,
	lintEnabled bool,
) error {
	gitDir, err := FindGitDir(s.opts.workDir)
	if err != nil {
		return fmt.Errorf("couldn't find .git/COMMIT_EDITMSG: %w", err)
	}

	raw, err := ReadCommitEditMsg(gitDir)
	if err != nil {
		return fmt.Errorf("couldn't find .git/COMMIT_EDITMSG: %w", err)
	}

	formatted, err := s.formatText(cmd, raw, format, processCommit, false, lintEnabled, false)
	if err != nil {
		return err
	}

	displayColor := s.resolveColor(cmd, true)

	preview, err := s.formatText(cmd, formatted, format, processCommit, false, false, displayColor)
	if err != nil {
		return err
	}

	if err := WriteCommitEditMsg(gitDir, formatted); err != nil {
		return err
	}

	st := style.New(displayColor)
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", st.Green("✔"), preview)

	return err
}

func (s *appState) runStdin(
	cmd *cobra.Command,
	format devmoji.Format,
	processCommit bool,
	processLog bool,
	lintEnabled bool,
	colorEnabled bool,
) error {
	scanner := bufio.NewScanner(cmd.InOrStdin())
	scanner.Buffer(make([]byte, 0, 64*1024), maxStdinLineSize)

	firstLine := true
	ctx := cmd.Context()

	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return err
		}

		line := scanner.Text()

		formatted, err := s.formatText(
			cmd,
			line,
			format,
			processCommit && firstLine,
			processLog,
			lintEnabled,
			colorEnabled,
		)
		if err != nil {
			return err
		}

		if _, err := fmt.Fprintln(cmd.OutOrStdout(), formatted); err != nil {
			return err
		}

		firstLine = false
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading stdin: %w", err)
	}

	return nil
}
