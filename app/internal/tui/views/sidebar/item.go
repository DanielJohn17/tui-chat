package sidebar

import (
	"fmt"

	"github.com/DanielJohn17/tui-chat/app/internal/tui/client"
	"github.com/DanielJohn17/tui-chat/app/internal/tui/theme"
	"github.com/charmbracelet/lipgloss"
)

func renderChatItem(ch client.Chat, isSelected bool, contentWidth int) string {
	var onlineDot string
	if ch.Online {
		onlineDot = theme.StyleSuccess.Render("●")
	} else {
		onlineDot = theme.StyleDim.Render("○")
	}

	if isSelected {
		itemWidth := contentWidth - 1
		if itemWidth < 10 {
			itemWidth = 10
		}
		marker := theme.StyleTitle.Render("▶ ")
		nameText := lipgloss.NewStyle().Bold(true).Foreground(theme.ColorCyan).Render(theme.Truncate(ch.Name, itemWidth-8))
		row1 := lipgloss.JoinHorizontal(lipgloss.Center, marker, nameText)
		rSpaces := itemWidth - lipgloss.Width(row1) - lipgloss.Width(onlineDot)
		if rSpaces < 1 {
			rSpaces = 1
		}
		row1 = lipgloss.JoinHorizontal(lipgloss.Center, row1, lipgloss.NewStyle().Width(rSpaces).Render(""), onlineDot)

		userHandle := lipgloss.NewStyle().Foreground(theme.ColorMagenta).Render("@" + theme.Truncate(ch.Username, itemWidth/2))
		timeText := theme.StyleDim.Render(ch.Time)
		uSpaces := itemWidth - lipgloss.Width(userHandle) - lipgloss.Width(timeText)
		if uSpaces < 1 {
			uSpaces = 1
		}
		row2 := lipgloss.JoinHorizontal(lipgloss.Center, userHandle, lipgloss.NewStyle().Width(uSpaces).Render(""), timeText)

		var rows []string
		rows = append(rows, row1, row2)
		if ch.LastMessage != "" {
			row3 := theme.StyleDim.Render(theme.Truncate(ch.LastMessage, itemWidth-2))
			rows = append(rows, row3)
		}

		return lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(theme.ColorCyan).
			Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
	}

	// Unselected chat item
	marker := "  "
	nameText := lipgloss.NewStyle().Foreground(theme.ColorWhite).Render(theme.Truncate(ch.Name, contentWidth-8))
	row1 := lipgloss.JoinHorizontal(lipgloss.Center, marker, nameText)
	rSpaces := contentWidth - lipgloss.Width(row1) - lipgloss.Width(onlineDot)
	if rSpaces < 1 {
		rSpaces = 1
	}
	row1 = lipgloss.JoinHorizontal(lipgloss.Center, row1, lipgloss.NewStyle().Width(rSpaces).Render(""), onlineDot)

	var unreadPill string
	if ch.Unread > 0 {
		unreadPill = lipgloss.NewStyle().Bold(true).Foreground(theme.ColorYellow).Render(fmt.Sprintf("(%d)", ch.Unread))
	}
	timeText := theme.StyleDim.Render(ch.Time)
	trailing := lipgloss.JoinHorizontal(lipgloss.Center, unreadPill, " ", timeText)
	userHandle := theme.StyleDim.Render("@" + theme.Truncate(ch.Username, contentWidth/2))
	uSpaces := contentWidth - lipgloss.Width(userHandle) - lipgloss.Width(trailing)
	if uSpaces < 1 {
		uSpaces = 1
	}
	row2 := lipgloss.JoinHorizontal(lipgloss.Center, userHandle, lipgloss.NewStyle().Width(uSpaces).Render(""), trailing)

	return lipgloss.JoinVertical(lipgloss.Left, row1, row2)
}
