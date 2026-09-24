package statusbar

import (
	"fmt"

	"github.com/DanielJohn17/tui-chat/app/internal/tui/client"
	"github.com/DanielJohn17/tui-chat/app/internal/tui/theme"
	"github.com/charmbracelet/lipgloss"
)

type ConnectionStatus int

const (
	StatusConnected ConnectionStatus = iota
	StatusConnecting
	StatusReconnecting
	StatusOffline
)

func Render(activeChat *client.Chat, isInputFocused bool, status ConnectionStatus, retrySeconds int, spinnerFrame string, notificationText string, width int) string {
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

	if spinnerFrame == "" {
		spinnerFrame = "⟳"
	}

	var onlineStatus string
	switch status {
	case StatusConnected:
		if width < 85 {
			onlineStatus = theme.StyleSuccess.Render("●")
		} else {
			onlineStatus = theme.StyleSuccess.Render("● LIVE")
		}
	case StatusConnecting:
		if width < 85 {
			onlineStatus = lipgloss.NewStyle().Foreground(theme.ColorYellow).Render(spinnerFrame)
		} else {
			onlineStatus = lipgloss.NewStyle().Foreground(theme.ColorYellow).Render(spinnerFrame + " CONNECTING...")
		}
	case StatusReconnecting:
		if width < 85 {
			if retrySeconds > 0 {
				onlineStatus = lipgloss.NewStyle().Foreground(theme.ColorYellow).Render(fmt.Sprintf("%s %ds", spinnerFrame, retrySeconds))
			} else {
				onlineStatus = lipgloss.NewStyle().Foreground(theme.ColorYellow).Render(spinnerFrame)
			}
		} else {
			if retrySeconds > 0 {
				onlineStatus = lipgloss.NewStyle().Foreground(theme.ColorYellow).Render(fmt.Sprintf("%s RETRYING IN %ds...", spinnerFrame, retrySeconds))
			} else {
				onlineStatus = lipgloss.NewStyle().Foreground(theme.ColorYellow).Render(spinnerFrame + " RETRYING...")
			}
		}
	case StatusOffline:
		fallthrough
	default:
		if width < 85 {
			onlineStatus = lipgloss.NewStyle().Foreground(theme.ColorRed).Render("○")
		} else {
			onlineStatus = lipgloss.NewStyle().Foreground(theme.ColorRed).Render("○ OFFLINE")
		}
	}

	var middleText string
	if notificationText != "" {
		availNotifW := contentWidth - lipgloss.Width(leftText) - lipgloss.Width(onlineStatus) - 6
		if availNotifW < 12 {
			availNotifW = 12
		}
		middleText = lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.ColorBlack).
			Background(theme.ColorYellow).
			Padding(0, 1).
			Render(theme.Truncate(notificationText, availNotifW))
	} else {
		var shortcutsText string
		if width < 85 {
			if status == StatusOffline || status == StatusReconnecting {
				shortcutsText = "j/k • i: write • n: DM • r: retry • ?: help • q: quit"
			} else {
				shortcutsText = "j/k • i: write • n: DM • ?: help • q: quit"
			}
		} else if width < 105 {
			if status == StatusOffline || status == StatusReconnecting {
				shortcutsText = "j/k: nav • i: write • n: DM • r: retry • p: profile • ?: help • q: quit"
			} else {
				shortcutsText = "j/k: nav • i: write • n: DM • p: profile • ?: help • q: quit"
			}
		} else {
			if status == StatusOffline || status == StatusReconnecting {
				shortcutsText = "j/k: select • i: write • n: new DM • r: retry • p: profile • ?: help • q: quit"
			} else {
				shortcutsText = "j/k: select • i: write • n: new DM • p: profile • ?: help • q: quit"
			}
		}
		middleText = theme.StyleDim.Render(shortcutsText)
	}

	spaces1 := (contentWidth - lipgloss.Width(leftText) - lipgloss.Width(middleText) - lipgloss.Width(onlineStatus)) / 2
	if spaces1 < 1 {
		spaces1 = 1
	}
	spaces2 := contentWidth - lipgloss.Width(leftText) - spaces1 - lipgloss.Width(middleText) - lipgloss.Width(onlineStatus)
	if spaces2 < 1 {
		spaces2 = 1
	}

	bar := lipgloss.JoinHorizontal(lipgloss.Center,
		leftText,
		lipgloss.NewStyle().Width(spaces1).Render(""),
		middleText,
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
