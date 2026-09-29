package sidebar

import (
	"github.com/DanielJohn17/tui-chat/internal/tui/theme"
	"github.com/charmbracelet/lipgloss"
)

func renderHeader(contentWidth int) string {
	headerLeft := theme.StyleTitle.Render("◈ ") + theme.RenderLineTalkWordmark()
	headerRight := theme.StyleSuccess.Render("[+n]")
	headerSpaces := contentWidth - lipgloss.Width(headerLeft) - lipgloss.Width(headerRight)
	if headerSpaces >= 1 {
		return lipgloss.JoinHorizontal(lipgloss.Center,
			headerLeft,
			lipgloss.NewStyle().Width(headerSpaces).Render(""),
			headerRight,
		)
	}
	if contentWidth >= lipgloss.Width(headerLeft) {
		return headerLeft
	}
	return theme.Truncate(headerLeft, max(1, contentWidth))
}
