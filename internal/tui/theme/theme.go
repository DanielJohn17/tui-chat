package theme

import (
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
