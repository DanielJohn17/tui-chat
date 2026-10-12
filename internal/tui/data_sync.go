package tui

import (
	"context"
	"strings"

	"github.com/DanielJohn17/tui-chat/internal/tui/client"
	"github.com/DanielJohn17/tui-chat/internal/tui/views/chat"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *AppModel) handleChatsLoaded(msg ChatsLoadedMsg) (tea.Model, tea.Cmd) {
	if msg.Err != nil {
		errStr := strings.ToLower(msg.Err.Error())
		if strings.Contains(errStr, "unauthorized") || strings.Contains(errStr, "invalid token") || strings.Contains(errStr, "401") || strings.Contains(errStr, "unauthenticated") {
			m.cancelDiscovery()
			m.modal = ModalNone
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
			if ch.CloudID != 0 {
				m.chatView.SetFetchStatus(ch.CloudID, chat.StatusLoading)
			}
			if ch.SecretID != 0 {
				m.chatView.SetFetchStatus(ch.SecretID, chat.StatusLoading)
			}
		}

		if m.chatView.ActiveChatID() == 0 {
			m.sidebarView.SelectIndex(0)
			if firstChat := m.sidebarView.SelectedChat(); firstChat != nil {
				_ = m.selectChat(firstChat)
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
			if ch.CloudID != 0 {
				m.chatView.SetFetchStatus(ch.CloudID, chat.StatusSuccess)
			}
			if ch.SecretID != 0 {
				m.chatView.SetFetchStatus(ch.SecretID, chat.StatusSuccess)
			}
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
				return *m, markReadCmd(m.client, msg.ChatID, lastMsg.ID)
			}
		}
	} else {
		if len(m.client.Messages(msg.ChatID)) == 0 {
			m.chatView.SetFetchError(msg.ChatID, msg.Err.Error())
		}
	}
	return *m, nil
}

type EnsureConversationMsg struct {
	RecipientID int64
	IsSecret    bool
	Chat        client.Chat
	Err         error
}

func ensureConversationCmd(c client.Client, recipientID int64, isSecret bool) tea.Cmd {
	return func() tea.Msg {
		ch, err := c.OpenDirectConversation(context.Background(), recipientID, isSecret)
		return EnsureConversationMsg{
			RecipientID: recipientID,
			IsSecret:    isSecret,
			Chat:        ch,
			Err:         err,
		}
	}
}

func (m *AppModel) selectChat(ch *client.Chat) tea.Cmd {
	if ch == nil {
		return nil
	}
	targetTab := m.chatView.ActiveTab()
	var convID int64
	if targetTab == chat.TabSecret {
		convID = ch.SecretID
		if convID == 0 && ch.CloudID != 0 {
			targetTab = chat.TabCloud
			m.chatView.SetActiveTab(chat.TabCloud)
			convID = ch.CloudID
		}
	} else {
		convID = ch.CloudID
		if convID == 0 && ch.SecretID != 0 {
			targetTab = chat.TabSecret
			m.chatView.SetActiveTab(chat.TabSecret)
			convID = ch.SecretID
		}
	}
	if convID == 0 {
		convID = ch.ID
	}

	if targetTab == chat.TabSecret {
		peerPub, err := m.client.GetPeerPublicKey(ch.RecipientID)
		hasKey := (err == nil && len(peerPub) == 32)
		m.chatView.SetPeerKeyStatus(hasKey)
	}

	m.chatView.SetActiveChat(convID)
	_ = m.client.FocusConv(convID)
	_ = m.client.MarkRead(convID, 0)
	zero := 0
	m.client.UpdateChatSnippet(convID, "", "", 0, &zero)
	if len(m.client.Messages(convID)) == 0 && m.chatView.FetchStatus(convID) != chat.StatusLoading {
		m.chatView.SetFetchStatus(convID, chat.StatusLoading)
		return fetchMessagesCmd(m.client, convID)
	}
	return nil
}

func (m *AppModel) toggleChatTab() (tea.Model, tea.Cmd) {
	sel := m.sidebarView.SelectedChat()
	if sel == nil {
		return *m, nil
	}
	targetTab := chat.TabSecret
	if m.chatView.ActiveTab() == chat.TabSecret {
		targetTab = chat.TabCloud
	}
	return m.switchChatTab(sel, targetTab)
}

func (m *AppModel) switchChatTab(sel *client.Chat, targetTab chat.ChatTab) (tea.Model, tea.Cmd) {
	m.chatView.SetActiveTab(targetTab)

	var targetConvID int64
	if targetTab == chat.TabSecret {
		targetConvID = sel.SecretID
	} else {
		targetConvID = sel.CloudID
	}

	// If secret mode, check if peer has public key
	if targetTab == chat.TabSecret {
		peerPub, err := m.client.GetPeerPublicKey(sel.RecipientID)
		hasKey := (err == nil && len(peerPub) == 32)
		m.chatView.SetPeerKeyStatus(hasKey)
	}

	if targetConvID != 0 {
		m.chatView.SetActiveChat(targetConvID)
		_ = m.client.FocusConv(targetConvID)
		_ = m.client.MarkRead(targetConvID, 0)
		zero := 0
		m.client.UpdateChatSnippet(targetConvID, "", "", 0, &zero)
		if len(m.client.Messages(targetConvID)) == 0 && m.chatView.FetchStatus(targetConvID) != chat.StatusLoading {
			m.chatView.SetFetchStatus(targetConvID, chat.StatusLoading)
			return *m, fetchMessagesCmd(m.client, targetConvID)
		}
		return *m, nil
	}

	// Target conversation does not exist yet; create it on the fly
	isSecret := (targetTab == chat.TabSecret)
	return *m, ensureConversationCmd(m.client, sel.RecipientID, isSecret)
}
