package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DO3B/gotmojify/internal/config"
)

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("defaults load expected types and devmojis", func(t *testing.T) {
		t.Parallel()

		is := assert.New(t)
		must := require.New(t)

		cfg, err := config.New(nil)
		must.NoError(err)

		is.Contains(cfg.Types, "feat")
		is.Contains(cfg.Types, "ci")
		is.Len(cfg.Types, 10)
		is.Equal(19, cfg.Pack.Len())

		feat, ok := cfg.Pack.Get("feat")
		must.True(ok)
		is.Equal("sparkles", feat.Emoji)
		is.Equal("a new feature", feat.Description)

		sec, ok := cfg.Pack.Get("security")
		must.True(ok)
		is.Equal("lock", sec.Emoji)
		is.Equal("Fix security or privacy issues.", sec.Description)
	})

	t.Run("overrides existing code preserving description", func(t *testing.T) {
		t.Parallel()

		is := assert.New(t)
		must := require.New(t)

		cfg, err := config.New(&config.Options{
			Types: []string{"lint", "feat"},
			Devmoji: []config.DevmojiDef{
				{Code: "feat", Emoji: "poop"},
				{Code: "fuckup", Emoji: "poop", Description: "something got messed up pretty bad"},
			},
		})
		must.NoError(err)

		feat, ok := cfg.Pack.Get("feat")
		must.True(ok)
		is.Equal("poop", feat.Emoji)
		is.Equal("a new feature", feat.Description)

		fuckup, ok := cfg.Pack.Get("fuckup")
		must.True(ok)
		is.Equal("poop", fuckup.Emoji)
		is.Equal("something got messed up pretty bad", fuckup.Description)

		is.Contains(cfg.Types, "lint")
		is.Len(cfg.Types, 11)
	})

	t.Run("missing code returns ErrMissingCode", func(t *testing.T) {
		t.Parallel()

		is := assert.New(t)
		_, err := config.New(&config.Options{
			Devmoji: []config.DevmojiDef{{Emoji: "poop"}},
		})
		is.ErrorIs(err, config.ErrMissingCode)
	})

	t.Run("missing emoji returns ErrMissingEmoji", func(t *testing.T) {
		t.Parallel()

		is := assert.New(t)
		_, err := config.New(&config.Options{
			Devmoji: []config.DevmojiDef{{Code: "foo"}},
		})
		is.ErrorIs(err, config.ErrMissingEmoji)
	})

	t.Run("config without devmoji succeeds", func(t *testing.T) {
		t.Parallel()

		is := assert.New(t)
		must := require.New(t)

		cfg, err := config.New(&config.Options{Types: []string{}})
		must.NoError(err)
		is.NotNil(cfg)
	})

	t.Run("invalid gitmoji returns ErrGitmojiNotFound", func(t *testing.T) {
		t.Parallel()

		is := assert.New(t)
		_, err := config.New(&config.Options{
			Devmoji: []config.DevmojiDef{{Code: "test", Gitmoji: "foobar"}},
		})
		is.ErrorIs(err, config.ErrGitmojiNotFound)
	})
}

func TestFindRoot(t *testing.T) {
	t.Parallel()

	is := assert.New(t)
	must := require.New(t)

	root := t.TempDir()
	nested := filepath.Join(root, "a", "b", "c")
	must.NoError(os.MkdirAll(nested, 0o750))
	must.NoError(os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/test\n"), 0o600))

	is.Equal(root, config.FindRoot("go.mod", nested))
	is.Empty(config.FindRoot("nonexistent-marker-file.xyz", nested))
}

func TestLoad(t *testing.T) {
	t.Parallel()

	t.Run("explicit yaml config file", func(t *testing.T) {
		t.Parallel()

		is := assert.New(t)
		must := require.New(t)

		dir := t.TempDir()
		cfgPath := filepath.Join(dir, "custom.yaml")
		yamlContent := `types:
  - lint
format: shortcode
devmoji:
  - code: fuckup
    emoji: poop
    description: something got messed up pretty bad
  - code: feat
    emoji: poop
`
		must.NoError(os.WriteFile(cfgPath, []byte(yamlContent), 0o600))

		cfg, err := config.Load(viper.New(), "custom.yaml", dir)
		must.NoError(err)

		is.Equal("shortcode", cfg.Format)
		is.Contains(cfg.Types, "lint")

		feat, ok := cfg.Pack.Get("feat")
		must.True(ok)
		is.Equal("poop", feat.Emoji)

		fuckup, ok := cfg.Pack.Get("fuckup")
		must.True(ok)
		is.Equal("poop", fuckup.Emoji)
	})

	t.Run("explicit missing file returns ErrNotFound", func(t *testing.T) {
		t.Parallel()

		is := assert.New(t)
		_, err := config.Load(viper.New(), "does-not-exist.yaml", t.TempDir())
		is.ErrorIs(err, config.ErrNotFound)
	})

	t.Run("auto-discovers config at parent git root", func(t *testing.T) {
		t.Parallel()

		is := assert.New(t)
		must := require.New(t)

		root := t.TempDir()
		nested := filepath.Join(root, "pkg", "sub")
		must.NoError(os.MkdirAll(filepath.Join(root, ".git"), 0o750))
		must.NoError(os.MkdirAll(nested, 0o750))

		jsonContent := `{
  "types": ["custom"],
  "devmoji": [
    {"code": "css", "gitmoji": "art"}
  ]
}`
		must.NoError(os.WriteFile(filepath.Join(root, ".gotmojify.json"), []byte(jsonContent), 0o600))

		cfg, err := config.Load(viper.New(), "", nested)
		must.NoError(err)

		is.Contains(cfg.Types, "custom")

		css, ok := cfg.Pack.Get("css")
		must.True(ok)
		is.Equal("art", css.Emoji)
		is.Equal("Improve structure / format of the code.", css.Description)
	})

	t.Run("falls back to defaults when no config exists", func(t *testing.T) {
		t.Parallel()

		is := assert.New(t)
		must := require.New(t)

		dir := t.TempDir()
		cfg, err := config.Load(viper.New(), "", dir)
		must.NoError(err)

		feat, ok := cfg.Pack.Get("feat")
		must.True(ok)
		is.Equal("sparkles", feat.Emoji)
	})
}
