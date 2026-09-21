package sidebar

import (
	"strings"

	"github.com/DanielJohn17/tui-chat/app/internal/tui/client"
	"github.com/DanielJohn17/tui-chat/app/internal/tui/theme"
	"github.com/charmbracelet/lipgloss"
)

func renderFooter(prof client.Profile, contentWidth int) string {
	profName := theme.Truncate(prof.Name, contentWidth/2)
	profHandle := "@" + theme.Truncate(prof.Username, contentWidth/2-2)
	pDot := theme.StyleSuccess.Render("●")
	pLeft := lipgloss.JoinHorizontal(lipgloss.Center, pDot, " ", lipgloss.NewStyle().Bold(true).Foreground(theme.ColorWhite).Render(profName))
	pRight := lipgloss.NewStyle().Foreground(theme.ColorMagenta).Render(profHandle)
	pSpaces := contentWidth - lipgloss.Width(pLeft) - lipgloss.Width(pRight)
	if pSpaces < 1 {
		pSpaces = 1
	}
	profileRow := lipgloss.JoinHorizontal(lipgloss.Center, pLeft, lipgloss.NewStyle().Width(pSpaces).Render(""), pRight)
	profileHint := theme.StyleDim.Render("p / click: edit profile")
	pHintSpaces := contentWidth - lipgloss.Width(profileHint)
	if pHintSpaces < 1 {
		pHintSpaces = 1
	}
	profileSubRow := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Right).Render(profileHint)

	divider := lipgloss.NewStyle().Foreground(theme.ColorBorderDim).Render(strings.Repeat("─", contentWidth))

	return lipgloss.JoinVertical(lipgloss.Left,
		divider,
		profileRow,
		profileSubRow,
	)
}
