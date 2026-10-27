package chat

import (
	"strings"

	"github.com/DanielJohn17/tui-chat/internal/tui/client"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) FocusInput() {
	if m.activeTab == TabSecret && !m.peerHasKey {
		m.BlurInput()
		return
	}
	m.isFocused = true
	m.input.Focus()
}

func (m *Model) BlurInput() {
	m.isFocused = false
	m.input.Blur()
}

func (m Model) IsInputFocused() bool {
	return m.isFocused
}

func (m Model) InputValue() string {
	return m.input.Value()
}

func (m *Model) updateInputPlaceholder() {
	if m.activeTab == TabSecret {
		if !m.peerHasKey {
			m.input.Placeholder = "Message input disabled: peer has not published an encryption key"
		} else {
			m.input.Placeholder = "Type a secret message... (Press Enter to send, Esc to unfocus)"
		}
	} else {
		m.input.Placeholder = "Type a message... (Press Enter to send, Esc to unfocus)"
	}
}

func (m *Model) updateViewportHeight() {
	extraLines := 9
	if m.activeTab == TabSecret {
		extraLines = 10
	}
	vpHeight := m.height - extraLines
	if vpHeight < 1 {
		vpHeight = 1
	}
	m.viewport.Height = vpHeight
}

func (m *Model) handleInputKey(msg tea.KeyMsg) (Model, tea.Cmd, bool) {
	switch msg.String() {
	case "ctrl+s":
		return *m, func() tea.Msg { return ToggleTabMsg{} }, true
	case "esc":
		m.BlurInput()
		return *m, nil, true
	case "enter":
		if m.activeTab == TabSecret && !m.peerHasKey {
			return *m, nil, true
		}
		text := strings.TrimSpace(m.input.Value())
		if text != "" && m.activeChatID != 0 {
			if err := m.client.SendWS(m.activeChatID, text); err != nil {
				// Do not append message or clear input on SendWS failure
				return *m, nil, true
			}
			nowStr := "now"
			m.client.AppendMessage(client.Message{
				ConvID:    m.activeChatID,
				SenderID:  m.client.Profile().ID,
				Sender:    "You",
				Text:      text,
				Timestamp: nowStr,
				Self:      true,
				IsSecret:  m.activeTab == TabSecret,
			})
			m.client.UpdateChatSnippet(m.activeChatID, text, nowStr, 0, nil)
			m.input.Reset()
			m.RefreshMessages()
			m.viewport.GotoBottom()
		}
		return *m, nil, true
	}
	return *m, nil, false
}
