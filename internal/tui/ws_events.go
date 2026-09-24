package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/DanielJohn17/tui-chat/internal/tui/client"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *AppModel) handleWSIncoming(msg WSIncomingMsg) (tea.Model, tea.Cmd) {
	var nextCmds []tea.Cmd
	nextCmds = append(nextCmds, m.listenWSCmd())

	switch p := msg.Event.(type) {
	case client.WSChatMessagePayload:
		if cmd := m.handleWSChatMessage(p); cmd != nil {
			nextCmds = append(nextCmds, cmd)
		}

	case client.WSChatNotificationPayload:
		if cmd := m.handleWSChatNotification(p); cmd != nil {
			nextCmds = append(nextCmds, cmd)
		}

	case client.WSConversationReadPayload:
		m.handleWSConversationRead(p)

	case client.WSUserPresencePayload:
		m.client.SetUserOnline(p.UserID, p.Online)

	case client.WSPresenceSnapshotPayload:
		m.client.SetOnlineUsers(p.OnlineUserIDs)

	case client.WSErrorPayload:
		model, cmd := m.handleWSError(p)
		if cmd != nil {
			nextCmds = append(nextCmds, cmd)
		}
		if model != nil {
			return model, tea.Batch(nextCmds...)
		}
	}

	return *m, tea.Batch(nextCmds...)
}

func (m *AppModel) triggerNotification(sender, content string) tea.Cmd {
	m.notificationID++
	curID := m.notificationID
	m.notificationText = fmt.Sprintf("🔔 @%s: %s", sender, content)

	clearCmd := tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return ClearNotificationMsg{ID: curID}
	})

	bellCmd := func() tea.Msg {
		fmt.Print("\a")
		return nil
	}

	return tea.Batch(clearCmd, bellCmd)
}

func (m *AppModel) handleWSChatMessage(p client.WSChatMessagePayload) tea.Cmd {
	senderName := "Recipient"
	chats := m.client.Chats()
	for _, ch := range chats {
		if ch.ID == p.ConvID {
			senderName = ch.Name
			if senderName == "" {
				senderName = ch.Username
			}
			break
		}
	}
	isSelf := p.SenderID == m.client.Profile().ID
	if isSelf {
		senderName = "You"
	}

	// Check if message is already stored in client memory (avoids duplication on sender window)
	existing := m.client.Messages(p.ConvID)
	alreadyExists := false
	for _, ex := range existing {
		if (p.ID > 0 && ex.ID == p.ID) || (isSelf && ex.ID == 0 && ex.Text == p.Content) {
			alreadyExists = true
			break
		}
	}

	if !alreadyExists {
		m.client.AppendMessage(client.Message{
			ID:        p.ID,
			SenderID:  p.SenderID,
			Sender:    senderName,
			ConvID:    p.ConvID,
			Text:      p.Content,
			Timestamp: "now",
			Self:      isSelf,
		})
	}

	var notifCmd tea.Cmd
	isActive := p.ConvID == m.chatView.ActiveChatID() && m.state == StateChat

	if isActive {
		m.chatView.RefreshMessages()
		if !isSelf {
			_ = m.client.MarkRead(p.ConvID, p.ID)
		}
		m.client.UpdateChatSnippet(p.ConvID, p.Content, "now", 0, nil)
	} else {
		unreadDelta := 1
		if isSelf {
			unreadDelta = 0
		}
		m.client.UpdateChatSnippet(p.ConvID, p.Content, "now", unreadDelta, nil)
		if !isSelf {
			notifCmd = m.triggerNotification(senderName, p.Content)
		}
	}

	// Retain sidebar selection focus on currently active chat
	if activeID := m.chatView.ActiveChatID(); activeID != 0 {
		m.sidebarView.SelectChatByID(activeID)
	}

	return notifCmd
}

func (m *AppModel) handleWSChatNotification(p client.WSChatNotificationPayload) tea.Cmd {
	senderName := p.SenderName
	if senderName == "" {
		senderName = "User"
	}

	// 1. Check if conversation exists in sidebar list; if not, create/prepend it!
	chats := m.client.Chats()
	found := false
	for _, ch := range chats {
		if ch.ID == p.ConvID {
			found = true
			break
		}
	}
	if !found {
		newChat := client.Chat{
			ID:          p.ConvID,
			RecipientID: p.SenderID,
			Name:        p.SenderName,
			Username:    p.SenderName,
			LastMessage: p.Content,
			Time:        "now",
			Unread:      p.UnreadCount,
			Online:      true,
		}
		m.client.SetChats(append([]client.Chat{newChat}, chats...))
	} else {
		m.client.UpdateChatSnippet(p.ConvID, p.Content, "now", 0, &p.UnreadCount)
	}

	// 2. Append message to local message buffer so opening this chat shows the message
	existing := m.client.Messages(p.ConvID)
	alreadyExists := false
	for _, ex := range existing {
		if ex.Text == p.Content && ex.Timestamp == "now" {
			alreadyExists = true
			break
		}
	}
	if !alreadyExists {
		m.client.AppendMessage(client.Message{
			ID:        0,
			SenderID:  p.SenderID,
			Sender:    senderName,
			ConvID:    p.ConvID,
			Text:      p.Content,
			Timestamp: "now",
			Self:      false,
		})
	}

	// 3. Retain sidebar selection focus on currently active chat
	if activeID := m.chatView.ActiveChatID(); activeID != 0 {
		m.sidebarView.SelectChatByID(activeID)
	}

	// 4. Trigger visual notification and terminal bell
	return m.triggerNotification(senderName, p.Content)
}

func (m *AppModel) handleWSConversationRead(p client.WSConversationReadPayload) {
	if p.UserID == m.client.Profile().ID {
		zero := 0
		m.client.UpdateChatSnippet(p.ConvID, "", "", 0, &zero)
		if activeID := m.chatView.ActiveChatID(); activeID != 0 {
			m.sidebarView.SelectChatByID(activeID)
		}
	}
}

func (m *AppModel) handleWSError(p client.WSErrorPayload) (tea.Model, tea.Cmd) {
	if strings.Contains(strings.ToLower(p.Error), "unauthorized") {
		_ = m.client.Logout()
		m.state = StateAuth
		m.authView.SetErrorMessage("Session expired. Please log in again.")
		return *m, nil
	}
	if p.Type == "disconnected" || strings.Contains(strings.ToLower(p.Error), "websocket connection lost") {
		return nil, m.scheduleDisconnectRetry()
	}
	return nil, nil
}
