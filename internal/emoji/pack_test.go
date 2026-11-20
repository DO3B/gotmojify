package emoji_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DO3B/gotmojify/internal/emoji"
)

const (
	bareSparkles    = "sparkles"
	wrappedSparkles = ":sparkles:"
)

func TestWrapAndUnwrap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		wantWrap string
		wantBare string
	}{
		{
			name:     "bare code",
			input:    bareSparkles,
			wantWrap: wrappedSparkles,
			wantBare: bareSparkles,
		},
		{
			name:     "already wrapped code",
			input:    wrappedSparkles,
			wantWrap: wrappedSparkles,
			wantBare: bareSparkles,
		},
		{
			name:     "leading colon only",
			input:    ":sparkles",
			wantWrap: wrappedSparkles,
			wantBare: bareSparkles,
		},
		{
			name:     "trailing colon only",
			input:    "sparkles:",
			wantWrap: wrappedSparkles,
			wantBare: bareSparkles,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			is := assert.New(t)
			is.Equal(tt.wantWrap, emoji.Wrap(tt.input))
			is.Equal(tt.wantBare, emoji.Unwrap(tt.input))
			is.Equal(tt.wantWrap, emoji.Wrap(emoji.Unwrap(tt.input)))
		})
	}
}

func TestPack_AddUpdateAndOrder(t *testing.T) {
	t.Parallel()

	is := assert.New(t)
	must := require.New(t)

	pack := emoji.NewPack()
	pack.Add(emoji.Emoji{Code: ":feat:", Emoji: "sparkles", Description: "feature"})
	pack.Add(emoji.Emoji{Code: "fix", Emoji: "bug", Description: "bugfix"})
	pack.Add(emoji.Emoji{Code: "feat", Emoji: "boom", Description: "updated feature"})

	is.Equal(2, pack.Len())

	all := pack.All()
	must.Len(all, 2)
	is.Equal("feat", all[0].Code)
	is.Equal("boom", all[0].Emoji)
	is.Equal("updated feature", all[0].Description)
	is.Equal("fix", all[1].Code)

	// Stale reverse mapping for "sparkles" should be removed on override.
	is.Empty(pack.GetCodes("sparkles"))

	codes := pack.GetCodes("boom")
	must.Len(codes, 1)
	is.Equal("feat", codes[0].Code)
}

func TestPack_VariationSelectorLookup(t *testing.T) {
	t.Parallel()

	is := assert.New(t)

	pack := emoji.NewPack()
	pack.Add(emoji.Emoji{Code: "zap", Emoji: "⚡\uFE0F"})
	pack.Add(emoji.Emoji{Code: bareSparkles, Emoji: "✨"})

	code, ok := pack.GetCode("⚡")
	is.True(ok)
	is.Equal(":zap:", code)

	code, ok = pack.GetCode("✨\uFE0F")
	is.True(ok)
	is.Equal(wrappedSparkles, code)
}

func TestGitHubAndGitmojiPacks(t *testing.T) {
	t.Parallel()

	t.Run("all gitmoji entries resolve in github pack", func(t *testing.T) {
		t.Parallel()

		is := assert.New(t)
		must := require.New(t)

		gm := emoji.Gitmoji()
		gh := emoji.GitHub()

		is.Greater(gh.Len(), 1400)
		is.Greater(gm.Len(), 60)

		for _, item := range gm.All() {
			shortcode := emoji.Wrap(item.Code)
			name := emoji.Unwrap(shortcode)

			is.Equal(shortcode, emoji.Wrap(shortcode))
			is.Equal(name, emoji.Unwrap(name))

			found, ok := gh.Get(shortcode)
			must.True(ok, "expected gitmoji %s in github pack", shortcode)
			is.NotEmpty(found.Emoji)

			reverseCode, ok := gh.GetCode(found.Emoji)
			is.True(ok, "expected reverse lookup for %s (%s)", shortcode, found.Emoji)
			is.NotEmpty(reverseCode)
		}
	})

	t.Run("unknown codes return false", func(t *testing.T) {
		t.Parallel()

		is := assert.New(t)
		gh := emoji.GitHub()

		_, ok := gh.Get("missing")
		is.False(ok)
		_, ok = gh.Get(":missing:")
		is.False(ok)
	})

	t.Run("gitmoji art exists", func(t *testing.T) {
		t.Parallel()

		is := assert.New(t)
		art, ok := emoji.Gitmoji().Get("art")
		is.True(ok)
		is.NotEmpty(art.Description)
	})
}
