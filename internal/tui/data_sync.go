package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *AppModel) handleChatsLoaded(msg ChatsLoadedMsg) (tea.Model, tea.Cmd) {
	if msg.Err != nil {
		errStr := strings.ToLower(msg.Err.Error())
		if strings.Contains(errStr, "unauthorized") || strings.Contains(errStr, "invalid token") || strings.Contains(errStr, "401") || strings.Contains(errStr, "unauthenticated") {
			_ = m.client.Logout()
			m.state = StateAuth
			m.authView.SetErrorMessage("Session expired. Please log in again.")
			return *m, nil
		}
	}
	if msg.Err == nil && len(msg.Chats) > 0 {
		if m.chatView.ActiveChatID() == 0 {
			m.sidebarView.SelectIndex(0)
			if firstChat := m.sidebarView.SelectedChat(); firstChat != nil {
				m.chatView.SetActiveChat(firstChat.ID)
				_ = m.client.FocusConv(firstChat.ID)
				_ = m.client.MarkRead(firstChat.ID, 0)
				zero := 0
				m.client.UpdateChatSnippet(firstChat.ID, "", "", 0, &zero)
				return *m, fetchMessagesCmd(m.client, firstChat.ID)
			}
		} else {
			m.sidebarView.SelectChatByID(m.chatView.ActiveChatID())
		}
	}
	return *m, nil
}

func (m *AppModel) handleMessagesLoaded(msg MessagesLoadedMsg) (tea.Model, tea.Cmd) {
	if msg.Err == nil && msg.ChatID == m.chatView.ActiveChatID() {
		m.chatView.RefreshMessages()
		msgs := m.client.Messages(msg.ChatID)
		if len(msgs) > 0 {
			lastMsg := msgs[len(msgs)-1]
			_ = m.client.MarkRead(msg.ChatID, lastMsg.ID)
		}
	}
	return *m, nil
}
