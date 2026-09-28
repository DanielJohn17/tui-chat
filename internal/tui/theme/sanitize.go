package theme

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	// oscRegex matches Operating System Command sequences: ESC ] ... (BEL | ESC \)
	oscRegex = regexp.MustCompile(`\x1b\][^\x07\x1b]*(?:\x07|\x1b\\|$)`)

	// csiRegex matches Control Sequence Introducer sequences: ESC [ ... [@-~]
	csiRegex = regexp.MustCompile(`\x1b\[[0-9:;<=>?]*[ -/]*[@-~]`)

	// otherEscapeRegex matches other escape sequences like DCS, APC, PM, or 2-char escapes
	otherEscapeRegex = regexp.MustCompile(`\x1b(?:[PX^_][^\x1b]*(?:\x1b\\|\x07|$)|[@-Z\\-_])`)
)

// StripANSI removes all ANSI escape sequences (CSI, OSC, APC, etc.) from a string.
func StripANSI(s string) string {
	s = oscRegex.ReplaceAllString(s, "")
	s = csiRegex.ReplaceAllString(s, "")
	s = otherEscapeRegex.ReplaceAllString(s, "")
	return s
}

// SanitizeText strips all ANSI escape sequences and dangerous control characters,
// while preserving newline and tab formatting.
func SanitizeText(s string) string {
	s = StripANSI(s)
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r == '\n' || r == '\t' {
			b.WriteRune(r)
		} else if !unicode.IsControl(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// SanitizeLine strips all ANSI escape sequences and control characters,
// replacing newlines/tabs with spaces and collapsing adjacent whitespace.
// Ideal for usernames, channel/chat names, message previews, and status bar text.
func SanitizeLine(s string) string {
	s = StripANSI(s)
	var b strings.Builder
	b.Grow(len(s))
	lastSpace := false
	for _, r := range s {
		if r == '\r' || r == '\n' || r == '\t' || r == ' ' {
			if !lastSpace && b.Len() > 0 {
				b.WriteByte(' ')
				lastSpace = true
			}
		} else if !unicode.IsControl(r) {
			b.WriteRune(r)
			lastSpace = false
		}
	}
	return strings.TrimSpace(b.String())
}
