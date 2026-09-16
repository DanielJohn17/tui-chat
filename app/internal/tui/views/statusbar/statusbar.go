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
		badgeText := "NAV MODE"
		nameLimit := 16
		if width < 85 {
			badgeText = "NAV"
			nameLimit = 10
		}
		leftText = lipgloss.JoinHorizontal(lipgloss.Center,
			theme.StyleKeyBadge.Render(badgeText),
			" ",
			theme.StyleSubtitle.Render(theme.Truncate(activeChat.Name, nameLimit)),
		)
	} else {
		leftText = theme.StyleKeyBadge.Render("NAV MODE")
	}

	var shortcutsText string
	if width < 85 {
		shortcutsText = "j/k • i: write • n: DM • ?: help • q: quit"
	} else if width < 105 {
		shortcutsText = "j/k: nav • i: write • n: DM • p: profile • ?: help • q: quit"
	} else {
		shortcutsText = "j/k: select • i: write • n: new DM • p: profile • ?: help • q: quit"
	}
	shortcuts := theme.StyleDim.Render(shortcutsText)

	var onlineStatus string
	if width < 85 {
		onlineStatus = theme.StyleSuccess.Render("●")
	} else {
		onlineStatus = theme.StyleSuccess.Render("● CONNECTED")
	}

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

	boxWidth := width - 2
	if boxWidth < 20 {
		boxWidth = 20
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorBorderDim).
		Width(boxWidth).
		Padding(0, 1).
		Render(bar)
}
