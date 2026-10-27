package chat

import (
	"fmt"
	"strings"

	"github.com/DanielJohn17/tui-chat/internal/tui/client"
	"github.com/DanielJohn17/tui-chat/internal/tui/theme"
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) ScrollUp(lines int) {
	m.viewport.LineUp(lines)
}

func (m *Model) ScrollDown(lines int) {
	m.viewport.LineDown(lines)
}

func (m *Model) RefreshMessages() {
	if m.activeChatID == 0 {
		m.viewport.SetContent("No conversation selected.")
		return
	}

	status := StatusIdle
	if m.convStatus != nil {
		status = m.convStatus[m.activeChatID]
	}

	messages := m.client.Messages(m.activeChatID)

	// Loading state when messages are not yet in memory
	if status == StatusLoading && len(messages) == 0 {
		sp := m.spinnerFrame
		if sp == "" {
			sp = "⠋"
		}
		loadingText := lipgloss.JoinHorizontal(lipgloss.Center,
			lipgloss.NewStyle().Foreground(theme.ColorYellow).Bold(true).Render(sp),
			" ",
			theme.StyleDim.Render("Loading conversation messages..."),
		)
		m.viewport.SetContent("\n\n  " + loadingText)
		return
	}

	// Error / Failure state when fetching failed
	if status == StatusFailed && len(messages) == 0 {
		errDetail := m.convErrors[m.activeChatID]
		errMsg := "Failed to load messages"
		if errDetail != "" {
			errMsg = fmt.Sprintf("Failed to load messages: %s", errDetail)
		}
		errTitle := lipgloss.NewStyle().Foreground(theme.ColorRed).Bold(true).Render("✖ " + errMsg)
		retryHint := theme.StyleDim.Render("Press [ r ] to retry  •  Click to reload")
		errBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.ColorRed).
			Padding(1, 2).
			Render(lipgloss.JoinVertical(lipgloss.Center, errTitle, "", retryHint))
		m.viewport.SetContent("\n  " + errBox)
		return
	}

	if len(messages) == 0 {
		m.viewport.SetContent(theme.StyleDim.Render("\n  No messages yet. Send a message to break the ice!"))
		return
	}

	var sb strings.Builder
	contentWidth := m.viewport.Width - 2
	if contentWidth < 20 {
		contentWidth = 20
	}

	for _, msg := range messages {
		sb.WriteString(renderMessageRow(msg, contentWidth) + "\n")
	}

	m.viewport.SetContent(sb.String())
	m.viewport.GotoBottom()
}

func renderMessageRow(msg client.Message, contentWidth int) string {
	var senderTag string
	timeTag := theme.StyleDim.Render(msg.Timestamp)
	cleanText := theme.SanitizeText(msg.Text)

	if msg.Self {
		senderTag = lipgloss.NewStyle().Bold(true).Foreground(theme.ColorCyan).Render("You")
	} else {
		cleanSender := theme.SanitizeLine(msg.Sender)
		senderTag = lipgloss.NewStyle().Bold(true).Foreground(theme.ColorMagenta).Render(cleanSender)
	}

	header := fmt.Sprintf("%s  %s", senderTag, timeTag)
	body := lipgloss.NewStyle().
		Foreground(theme.ColorWhite).
		Width(contentWidth).
		Render(cleanText)

	return fmt.Sprintf("  %s\n  %s\n", header, body)
}
