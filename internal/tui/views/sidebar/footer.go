package sidebar

import (
	"strings"

	"github.com/DanielJohn17/tui-chat/internal/tui/client"
	"github.com/DanielJohn17/tui-chat/internal/tui/theme"
	"github.com/charmbracelet/lipgloss"
)

func renderFooter(prof client.Profile, contentWidth int) string {
	if contentWidth < 1 {
		return ""
	}
	nameLimit := contentWidth / 2
	if nameLimit < 2 {
		nameLimit = 2
	}
	profName := theme.Truncate(prof.Name, nameLimit)
	profHandle := "@" + theme.Truncate(prof.Username, max(2, contentWidth/2-2))
	pDot := theme.StyleSuccess.Render("●")
	pLeft := lipgloss.JoinHorizontal(lipgloss.Center, pDot, " ", lipgloss.NewStyle().Bold(true).Foreground(theme.ColorWhite).Render(profName))
	pRight := lipgloss.NewStyle().Foreground(theme.ColorMagenta).Render(profHandle)
	pSpaces := contentWidth - lipgloss.Width(pLeft) - lipgloss.Width(pRight)
	var profileRow string
	if pSpaces >= 1 {
		profileRow = lipgloss.JoinHorizontal(lipgloss.Center, pLeft, lipgloss.NewStyle().Width(pSpaces).Render(""), pRight)
	} else if contentWidth >= lipgloss.Width(pLeft) {
		profileRow = pLeft
	} else {
		profileRow = theme.Truncate(pLeft, contentWidth)
	}

	var profileSubRow string
	if contentWidth >= 24 {
		profileHint := theme.StyleDim.Render("p / click: edit profile")
		profileSubRow = lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Right).Render(profileHint)
	} else if contentWidth >= 12 {
		profileHint := theme.StyleDim.Render(theme.Truncate("p: profile", contentWidth))
		profileSubRow = lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Right).Render(profileHint)
	}

	divider := lipgloss.NewStyle().Foreground(theme.ColorBorderDim).Render(strings.Repeat("─", contentWidth))

	parts := []string{divider, profileRow}
	if profileSubRow != "" {
		parts = append(parts, profileSubRow)
	}
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}
