package tui

import (
	"strings"

	"github.com/DanielJohn17/tui-chat/internal/tui/views/chat"
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
		// Set loading status for all conversations for background prefetch
		for _, ch := range msg.Chats {
			m.chatView.SetFetchStatus(ch.ID, chat.StatusLoading)
		}

		if m.chatView.ActiveChatID() == 0 {
			m.sidebarView.SelectIndex(0)
			if firstChat := m.sidebarView.SelectedChat(); firstChat != nil {
				m.chatView.SetActiveChat(firstChat.ID)
				_ = m.client.FocusConv(firstChat.ID)
				_ = m.client.MarkRead(firstChat.ID, 0)
				zero := 0
				m.client.UpdateChatSnippet(firstChat.ID, "", "", 0, &zero)
			}
		} else {
			m.sidebarView.SelectChatByID(m.chatView.ActiveChatID())
		}
		return *m, fetchBulkMessagesCmd(m.client)
	}
	return *m, nil
}

func (m *AppModel) handleBulkMessagesLoaded(msg BulkMessagesLoadedMsg) (tea.Model, tea.Cmd) {
	chats := m.client.Chats()
	if msg.Err == nil {
		for _, ch := range chats {
			m.chatView.SetFetchStatus(ch.ID, chat.StatusSuccess)
		}
		if activeID := m.chatView.ActiveChatID(); activeID != 0 {
			m.chatView.RefreshMessages()
			msgs := m.client.Messages(activeID)
			if len(msgs) > 0 {
				lastMsg := msgs[len(msgs)-1]
				_ = m.client.MarkRead(activeID, lastMsg.ID)
			}
		}
	} else {
		for _, ch := range chats {
			if len(m.client.Messages(ch.ID)) == 0 && m.chatView.FetchStatus(ch.ID) == chat.StatusLoading {
				m.chatView.SetFetchError(ch.ID, msg.Err.Error())
			}
		}
	}
	return *m, nil
}

func (m *AppModel) handleMessagesLoaded(msg MessagesLoadedMsg) (tea.Model, tea.Cmd) {
	if msg.Err == nil {
		m.chatView.SetFetchStatus(msg.ChatID, chat.StatusSuccess)
		if msg.ChatID == m.chatView.ActiveChatID() {
			m.chatView.RefreshMessages()
			msgs := m.client.Messages(msg.ChatID)
			if len(msgs) > 0 {
				lastMsg := msgs[len(msgs)-1]
				_ = m.client.MarkRead(msg.ChatID, lastMsg.ID)
			}
		}
	} else {
		if len(m.client.Messages(msg.ChatID)) == 0 {
			m.chatView.SetFetchError(msg.ChatID, msg.Err.Error())
		}
	}
	return *m, nil
}
