package commit_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DO3B/gotmojify/internal/commit"
	"github.com/DO3B/gotmojify/internal/config"
	"github.com/DO3B/gotmojify/internal/devmoji"
)

func newTestFormatter(t *testing.T) (*commit.Formatter, *devmoji.Converter) {
	t.Helper()

	cfg, err := config.New(nil)
	require.NoError(t, err)

	conv := devmoji.New(cfg)

	return commit.New(conv), conv
}

func TestFormatter_FormatCommit(t *testing.T) {
	t.Parallel()

	formatter, conv := newTestFormatter(t)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no scope commit msg",
			input:    "feat: testing",
			expected: "feat: ✨ testing",
		},
		{
			name:     "invalid type commit msg",
			input:    "foo: testing",
			expected: "foo: testing",
		},
		{
			name:     "type with scope commit msg",
			input:    "feat(cli): testing",
			expected: "feat(cli): ✨ testing",
		},
		{
			name:     "breaking changes",
			input:    "feat(cli)!: important",
			expected: "feat(cli)!: 💥 ✨ important",
		},
		{
			name:     "breaking changes multiline",
			input:    "feat(cli): important\nBREAKING CHANGE: foo bar",
			expected: "feat(cli): 💥 ✨ important\nBREAKING CHANGE: foo bar",
		},
		{
			name:     "type with scope release commit msg",
			input:    "chore(release): testing",
			expected: "chore(release): 🚀 testing",
		},
		{
			name:     "type with scope security commit msg",
			input:    "feat(security): testing",
			expected: conv.Emojify("feat(security): :sparkles: :security: testing"),
		},
		{
			name:     "invalid commit msg",
			input:    "invalid commit",
			expected: "invalid commit",
		},
		{
			name:     "type with existing devmoji",
			input:    "feat(cli): :test: testing",
			expected: "feat(cli): ✨ 🚨 testing",
		},
		{
			name:     "lowercase scope that includes numbers",
			input:    "feat(jira-123): testing",
			expected: "feat(jira-123): ✨ testing",
		},
		{
			name:     "uppercase scope that includes numbers",
			input:    "feat(JIRA-123): testing",
			expected: "feat(JIRA-123): ✨ testing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			is := assert.New(t)
			is.Equal(tt.expected, formatter.FormatCommit(tt.input, false))
		})
	}
}

func TestFormatter_FormatLog(t *testing.T) {
	t.Parallel()

	formatter, conv := newTestFormatter(t)

	t.Run("multiline log output", func(t *testing.T) {
		t.Parallel()

		is := assert.New(t)
		input := "bla feat(cli): testing\n" +
			//nolint:dupword // mirrors folke/devmoji test fixture
			"bla bla feat: :test: testing\n" +
			"feat(cli): :sparkles: foo\n" +
			"chore(release): deploy\n" +
			"eat(cli): :test: testing"
		expected := "bla feat(cli): ✨ testing\n" +
			//nolint:dupword // mirrors folke/devmoji test fixture
			"bla bla feat: ✨ 🚨 testing\n" +
			"feat(cli): ✨ foo\n" +
			"chore(release): 🚀 deploy\n" +
			"eat(cli): 🚨 testing"

		is.Equal(expected, formatter.FormatLog(input, false))
	})

	t.Run("already formatted git log line", func(t *testing.T) {
		t.Parallel()

		is := assert.New(t)
		text := "* 8f16492 - feat(cli): ✨ added cli for working with devmoji 🚀 (19 hours ago) <Folke Lemaitre>"

		is.Equal(text, formatter.FormatLog(text, false))
	})

	t.Run("colored log output includes ANSI codes", func(t *testing.T) {
		t.Parallel()

		is := assert.New(t)
		got := formatter.FormatLog("feat(cli): hello", true)
		is.Contains(got, "\x1b[")
		is.Contains(got, "✨ hello")
	})

	t.Run("multi log and commit idempotency", func(t *testing.T) {
		t.Parallel()

		is := assert.New(t)

		cases := []struct {
			input  string
			output string
		}{
			{"feat(foo): testing", "feat(foo): :sparkles: testing"},
			{"foo(security): testing", "foo(security): :security: testing"},
			{"feat(test): :art: testing", "feat(test): :sparkles: :test: :art: testing"},
			{"feat(test)!: :art: testing", "feat(test)!: :boom: :sparkles: :test: :art: testing"},
			{"feat(foo): :sparkles: :art: testing", "feat(foo): :sparkles: :art: testing"},
			{"feat(foo): :art: testing", "feat(foo): :sparkles: :art: testing"},
			{"fix(security): testing", "fix(security): :bug: :security: testing"},
			{"fix(docs): 🐛 📚 updated README.md", "fix(docs): 🐛 📚 updated README.md"},
			{"feat(test): testing", "feat(test): :sparkles: :test: testing"},
			{"chore(deps): testing", "chore(deps): :chore-deps: testing"},
		}

		for _, tc := range cases {
			expected := conv.Emojify(tc.output)
			is.Equal(expected, conv.Emojify(conv.Demojify(expected)))
			is.Equal(expected, formatter.FormatLog(tc.input, false))
			is.Equal(expected, formatter.FormatCommit(tc.input, false))
			is.Equal(conv.Emojify("not a header "+tc.input), formatter.FormatCommit("not a header "+tc.input, false))
			is.Equal(expected, formatter.FormatLog(formatter.FormatLog(tc.input, false), false))
			is.Equal(expected, formatter.FormatLog(formatter.FormatLog(formatter.FormatLog(tc.input, false), false), false))
		}
	})
}

