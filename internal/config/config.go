// Package config handles default settings, validation, and discovery of gotmojify/devmoji configuration files.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/spf13/viper"

	"github.com/DO3B/gotmojify/internal/emoji"
)

var (
	// ErrNotFound indicates an explicitly requested config file does not exist.
	ErrNotFound = errors.New("config: file not found")
	// ErrMissingCode indicates a custom devmoji entry is missing its required code field.
	ErrMissingCode = errors.New("config: devmoji code is missing")
	// ErrMissingEmoji indicates a custom devmoji entry lacks both emoji and gitmoji fields.
	ErrMissingEmoji = errors.New("config: missing emoji or gitmoji")
	// ErrGitmojiNotFound indicates a referenced gitmoji code does not exist in the built-in pack.
	ErrGitmojiNotFound = errors.New("config: gitmoji not found")
)

var configCandidates = []string{
	".gotmojify.yaml",
	".gotmojify.yml",
	".gotmojify.json",
	".gotmojify.toml",
	"gotmojify.yaml",
	"gotmojify.yml",
	"gotmojify.json",
	"gotmojify.toml",
	"devmoji.config.yaml",
	"devmoji.config.yml",
	"devmoji.config.json",
	"devmoji.config.toml",
}

// DevmojiDef defines a custom or overridden devmoji mapping in configuration.
type DevmojiDef struct {
	Code        string `json:"code" mapstructure:"code" toml:"code" yaml:"code"`
	Emoji       string `json:"emoji,omitempty" mapstructure:"emoji" toml:"emoji,omitempty" yaml:"emoji,omitempty"`
	Gitmoji     string `json:"gitmoji,omitempty" mapstructure:"gitmoji" toml:"gitmoji,omitempty" yaml:"gitmoji,omitempty"`
	Description string `json:"description,omitempty" mapstructure:"description" toml:"description,omitempty" yaml:"description,omitempty"`
}

// Options represents user-configurable gotmojify settings.
type Options struct {
	Types   []string     `json:"types,omitempty" mapstructure:"types" toml:"types,omitempty" yaml:"types,omitempty"`
	Devmoji []DevmojiDef `json:"devmoji,omitempty" mapstructure:"devmoji" toml:"devmoji,omitempty" yaml:"devmoji,omitempty"`
	Format  string       `json:"format,omitempty" mapstructure:"format" toml:"format,omitempty" yaml:"format,omitempty"`
}

// Config holds the resolved configuration options and compiled emoji pack.
type Config struct {
	Types   []string
	Devmoji []DevmojiDef
	Format  string
	Pack    *emoji.Pack
}

// Defaults returns the built-in conventional commit types and default devmojis matching folke/devmoji.
func Defaults() Options {
	return Options{
		Types: []string{
			"feat",
			"fix",
			"docs",
			"style",
			"refactor",
			"perf",
			"test",
			"chore",
			"build",
			"ci",
		},
		Devmoji: []DevmojiDef{
			{Code: "feat", Description: "a new feature", Emoji: "sparkles"},
			{Code: "fix", Description: "a bug fix", Emoji: "bug"},
			{Code: "docs", Description: "documentation only changes", Emoji: "books"},
			{
				Code:        "style",
				Description: "changes that do not affect the meaning of the code (white-space, formatting, missing semi-colons, etc)",
				Emoji:       "art",
			},
			{
				Code:        "refactor",
				Description: "a code change that neither fixes a bug nor adds a feature",
				Emoji:       "recycle",
			},
			{
				Code:        "perf",
				Description: "a code change that improves performance",
				Emoji:       "zap",
			},
			{
				Code:        "test",
				Description: "adding missing or correcting existing tests",
				Emoji:       "rotating_light",
			},
			{
				Code:        "chore",
				Description: "changes to the build process or auxiliary tools and libraries such as documentation generation",
				Emoji:       "wrench",
			},
			{
				Code:        "chore-release",
				Description: "code deployment or publishing to external repositories",
				Emoji:       "rocket",
			},
			{
				Code:        "chore-deps",
				Description: "add or delete dependencies",
				Emoji:       "link",
			},
			{
				Code:        "build",
				Description: "changes related to build processes",
				Emoji:       "package",
			},
			{
				Code:        "ci",
				Description: "updates to the continuous integration system",
				Emoji:       "construction_worker",
			},
			{
				Code:        "release",
				Description: "code deployment or publishing to external repositories",
				Emoji:       "rocket",
			},
			{Code: "security", Gitmoji: "lock"},
			{Code: "i18n", Gitmoji: "globe_with_meridians"},
			{Code: "breaking", Gitmoji: "boom"},
			{Code: "config", Gitmoji: "wrench", Emoji: "gear"},
			{Code: "add", Emoji: "heavy_plus_sign", Description: "add something"},
			{Code: "remove", Emoji: "heavy_minus_sign", Description: "remove something"},
		},
	}
}

