// Package emoji provides embedded GitHub and Gitmoji emoji packs and lookup indices.
package emoji

import (
	"slices"
	"strings"
)

const emojiVariation = "\uFE0F"

// Emoji represents a single emoji definition with its shortcode, symbol, and optional description.
type Emoji struct {
	Code        string `json:"code"`
	Emoji       string `json:"emoji"`
	Description string `json:"description,omitempty"`
}

// Pack stores emojis indexed by shortcode and reverse-indexed by emoji symbol while preserving insertion order.
type Pack struct {
	codes  map[string]Emoji
	order  []string
	emojis map[string][]Emoji
}

// NewPack creates an empty initialized Pack.
func NewPack() *Pack {
	return &Pack{
		codes:  map[string]Emoji{},
		order:  []string{},
		emojis: map[string][]Emoji{},
	}
}

// Wrap ensures a code is enclosed in colons (e.g., "sparkles" -> ":sparkles:").
func Wrap(code string) string {
	if strings.HasPrefix(code, ":") && strings.HasSuffix(code, ":") && len(code) >= 2 {
		return code
	}

	bare := Unwrap(code)

	return ":" + bare + ":"
}

// Unwrap strips leading and trailing colons from a shortcode.
func Unwrap(code string) string {
	code = strings.TrimPrefix(code, ":")
	code = strings.TrimSuffix(code, ":")

	return code
}

// Add inserts or updates an emoji definition in the pack, preserving deterministic order.
func (p *Pack) Add(def Emoji) {
	def.Code = Unwrap(def.Code)
	if def.Code == "" {
		return
	}

	if prev, exists := p.codes[def.Code]; exists {
		p.removeReverseEntry(prev.Emoji, def.Code)
	} else {
		p.order = append(p.order, def.Code)
	}

	p.codes[def.Code] = def
	p.emojis[def.Emoji] = append(p.emojis[def.Emoji], def)
}

func (p *Pack) removeReverseEntry(symbol, code string) {
	entries := p.emojis[symbol]
	if len(entries) == 0 {
		return
	}

	filtered := slices.DeleteFunc(entries, func(e Emoji) bool {
		return e.Code == code
	})
	if len(filtered) == 0 {
		delete(p.emojis, symbol)
		return
	}

	p.emojis[symbol] = filtered
}

// Get retrieves an emoji definition by shortcode (with or without colons).
func (p *Pack) Get(code string) (Emoji, bool) {
	e, ok := p.codes[Unwrap(code)]
	return e, ok
}

// GetCodes returns all emoji definitions matching a Unicode emoji symbol.
func (p *Pack) GetCodes(symbol string) []Emoji {
	if entries, ok := p.emojis[symbol]; ok && len(entries) > 0 {
		return slices.Clone(entries)
	}

	if !strings.HasSuffix(symbol, emojiVariation) {
		if entries, ok := p.emojis[symbol+emojiVariation]; ok && len(entries) > 0 {
			return slices.Clone(entries)
		}
	} else {
		trimmed := strings.TrimSuffix(symbol, emojiVariation)
		if entries, ok := p.emojis[trimmed]; ok && len(entries) > 0 {
			return slices.Clone(entries)
		}
	}

	return nil
}

// GetCode returns the wrapped shortcode for the first definition matching symbol.
func (p *Pack) GetCode(symbol string) (string, bool) {
	entries := p.GetCodes(symbol)
	if len(entries) == 0 {
		return "", false
	}

	return Wrap(entries[0].Code), true
}

// All returns a copy of all emoji definitions in insertion order.
func (p *Pack) All() []Emoji {
	out := make([]Emoji, 0, len(p.order))
	for _, code := range p.order {
		if e, ok := p.codes[code]; ok {
			out = append(out, e)
		}
	}

	return out
}

// Len returns the number of registered emoji definitions.
func (p *Pack) Len() int {
	return len(p.codes)
}
