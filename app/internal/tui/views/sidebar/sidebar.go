package sidebar

import (
	"fmt"

	"github.com/DanielJohn17/tui-chat/app/internal/tui/client"
	"github.com/DanielJohn17/tui-chat/app/internal/tui/theme"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	client        client.Client
	selectedIndex int
	width         int
	height        int
}

func New(c client.Client) Model {
	return Model{
		client:        c,
		selectedIndex: 0,
		width:         30,
		height:        24,
	}
}

func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m *Model) MoveUp() {
	chats := m.client.Chats()
	if len(chats) == 0 {
		return
	}
	if m.selectedIndex > 0 {
		m.selectedIndex--
	}
}

func (m *Model) MoveDown() {
	chats := m.client.Chats()
	if len(chats) == 0 {
		return
	}
	if m.selectedIndex < len(chats)-1 {
		m.selectedIndex++
	}
}

func (m Model) SelectedIndex() int {
	return m.selectedIndex
}

func (m Model) SelectedChat() *client.Chat {
	chats := m.client.Chats()
	if len(chats) == 0 || m.selectedIndex < 0 || m.selectedIndex >= len(chats) {
		return nil
	}
	return &chats[m.selectedIndex]
}

func (m Model) View() string {
	chats := m.client.Chats()
	contentWidth := m.width - 4
	if contentWidth < 20 {
		contentWidth = 20
	}

	// 1. Header
	headerLeft := theme.StyleTitle.Render("◈ CONVERSATIONS")
	headerRight := theme.StyleSuccess.Render("[+n]")
	headerSpaces := contentWidth - lipgloss.Width(headerLeft) - lipgloss.Width(headerRight)
	if headerSpaces < 1 {
		headerSpaces = 1
	}
	header := lipgloss.JoinHorizontal(lipgloss.Center,
		headerLeft,
		lipgloss.NewStyle().Width(headerSpaces).Render(""),
		headerRight,
	)

	// 2. Chat list items
	var chatItems []string
	availableListHeight := m.height - 6
	if availableListHeight < 4 {
		availableListHeight = 4
	}

	for i, ch := range chats {
		isSelected := i == m.selectedIndex

		var onlineDot string
		if ch.Online {
			onlineDot = theme.StyleSuccess.Render("●")
		} else {
			onlineDot = theme.StyleDim.Render("○")
		}

		var row1, row2, row3 string
		if isSelected {
			marker := theme.StyleTitle.Render("▶ ")
			nameText := lipgloss.NewStyle().Bold(true).Foreground(theme.ColorCyan).Render(theme.Truncate(ch.Name, contentWidth-8))
			row1 = lipgloss.JoinHorizontal(lipgloss.Center, marker, nameText)
			rSpaces := contentWidth - lipgloss.Width(row1) - lipgloss.Width(onlineDot)
			if rSpaces < 1 {
				rSpaces = 1
			}
			row1 = lipgloss.JoinHorizontal(lipgloss.Center, row1, lipgloss.NewStyle().Width(rSpaces).Render(""), onlineDot)

			userHandle := lipgloss.NewStyle().Foreground(theme.ColorMagenta).Render("@" + theme.Truncate(ch.Username, contentWidth/2))
			timeText := theme.StyleDim.Render(ch.Time)
			uSpaces := contentWidth - lipgloss.Width(userHandle) - lipgloss.Width(timeText)
			if uSpaces < 1 {
				uSpaces = 1
			}
			row2 = lipgloss.JoinHorizontal(lipgloss.Center, userHandle, lipgloss.NewStyle().Width(uSpaces).Render(""), timeText)

			if ch.LastMessage != "" {
				row3 = theme.StyleDim.Render(theme.Truncate(ch.LastMessage, contentWidth-2))
			}

			itemBox := lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(theme.ColorCyan).
				Padding(0, 1).
				Render(lipgloss.JoinVertical(lipgloss.Left, row1, row2, row3))
			chatItems = append(chatItems, itemBox)
		} else {
			marker := "  "
			nameText := lipgloss.NewStyle().Foreground(theme.ColorWhite).Render(theme.Truncate(ch.Name, contentWidth-8))
			row1 = lipgloss.JoinHorizontal(lipgloss.Center, marker, nameText)
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
			row2 = lipgloss.JoinHorizontal(lipgloss.Center, userHandle, lipgloss.NewStyle().Width(uSpaces).Render(""), trailing)

			itemBox := lipgloss.NewStyle().
				Padding(0, 1).
				Render(lipgloss.JoinVertical(lipgloss.Left, row1, row2))
			chatItems = append(chatItems, itemBox)
		}
	}

	chatsList := lipgloss.JoinVertical(lipgloss.Left, chatItems...)

	// 3. User Profile Footer
	prof := m.client.Profile()
	profName := theme.Truncate(prof.Name, contentWidth/2)
	profHandle := "@" + theme.Truncate(prof.Username, contentWidth/2-2)
	pDot := theme.StyleSuccess.Render("●")
	pLeft := lipgloss.JoinHorizontal(lipgloss.Center, pDot, " ", lipgloss.NewStyle().Bold(true).Foreground(theme.ColorWhite).Render(profName))
	pRight := lipgloss.NewStyle().Foreground(theme.ColorMagenta).Render(profHandle)
	pSpaces := contentWidth - lipgloss.Width(pLeft) - lipgloss.Width(pRight)
	if pSpaces < 1 {
		pSpaces = 1
	}
	footer := lipgloss.JoinHorizontal(lipgloss.Center, pLeft, lipgloss.NewStyle().Width(pSpaces).Render(""), pRight)

	// Combine into container
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorBorderDim).
		Width(m.width).
		Height(m.height).
		Padding(0, 1).
		Render(lipgloss.JoinVertical(lipgloss.Left,
			header,
			"",
			chatsList,
			"",
			footer,
		))

	return box
}
