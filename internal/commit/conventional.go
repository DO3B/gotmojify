// Package commit implements Conventional Commits formatting and linting with devmoji insertion.
package commit

import (
	"regexp"
	"slices"
	"strings"

	"github.com/DO3B/gotmojify/internal/devmoji"
	"github.com/DO3B/gotmojify/internal/style"
)

var (
	commitRegex        = regexp.MustCompile(`(?im)(?P<type>:?[a-z-]+)(?:\((?P<scope>[a-z0-9-]+)\))?(?P<breaking>!)?:\s*(?:(?P<other>(?::[a-z-]+:\s*)+)\s*)?`)
	lintRegex          = regexp.MustCompile(`(?i)^(?P<type>:?[a-z-]+)(?:\((?P<scope>[a-z0-9-]+)\))?(?P<breaking>!)?:[\s]+(?P<description>.*)`)
	breakingDescRegex  = regexp.MustCompile(`(?m)^\s*BREAKING CHANGE`)
	mergeRegex         = regexp.MustCompile(`^Merge (remote-tracking )?branch`)
	fixupRegex         = regexp.MustCompile(`^(fixup|squash)!`)
	revertRegex        = regexp.MustCompile(`^[rR]evert`)
	trailingSpaceRegex = regexp.MustCompile(`\s*$`)
)

// Formatter formats and lints Conventional Commit messages.
type Formatter struct {
	converter *devmoji.Converter
}

// New creates a Formatter using the given devmoji Converter.
func New(converter *devmoji.Converter) *Formatter {
	return &Formatter{converter: converter}
}

// FormatCommit formats a conventional commit header at the start of text and emojifies the message.
func (f *Formatter) FormatCommit(text string, colorEnabled bool) string {
	return f.format(text, true, colorEnabled)
}

// FormatLog formats all conventional commit headers anywhere in text (such as git log output).
func (f *Formatter) FormatLog(text string, colorEnabled bool) string {
	return f.format(text, false, colorEnabled)
}

// FormatEmoji resolves the emoji sequence for a conventional commit's type, scope, extra codes, and breaking flag.
func (f *Formatter) FormatEmoji(commitType, scope, other string, isBreaking bool) string {
	typeEmoji, scopeEmoji := f.resolveTypeScopeEmojis(commitType, scope)

	inputs := make([]string, 0, 4)
	if isBreaking {
		inputs = append(inputs, "boom")
	}

	inputs = appendNonEmpty(inputs, typeEmoji, scopeEmoji)
	inputs = appendOtherCodes(inputs, other)

	return strings.Join(f.dedupeResolved(inputs), " ")
}

func (f *Formatter) resolveTypeScopeEmojis(commitType, scope string) (string, string) {
	pack := f.converter.Config().Pack

	var typeEmoji, scopeEmoji string

	if def, ok := pack.Get(commitType); ok {
		typeEmoji = def.Emoji
	}

	if scope == "" {
		return typeEmoji, ""
	}

	if def, ok := pack.Get(commitType + "-" + scope); ok {
		return def.Emoji, ""
	}

	if def, ok := pack.Get(scope); ok {
		scopeEmoji = def.Emoji
	}

	return typeEmoji, scopeEmoji
}

func appendNonEmpty(dst []string, values ...string) []string {
	for _, val := range values {
		if val != "" {
			dst = append(dst, val)
		}
	}

	return dst
}

func appendOtherCodes(dst []string, other string) []string {
	if other == "" {
		return dst
	}

	matches := devmoji.ShortcodeRegex.FindAllStringSubmatch(other, -1)
	for _, m := range matches {
		if len(m) > 1 {
			dst = append(dst, m[1])
		}
	}

	return dst
}

func (f *Formatter) dedupeResolved(codes []string) []string {
	resolved := make([]string, 0, len(codes))
	for _, code := range codes {
		if code == "" {
			continue
		}

		sym := f.converter.Resolve(code)
		if !slices.Contains(resolved, sym) {
			resolved = append(resolved, sym)
		}
	}

	return resolved
}

func (f *Formatter) format(text string, firstOnly, colorEnabled bool) string {
	text = f.converter.Devmojify(text)
	hasBreakingBody := breakingDescRegex.MatchString(text)
	st := style.New(colorEnabled)

	matches := commitRegex.FindAllStringSubmatchIndex(text, -1)
	if len(matches) == 0 {
		return f.converter.Emojify(text)
	}

	var out strings.Builder

	lastEnd := 0

	for _, idx := range matches {
		matchStart, matchEnd := idx[0], idx[1]
		out.WriteString(text[lastEnd:matchStart])

		match := text[matchStart:matchEnd]
		replacement := f.formatMatch(
			match,
			extractGroup(text, idx, 1),
			extractGroup(text, idx, 2),
			extractGroup(text, idx, 3),
			extractGroup(text, idx, 4),
			matchStart,
			firstOnly,
			hasBreakingBody,
			st,
		)
		out.WriteString(replacement)

		lastEnd = matchEnd
	}

	out.WriteString(text[lastEnd:])

	return f.converter.Emojify(out.String())
}

func extractGroup(text string, idx []int, group int) string {
	start, end := idx[group*2], idx[group*2+1]
	if start < 0 || end < 0 {
		return ""
	}

	return text[start:end]
}

func (f *Formatter) formatMatch(
	match string,
	commitType string,
	scope string,
	breaking string,
	other string,
	offset int,
	firstOnly bool,
	hasBreakingBody bool,
	st *style.Styler,
) string {
	if firstOnly && offset > 0 {
		return match
	}

	if strings.HasPrefix(commitType, ":") {
		return match
	}

	emojiStr := f.FormatEmoji(commitType, scope, other, breaking != "" || hasBreakingBody)
	if emojiStr == "" {
		return match
	}

	var header strings.Builder
	header.WriteString(st.Blue(commitType))

	if scope != "" {
		header.WriteString(st.BoldBlue("(" + scope + ")"))
	}

	if breaking != "" {
		header.WriteString(st.Blue("!"))
	}

	header.WriteString(st.Blue(":"))
	header.WriteString(" ")
	header.WriteString(emojiStr)

	if wsLoc := trailingSpaceRegex.FindStringIndex(match); wsLoc != nil && wsLoc[0] > 0 {
		header.WriteString(match[wsLoc[0]:])
	}

	return header.String()
}

// Lint validates a commit message header against Conventional Commits rules.
func (f *Formatter) Lint(text string, colorEnabled bool) []string {
	firstLine, _, _ := strings.Cut(text, "\n")
	firstLine = strings.TrimSuffix(firstLine, "\r")

	isExempt := mergeRegex.MatchString(firstLine) ||
		fixupRegex.MatchString(firstLine) ||
		revertRegex.MatchString(firstLine)
	if isExempt {
		return []string{}
	}

	st := style.New(colorEnabled)
	errs := make([]string, 0)
	types := f.converter.Config().Types

	match := lintRegex.FindStringSubmatch(firstLine)
	if match != nil {
		commitType := match[1]
		description := match[4]

		if !slices.Contains(types, commitType) {
			errs = append(errs, "Type should be one of: "+st.Grey(strings.Join(types, ", ")))
		}

		if strings.TrimSpace(description) == "" {
			errs = append(errs, "Missing description")
		}
	} else {
		errs = append(errs, "Expecting a commit message like:")
		errs = append(errs, "  "+st.Blue("type")+st.BoldBlue("(scope):")+st.Dim(" description"))
	}

	if len(errs) > 0 {
		errs = append(errs, "Get help at https://www.conventionalcommits.org/")
	}

	return errs
}
