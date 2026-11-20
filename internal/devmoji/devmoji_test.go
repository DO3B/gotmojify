package devmoji_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DO3B/gotmojify/internal/config"
	"github.com/DO3B/gotmojify/internal/devmoji"
)

const (
	missingCode = ":missing:"
	plainSample = "testing 123"
)

type transformCase struct {
	name     string
	input    string
	expected string
}

func runTransformCases(t *testing.T, fn func(string) string, cases []transformCase) {
	t.Helper()

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			is := assert.New(t)
			is.Equal(tt.expected, fn(tt.input))
		})
	}
}

func newTestConverter(t *testing.T) *devmoji.Converter {
	t.Helper()

	cfg, err := config.New(nil)
	require.NoError(t, err)

	return devmoji.New(cfg)
}

func TestConverter_Emojify(t *testing.T) {
	t.Parallel()

	conv := newTestConverter(t)
	runTransformCases(t, conv.Emojify, []transformCase{
		{name: "github shortcode", input: ":rocket:", expected: "🚀"},
		{name: "devmoji code", input: ":feat:", expected: "✨"},
		{name: "unknown shortcode", input: missingCode, expected: missingCode},
		{name: "plain text", input: plainSample, expected: plainSample},
	})
}

func TestConverter_Strip(t *testing.T) {
	t.Parallel()

	conv := newTestConverter(t)
	runTransformCases(t, conv.Strip, []transformCase{
		{name: "unicode emoji in commit", input: "build: 📦 added", expected: "build: added"},
		{name: "shortcode only", input: missingCode, expected: ""},
		{name: "trailing devmoji", input: "testing 123 :feat:", expected: plainSample},
		{name: "multiple emojis", input: "test 🚀 :boom: :sparkles: :security:", expected: "test"},
	})
}

func TestConverter_Demojify(t *testing.T) {
	t.Parallel()

	conv := newTestConverter(t)
	runTransformCases(t, conv.Demojify, []transformCase{
		{name: "unicode rocket", input: "🚀", expected: ":rocket:"},
		{name: "devmoji test shortcode", input: ":test:", expected: ":rotating_light:"},
		{name: "unknown shortcode kept", input: missingCode, expected: missingCode},
		{name: "plain text", input: plainSample, expected: plainSample},
		{name: "mixed input", input: "This is a :test: of the first :release: :boom: ✨", expected: "This is a :rotating_light: of the first :rocket: :boom: :sparkles:"},
	})
}

func TestConverter_Devmojify(t *testing.T) {
	t.Parallel()

	conv := newTestConverter(t)
	runTransformCases(t, conv.Devmojify, []transformCase{
		{name: "rocket shortcode", input: ":rocket:", expected: ":chore-release:"},
		{name: "sparkles shortcode", input: ":sparkles:", expected: ":feat:"},
		{name: "rocket unicode", input: "🚀", expected: ":chore-release:"},
		{name: "plain text", input: plainSample, expected: plainSample},
		{name: "combined input", input: "🚀 :boom: :sparkles:", expected: ":chore-release: :breaking: :feat:"},
	})
}

func TestConverter_ResolveCycleProtection(t *testing.T) {
	t.Parallel()

	is := assert.New(t)
	must := require.New(t)

	cfg, err := config.New(&config.Options{
		Devmoji: []config.DevmojiDef{
			{Code: "art", Emoji: "art"},
		},
	})
	must.NoError(err)

	conv := devmoji.New(cfg)
	is.Equal("🎨", conv.Resolve("art"))
}

func TestParseAndApplyFormat(t *testing.T) {
	t.Parallel()

	conv := newTestConverter(t)

	tests := []struct {
		name     string
		format   string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "unicode format",
			format:   "unicode",
			input:    "This is a :test: of the first :release: :boom: :sparkles:",
			expected: "This is a 🚨 of the first 🚀 💥 ✨",
		},
		{
			name:     "shortcode format",
			format:   "shortcode",
			input:    ":sparkles: 🚀",
			expected: ":sparkles: :rocket:",
		},
		{
			name:     "devmoji format",
			format:   "devmoji",
			input:    "🚀 :boom: :sparkles:",
			expected: ":chore-release: :breaking: :feat:",
		},
		{
			name:     "strip format",
			format:   "strip",
			input:    "test 🚀 :boom: :sparkles: :security:",
			expected: "test",
		},
		{
			name:    "invalid format",
			format:  "xml",
			input:   "hello",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			is := assert.New(t)
			must := require.New(t)

			parsed, err := devmoji.ParseFormat(tt.format)
			if tt.wantErr {
				is.ErrorIs(err, devmoji.ErrInvalidFormat)
				return
			}

			must.NoError(err)

			got, err := conv.ApplyFormat(tt.input, parsed)
			must.NoError(err)
			is.Equal(tt.expected, got)
		})
	}
}

func BenchmarkEmojify(b *testing.B) {
	cfg, err := config.New(nil)
	if err != nil {
		b.Fatal(err)
	}

	conv := devmoji.New(cfg)
	sample := "feat(cli): :sparkles: added :rocket: support for :test: and :security:"

	for b.Loop() {
		_ = conv.Emojify(sample)
	}
}
