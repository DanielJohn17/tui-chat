package theme

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Cyberpunk Dark Mode OLED Palette
var (
	ColorBlack     = lipgloss.Color("#000000")
	ColorDarkBg    = lipgloss.Color("#0B0C10")
	ColorPanelBg   = lipgloss.Color("#12131C")
	ColorBorderDim = lipgloss.Color("#2E3047")
	ColorDimText   = lipgloss.Color("#5E627D")
	ColorWhite     = lipgloss.Color("#F4F4F8")
	ColorMagenta   = lipgloss.Color("#FF007F")
	ColorCyan      = lipgloss.Color("#00F0FF")
	ColorGreen     = lipgloss.Color("#00FF85")
	ColorYellow    = lipgloss.Color("#FFE600")
	ColorRed       = lipgloss.Color("#FF3366")
)

// Typography & Text Styles
var (
	StyleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorMagenta)

	StyleSubtitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorCyan)

	StyleDim = lipgloss.NewStyle().
			Foreground(ColorDimText)

	StyleSuccess = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorGreen)

	StyleWarning = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorYellow)

	StyleError = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorRed)

	StyleBadge = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorBlack).
			Background(ColorMagenta).
			Padding(0, 1)

	StyleKeyBadge = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorBlack).
			Background(ColorCyan).
			Padding(0, 1)

	StyleShortcut = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorYellow)
)

// Container & Panel Styles
var (
	StyleBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorderDim).
			Padding(0, 1)

	StyleBoxFocused = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorCyan).
			Padding(0, 1)

	StyleModal = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorMagenta).
			Padding(1, 2)

	StyleStatusBar = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorderDim).
			Padding(0, 1)

	StyleInputPrompt = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorCyan)
)

// Helper for text truncation using visual cell width to preserve UTF-8 runes and emojis
func Truncate(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= maxLen {
		return s
	}
	if maxLen <= 1 {
		return "…"
	}

	targetWidth := maxLen - 1
	var curWidth int
	var runes []rune
	for _, r := range s {
		rw := lipgloss.Width(string(r))
		if curWidth+rw > targetWidth {
			break
		}
		curWidth += rw
		runes = append(runes, r)
	}
	return string(runes) + "…"
}

// RenderLineTalkWordmark returns a compact two-tone colored wordmark: LINE in Magenta, TALK in Cyan
func RenderLineTalkWordmark() string {
	lineStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorMagenta)
	talkStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorCyan)
	return lineStyle.Render("LINE") + talkStyle.Render("TALK")
}

// RenderLineTalkLogo renders the 3-line ASCII block wordmark centered for card headers
func RenderLineTalkLogo(containerWidth int) string {
	lineStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorMagenta)
	talkStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorCyan)
	dimStyle := lipgloss.NewStyle().Foreground(ColorDimText)

	r1 := lineStyle.Render("█   █ █▄ █ ████") + "   " + talkStyle.Render("▀█▀ ███ █   █ █")
	r2 := lineStyle.Render("█   █ █ ▀█ █▄▄ ") + "    " + talkStyle.Render("█  █▄█ █   ██ ")
	r3 := lineStyle.Render("███ █ █  █ ████") + "    " + talkStyle.Render("█  █ █ ███ █ █")

	center := func(s string) string {
		if containerWidth > 0 {
			return lipgloss.NewStyle().Width(containerWidth).Align(lipgloss.Center).Render(s)
		}
		return s
	}

	sub := dimStyle.Render("SECURE TERMINAL COMMUNICATION")

	return lipgloss.JoinVertical(lipgloss.Center,
		center(r1),
		center(r2),
		center(r3),
		center(sub),
	)
}

// GetBackgroundSeq returns the ANSI escape sequence for a background color
// matching the current terminal color profile.
func GetBackgroundSeq(c lipgloss.TerminalColor) string {
	rendered := lipgloss.NewStyle().Background(c).Render("X")
	idx := strings.Index(rendered, "X")
	if idx > 0 {
		return rendered[:idx]
	}
	return ""
}

// isResetSequence checks if an SGR ANSI parameter string resets the background color
// without immediately setting a new custom background.
func isResetSequence(params string) bool {
	if params == "" || params == "0" || params == "49" {
		return true
	}

	parts := strings.Split(params, ";")
	hasReset := false
	hasNewBg := false

	for i := 0; i < len(parts); i++ {
		p := parts[i]
		switch p {
		case "0", "":
			hasReset = true
			hasNewBg = false
		case "49":
			hasReset = true
			hasNewBg = false
		case "38":
			// 38;5;n or 38;2;r;g;b (Foreground color)
			// Skip parameters belonging to 38 so color component numbers are not misparsed!
			if i+1 < len(parts) {
				if parts[i+1] == "5" {
					i += 2
				} else if parts[i+1] == "2" {
					i += 4
				}
			}
		case "48":
			// 48;5;n or 48;2;r;g;b (Background color)
			hasNewBg = true
			if i+1 < len(parts) {
				if parts[i+1] == "5" {
					i += 2
				} else if parts[i+1] == "2" {
					i += 4
				}
			}
		case "40", "41", "42", "43", "44", "45", "46", "47",
			"100", "101", "102", "103", "104", "105", "106", "107":
			hasNewBg = true
		}
	}

	return hasReset && !hasNewBg
}

// applyBgToLine applies bgSeq at the beginning of the line and restores it after any reset sequence.
func applyBgToLine(line string, bgSeq string) string {
	if bgSeq == "" {
		return line
	}

	var b strings.Builder
	b.WriteString(bgSeq)

	i := 0
	n := len(line)
	for i < n {
		if line[i] == '\x1b' && i+1 < n && line[i+1] == '[' {
			start := i
			i += 2
			for i < n && (line[i] >= 0x20 && line[i] <= 0x3F) {
				i++
			}
			if i < n {
				finalByte := line[i]
				i++
				seq := line[start:i]
				b.WriteString(seq)

				if finalByte == 'm' {
					params := line[start+2 : i-1]
					if isResetSequence(params) {
						b.WriteString(bgSeq)
					}
				}
			}
		} else {
			b.WriteByte(line[i])
			i++
		}
	}

	return b.String()
}

// EnforceDefaultBackground wraps the rendered view so that every character cell and line
// is painted with ColorDarkBg, ensuring readability regardless of terminal background theme.
func EnforceDefaultBackground(view string, width, height int) string {
	bgSeq := GetBackgroundSeq(ColorDarkBg)
	if bgSeq == "" {
		return view
	}

	lines := strings.Split(view, "\n")
	var result strings.Builder

	for i, line := range lines {
		if i > 0 {
			result.WriteByte('\n')
		}

		lineWidth := lipgloss.Width(line)
		processed := applyBgToLine(line, bgSeq)
		result.WriteString(processed)

		if width > 0 && lineWidth < width {
			result.WriteString(bgSeq)
			result.WriteString(strings.Repeat(" ", width-lineWidth))
			result.WriteString("\x1b[0m")
		} else {
			result.WriteString("\x1b[0m")
		}
	}

	if width > 0 && height > len(lines) {
		blankLine := bgSeq + strings.Repeat(" ", width) + "\x1b[0m"
		for i := len(lines); i < height; i++ {
			result.WriteByte('\n')
			result.WriteString(blankLine)
		}
	}

	return result.String()
}
