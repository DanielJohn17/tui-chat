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

// Helper for text truncation
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-1] + "…"
}