// New creates a Config initialized with built-in defaults merged with optional overrides.
func New(opts *Options) (*Config, error) {
	cfg := &Config{
		Types:   []string{},
		Devmoji: []DevmojiDef{},
		Pack:    emoji.NewPack(),
	}

	if err := cfg.merge(Defaults()); err != nil {
		return nil, fmt.Errorf("loading defaults: %w", err)
	}

	if opts != nil {
		if opts.Format != "" {
			cfg.Format = opts.Format
		}

		if err := cfg.merge(*opts); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

func (c *Config) merge(opts Options) error {
	for _, t := range opts.Types {
		if t != "" && !slices.Contains(c.Types, t) {
			c.Types = append(c.Types, t)
		}
	}

	gitmojis := emoji.Gitmoji()

	for _, def := range opts.Devmoji {
		if def.Code == "" {
			return ErrMissingCode
		}

		if def.Gitmoji != "" {
			gm, ok := gitmojis.Get(def.Gitmoji)
			if !ok {
				return fmt.Errorf("%w: %q", ErrGitmojiNotFound, def.Gitmoji)
			}

			if def.Emoji == "" {
				def.Emoji = def.Gitmoji
			}

			if def.Description == "" {
				def.Description = gm.Description
			}
		}

		if def.Emoji == "" {
			return fmt.Errorf("%w: for code %q", ErrMissingEmoji, def.Code)
		}

		existing, hasExisting := c.Pack.Get(def.Code)

		desc := def.Description
		if desc == "" && hasExisting {
			desc = existing.Description
		}

		c.Pack.Add(emoji.Emoji{
			Code:        def.Code,
			Emoji:       def.Emoji,
			Description: desc,
		})
		c.Devmoji = append(c.Devmoji, def)
	}

	return nil
}

// FindRoot walks upward from cwd to locate the nearest directory containing target.
func FindRoot(target, cwd string) string {
	dir := filepath.Clean(cwd)
	for {
		candidate := filepath.Join(dir, target)
		if _, err := os.Stat(candidate); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}

		dir = parent
	}
}

// Load discovers and reads configuration using the provided isolated Viper instance.
func Load(v *viper.Viper, configFile, cwd string) (*Config, error) {
	if v == nil {
		v = viper.New()
	}

	if cwd == "" {
		var err error

		cwd, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("getting working directory: %w", err)
		}
	}

	if err := configureSource(v, configFile, cwd); err != nil {
		return nil, err
	}

	if v.ConfigFileUsed() != "" {
		if err := v.ReadInConfig(); err != nil {
			if _, ok := errors.AsType[viper.ConfigFileNotFoundError](err); !ok {
				return nil, fmt.Errorf("reading config: %w", err)
			}
		}
	}

	var opts Options
	if err := v.Unmarshal(&opts); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return New(&opts)
}

func configureSource(v *viper.Viper, configFile, cwd string) error {
	if configFile == "" {
		if discovered := discoverConfigFile(cwd); discovered != "" {
			v.SetConfigFile(discovered)
		}

		return nil
	}

	resolved := resolveConfigPath(configFile, cwd)
	if _, err := os.Stat(resolved); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("%w: %s", ErrNotFound, configFile)
		}

		return fmt.Errorf("checking config file %s: %w", configFile, err)
	}

	v.SetConfigFile(resolved)

	return nil
}

func resolveConfigPath(configFile, cwd string) string {
	if filepath.IsAbs(configFile) {
		return filepath.Clean(configFile)
	}

	return filepath.Join(cwd, configFile)
}

func discoverConfigFile(cwd string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = ""
	}

	rawDirs := []string{
		cwd,
		FindRoot("go.mod", cwd),
		FindRoot("package.json", cwd),
		FindRoot(".git", cwd),
		home,
	}

	seen := make(map[string]struct{}, len(rawDirs))
	for _, dir := range rawDirs {
		if dir == "" {
			continue
		}

		clean := filepath.Clean(dir)
		if _, ok := seen[clean]; ok {
			continue
		}

		seen[clean] = struct{}{}

		for _, name := range configCandidates {
			candidate := filepath.Join(clean, name)

			info, statErr := os.Stat(candidate)
			if statErr == nil && !info.IsDir() {
				return candidate
			}
		}
	}

	return ""
}
