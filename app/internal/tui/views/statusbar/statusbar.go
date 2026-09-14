package statusbar

import (
	"github.com/DanielJohn17/tui-chat/app/internal/tui/client"
	"github.com/DanielJohn17/tui-chat/app/internal/tui/theme"
	"github.com/charmbracelet/lipgloss"
)

func Render(activeChat *client.Chat, isInputFocused bool, width int) string {
	contentWidth := width - 4
	if contentWidth < 40 {
		contentWidth = 40
	}

	var leftText string
	if isInputFocused {
		leftText = lipgloss.JoinHorizontal(lipgloss.Center,
			theme.StyleBadge.Render("INPUT MODE"),
			" ",
			theme.StyleDim.Render("Esc to unfocus"),
		)
	} else if activeChat != nil {
		leftText = lipgloss.JoinHorizontal(lipgloss.Center,
			theme.StyleKeyBadge.Render("NAV MODE"),
			" ",
			theme.StyleSubtitle.Render(activeChat.Name),
		)
	} else {
		leftText = theme.StyleKeyBadge.Render("NAV MODE")
	}

	shortcuts := theme.StyleDim.Render("j/k: select • i: write • n: new DM • p: profile • h: help • q: quit")
	onlineStatus := theme.StyleSuccess.Render("● CONNECTED")

	spaces1 := (contentWidth - lipgloss.Width(leftText) - lipgloss.Width(shortcuts) - lipgloss.Width(onlineStatus)) / 2
	if spaces1 < 1 {
		spaces1 = 1
	}
	spaces2 := contentWidth - lipgloss.Width(leftText) - spaces1 - lipgloss.Width(shortcuts) - lipgloss.Width(onlineStatus)
	if spaces2 < 1 {
		spaces2 = 1
	}

	bar := lipgloss.JoinHorizontal(lipgloss.Center,
		leftText,
		lipgloss.NewStyle().Width(spaces1).Render(""),
		shortcuts,
		lipgloss.NewStyle().Width(spaces2).Render(""),
		onlineStatus,
	)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorBorderDim).
		Width(width).
		Padding(0, 1).
		Render(bar)
}
