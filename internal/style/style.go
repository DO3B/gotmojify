// Package style provides race-free terminal color styling helpers.
package style

import "github.com/fatih/color"

// Styler formats strings with optional ANSI color codes without relying on global state.
type Styler struct {
	isEnabled        bool
	blue             *color.Color
	boldBlue         *color.Color
	blueDimUnderline *color.Color
	green            *color.Color
	red              *color.Color
	grey             *color.Color
	dim              *color.Color
}

// New creates a Styler with ANSI styling enabled or disabled.
func New(enabled bool) *Styler {
	return &Styler{
		isEnabled:        enabled,
		blue:             newColor(enabled, color.FgBlue),
		boldBlue:         newColor(enabled, color.FgBlue, color.Bold),
		blueDimUnderline: newColor(enabled, color.FgBlue, color.Faint, color.Underline),
		green:            newColor(enabled, color.FgGreen),
		red:              newColor(enabled, color.FgRed),
		grey:             newColor(enabled, color.FgHiBlack),
		dim:              newColor(enabled, color.Faint),
	}
}

func newColor(enabled bool, attrs ...color.Attribute) *color.Color {
	c := color.New(attrs...)
	if enabled {
		c.EnableColor()
	} else {
		c.DisableColor()
	}

	return c
}

// IsEnabled reports whether ANSI color output is active.
func (s *Styler) IsEnabled() bool {
	return s.isEnabled
}

// Blue renders text in blue.
func (s *Styler) Blue(text string) string {
	return s.blue.Sprint(text)
}

// BoldBlue renders text in bold blue.
func (s *Styler) BoldBlue(text string) string {
	return s.boldBlue.Sprint(text)
}

// BlueDimUnderline renders text in faint underlined blue.
func (s *Styler) BlueDimUnderline(text string) string {
	return s.blueDimUnderline.Sprint(text)
}

// Green renders text in green.
func (s *Styler) Green(text string) string {
	return s.green.Sprint(text)
}

// Red renders text in red.
func (s *Styler) Red(text string) string {
	return s.red.Sprint(text)
}

// Grey renders text in grey.
func (s *Styler) Grey(text string) string {
	return s.grey.Sprint(text)
}

// Dim renders text in faint style.
func (s *Styler) Dim(text string) string {
	return s.dim.Sprint(text)
}