func TestFormatter_Lint(t *testing.T) {
	t.Parallel()

	formatter, _ := newTestFormatter(t)

	validMessages := []string{
		"style: 🎨 prettier 2.0",
		"fixup! foo",
		"squash!",
		"revert something",
		"Merge branch 'master' of github.com:folke/devmoji",
		"Merge remote-tracking branch 'master' of github.com:folke/devmoji",
		"style: 🎨 Prettier 2.0",
		"chore(release): 2.1.8 [skip ci]",
		"fix(deps): update dependency chalk to v4 (#49)",
		"chore(deps): update all non-major dependencies (#48)",
		"chore(release): 2.1.7 [skip ci]",
		"fix: 🐛 🔒️ upgrade minimist which had a security vulnerability",
		"docs: fixed config example",
		"feat(TextInput): add new variant",
		"feat(JIRA-123): add new variant",
		"feat(jira-123): add new variant",
	}

	for _, msg := range validMessages {
		t.Run("valid_"+msg, func(t *testing.T) {
			t.Parallel()

			is := assert.New(t)
			is.Empty(formatter.Lint(msg, false))
		})
	}

	invalidMessages := []string{
		"Style: 🎨 prettier 2.0",
		"Merging branch 'master' of github.com:folke/devmoji",
		"style 🎨 Prettier 2.0",
		"chorerelease): 2.1.8 [skip ci]",
		"fix(deps) update dependency chalk to v4 (#49)",
		"chore(depS: update all non-major dependencies (#48)",
		"2.1.7 [skip ci]",
		"🔒️ upgrade minimist which had a security vulnerability",
		"update all non-major dependencies (#47)",
		"feat:    ",
	}

	for _, msg := range invalidMessages {
		t.Run("invalid_"+msg, func(t *testing.T) {
			t.Parallel()

			is := assert.New(t)
			errs := formatter.Lint(msg, false)
			is.NotEmpty(errs)
			is.Contains(errs[len(errs)-1], "conventionalcommits.org")
		})
	}
}

func FuzzFormatCommit(f *testing.F) {
	f.Add("feat(cli): add feature")
	f.Add("fix!: breaking bug\n\nBREAKING CHANGE: something")
	f.Add("invalid commit message")
	f.Add(":sparkles: already emojified")

	cfg, err := config.New(nil)
	if err != nil {
		f.Fatal(err)
	}

	formatter := commit.New(devmoji.New(cfg))

	f.Fuzz(func(t *testing.T, input string) {
		first := formatter.FormatCommit(input, false)

		second := formatter.FormatCommit(first, false)
		if first != second {
			t.Fatalf("FormatCommit not idempotent on %q: first=%q second=%q", input, first, second)
		}
	})
}
