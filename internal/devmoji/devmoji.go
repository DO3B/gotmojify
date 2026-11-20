// Package devmoji implements core conversion between unicode emojis, GitHub shortcodes, devmoji codes, and stripped text.
package devmoji

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/DO3B/gotmojify/internal/config"
	"github.com/DO3B/gotmojify/internal/emoji"
)

// Format represents an output emoji representation format.
type Format string

const (
	// FormatUnknown is the uninitialized zero-value format sentinel.
	FormatUnknown Format = ""
	// FormatUnicode converts shortcodes and devmoji codes to Unicode emojis.
	FormatUnicode Format = "unicode"
	// FormatShortcode converts Unicode emojis and devmoji codes to GitHub shortcodes.
	FormatShortcode Format = "shortcode"
	// FormatDevmoji converts Unicode emojis and GitHub shortcodes to devmoji codes where mapped.
	FormatDevmoji Format = "devmoji"
	// FormatStrip removes all emojis and shortcodes from text.
	FormatStrip Format = "strip"
)

// ErrInvalidFormat indicates an unrecognized output format string.
var ErrInvalidFormat = errors.New("devmoji: invalid format")

var (
	// ShortcodeRegex matches colon-delimited emoji shortcodes such as :feat: or :+1:.
	ShortcodeRegex      = regexp.MustCompile(`:([a-zA-Z0-9_\-+]+):`)
	shortcodeSpaceRegex = regexp.MustCompile(`\s?:([a-zA-Z0-9_\-+]+):`)
	unicodeRegex        = regexp.MustCompile(`(?:[\x{00A9}\x{00AE}]|[\x{2000}-\x{3300}]|[\x{1F000}-\x{1FBFF}])\x{FE0F}?(?:\x{200D}(?:[\x{00A9}\x{00AE}]|[\x{2000}-\x{3300}]|[\x{1F000}-\x{1FBFF}])\x{FE0F}?)*`)
)

// ParseFormat validates and converts a raw format string to a Format value.
func ParseFormat(raw string) (Format, error) {
	switch Format(raw) {
	case FormatUnicode:
		return FormatUnicode, nil
	case FormatShortcode:
		return FormatShortcode, nil
	case FormatDevmoji:
		return FormatDevmoji, nil
	case FormatStrip:
		return FormatStrip, nil
	case FormatUnknown:
		return FormatUnknown, fmt.Errorf("%w: %q", ErrInvalidFormat, raw)
	default:
		return FormatUnknown, fmt.Errorf("%w: %q", ErrInvalidFormat, raw)
	}
}

// ValidFormats returns the supported output format names in display order.
func ValidFormats() []string {
	return []string{
		string(FormatUnicode),
		string(FormatShortcode),
		string(FormatDevmoji),
		string(FormatStrip),
	}
}

// Converter transforms text across emoji formats using the configured devmoji pack.
type Converter struct {
	config *config.Config
}

// New creates a Converter backed by cfg.
func New(cfg *config.Config) *Converter {
	return &Converter{config: cfg}
}

// Config returns the configuration backing the converter.
func (c *Converter) Config() *config.Config {
	return c.config
}

// Resolve resolves a devmoji or GitHub shortcode to its Unicode symbol, or returns a wrapped shortcode if unknown.
func (c *Converter) Resolve(code string) string {
	visited := map[string]struct{}{}
	curr := emoji.Unwrap(code)

	for {
		if _, seen := visited[curr]; seen {
			break
		}

		visited[curr] = struct{}{}

		if def, ok := c.config.Pack.Get(curr); ok {
			curr = emoji.Unwrap(def.Emoji)
			continue
		}

		break
	}

	if gh, ok := emoji.GitHub().Get(curr); ok {
		return gh.Emoji
	}

	return emoji.Wrap(curr)
}

// Emojify converts all Unicode emojis and shortcodes in text to Unicode emojis.
func (c *Converter) Emojify(text string) string {
	normalized := c.Demojify(text)

	return ShortcodeRegex.ReplaceAllStringFunc(normalized, func(match string) string {
		return c.Resolve(match)
	})
}

// Demojify converts recognized devmoji codes and Unicode emojis in text to GitHub shortcodes.
func (c *Converter) Demojify(text string) string {
	gh := emoji.GitHub()

	text = unicodeRegex.ReplaceAllStringFunc(text, func(symbol string) string {
		if code, ok := gh.GetCode(symbol); ok {
			return code
		}

		return symbol
	})

	return ShortcodeRegex.ReplaceAllStringFunc(text, func(match string) string {
		if _, ok := c.config.Pack.Get(match); !ok {
			return match
		}

		resolved := c.Resolve(match)
		if code, ok := gh.GetCode(resolved); ok {
			return code
		}

		return match
	})
}

// Devmojify converts emojis and GitHub shortcodes in text to matching devmoji shortcodes when possible.
func (c *Converter) Devmojify(text string) string {
	normalized := c.Demojify(text)
	gh := emoji.GitHub()
	gm := emoji.Gitmoji()

	return ShortcodeRegex.ReplaceAllStringFunc(normalized, func(match string) string {
		item, ok := gh.Get(match)
		if !ok {
			return match
		}

		candidates := make([]emoji.Emoji, 0)
		candidates = append(candidates, gm.GetCodes(item.Emoji)...)
		candidates = append(candidates, gh.GetCodes(item.Emoji)...)

		for _, candidate := range candidates {
			if code, found := c.config.Pack.GetCode(candidate.Code); found {
				return code
			}
		}

		return match
	})
}

// Strip removes all emojis and shortcodes (along with a single preceding space if present) from text.
func (c *Converter) Strip(text string) string {
	normalized := c.Demojify(text)

	return shortcodeSpaceRegex.ReplaceAllString(normalized, "")
}

// ApplyFormat converts text into the requested output Format.
func (c *Converter) ApplyFormat(text string, format Format) (string, error) {
	switch format {
	case FormatUnicode:
		return c.Emojify(text), nil
	case FormatShortcode:
		return c.Demojify(text), nil
	case FormatDevmoji:
		return c.Devmojify(text), nil
	case FormatStrip:
		return c.Strip(text), nil
	case FormatUnknown:
		return "", fmt.Errorf("%w: %q", ErrInvalidFormat, format)
	default:
		return "", fmt.Errorf("%w: %q", ErrInvalidFormat, format)
	}
}
