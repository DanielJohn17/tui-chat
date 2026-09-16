package sidebar

import (
	"fmt"

	"github.com/DanielJohn17/tui-chat/app/internal/tui/client"
	"github.com/DanielJohn17/tui-chat/app/internal/tui/theme"
	"github.com/charmbracelet/lipgloss"
)

func renderChatItem(ch client.Chat, isSelected bool, contentWidth int) string {
	innerWidth := contentWidth - 2
	if innerWidth < 12 {
		innerWidth = 12
	}

	var onlineDot string
	if ch.Online {
		onlineDot = theme.StyleSuccess.Render("●")
	} else {
		onlineDot = theme.StyleDim.Render("○")
	}

	// Line 1: Indicator + Name (left) and Time (right)
	var indicator string
	var nameStyle lipgloss.Style
	if isSelected {
		indicator = theme.StyleTitle.Render("▶ ")
		nameStyle = lipgloss.NewStyle().Bold(true).Foreground(theme.ColorCyan)
	} else {
		indicator = onlineDot + " "
		nameStyle = lipgloss.NewStyle().Bold(true).Foreground(theme.ColorWhite)
	}

	timeText := theme.StyleDim.Render(ch.Time)
	timeW := lipgloss.Width(timeText)

	availNameW := innerWidth - timeW - lipgloss.Width(indicator) - 1
	if availNameW < 4 {
		availNameW = 4
	}
	leftPart1 := indicator + nameStyle.Render(theme.Truncate(ch.Name, availNameW))
	spaces1 := innerWidth - lipgloss.Width(leftPart1) - timeW
	if spaces1 < 1 {
		spaces1 = 1
	}
	row1 := lipgloss.JoinHorizontal(lipgloss.Center, leftPart1, lipgloss.NewStyle().Width(spaces1).Render(""), timeText)

	// Line 2: Last message preview (left) and Unread badge pill (right)
	preview := ch.LastMessage
	if preview == "" {
		preview = "@" + ch.Username
	}

	var unreadBadge string
	if ch.Unread > 0 {
		badgeStr := fmt.Sprintf("%d", ch.Unread)
		if ch.Unread > 99 {
			badgeStr = "99+"
		}
		unreadBadge = lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.ColorBlack).
			Background(theme.ColorGreen).
			Padding(0, 1).
			Render(badgeStr)
	}

	unreadW := lipgloss.Width(unreadBadge)
	availMsgW := innerWidth - unreadW - 3
	if availMsgW < 4 {
		availMsgW = 4
	}

	var msgStyle lipgloss.Style
	if isSelected {
		msgStyle = lipgloss.NewStyle().Foreground(theme.ColorWhite)
	} else {
		msgStyle = theme.StyleDim
	}

	leftPart2 := "  " + msgStyle.Render(theme.Truncate(preview, availMsgW))
	spaces2 := innerWidth - lipgloss.Width(leftPart2) - unreadW
	if spaces2 < 1 {
		spaces2 = 1
	}
	row2 := lipgloss.JoinHorizontal(lipgloss.Center, leftPart2, lipgloss.NewStyle().Width(spaces2).Render(""), unreadBadge)

	rowBlank := lipgloss.NewStyle().Width(innerWidth).Render("")

	rows := lipgloss.JoinVertical(lipgloss.Left, row1, rowBlank, row2)

	if isSelected {
		return lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(theme.ColorCyan).
			Padding(0, 1, 0, 0).
			Render(rows)
	}

	return lipgloss.NewStyle().
		Padding(0, 1).
		Render(rows)
}
