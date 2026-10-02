package tui

import (
	"context"
	"time"
	"unicode/utf8"

	"github.com/DanielJohn17/tui-chat/internal/tui/client"
	"github.com/DanielJohn17/tui-chat/internal/tui/views/chat"
	tea "github.com/charmbracelet/bubbletea"
)

type DiscoveryDebounceMsg struct {
	Generation uint64
	Query      string
	Context    context.Context
}

type DiscoverySearchMsg struct {
	Generation uint64
	Users      []client.UserSummary
	Err        error
}

type DiscoveryOpenMsg struct {
	Generation uint64
	Chat       client.Chat
	Err        error
}

func (m *AppModel) cancelDiscovery() {
	m.discoveryGeneration++
	if m.discoveryCancel != nil {
		m.discoveryCancel()
		m.discoveryCancel = nil
	}
}

func (m *AppModel) openDiscovery() {
	m.cancelDiscovery()
	m.modal = ModalNewDM
	m.newDMView.Reset()
}

func (m *AppModel) scheduleDiscovery() tea.Cmd {
	m.cancelDiscovery()
	query := m.newDMView.Value()
	eligible := utf8.RuneCountInString(query) >= 2
	m.newDMView.SetLoading(eligible)
	if !eligible {
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	m.discoveryCancel = cancel
	generation := m.discoveryGeneration
	return func() tea.Msg {
		timer := time.NewTimer(250 * time.Millisecond)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return nil
		case <-timer.C:
			return DiscoveryDebounceMsg{Generation: generation, Query: query, Context: ctx}
		}
	}
}

func (m *AppModel) discoveryCurrent(generation uint64) bool {
	return m.state == StateChat && m.modal == ModalNewDM && generation == m.discoveryGeneration
}

func (m *AppModel) handleDiscoveryDebounce(msg DiscoveryDebounceMsg) (tea.Model, tea.Cmd) {
	if !m.discoveryCurrent(msg.Generation) || msg.Context == nil || msg.Context.Err() != nil || m.newDMView.Opening() {
		return *m, nil
	}
	c := m.client
	return *m, func() tea.Msg {
		users, err := c.SearchUsers(msg.Context, msg.Query)
		return DiscoverySearchMsg{Generation: msg.Generation, Users: users, Err: err}
	}
}

func (m *AppModel) handleDiscoverySearch(msg DiscoverySearchMsg) (tea.Model, tea.Cmd) {
	if m.discoveryCurrent(msg.Generation) && !m.newDMView.Opening() {
		m.newDMView.SetResults(msg.Users, msg.Err)
	}
	return *m, nil
}

func (m *AppModel) submitNewDM() (tea.Model, tea.Cmd) {
	if m.newDMView.Opening() {
		return *m, nil
	}
	user, ok := m.newDMView.Selected()
	if !ok {
		return *m, nil
	}
	m.cancelDiscovery()
	ctx, cancel := context.WithCancel(context.Background())
	m.discoveryCancel = cancel
	generation := m.discoveryGeneration
	m.newDMView.StartOpening()
	c := m.client
	return *m, func() tea.Msg {
		opened, err := c.OpenDirectConversation(ctx, user.ID)
		return DiscoveryOpenMsg{Generation: generation, Chat: opened, Err: err}
	}
}

func (m *AppModel) handleDiscoveryOpen(msg DiscoveryOpenMsg) (tea.Model, tea.Cmd) {
	if !m.discoveryCurrent(msg.Generation) || !m.newDMView.Opening() {
		return *m, nil
	}
	if msg.Err != nil {
		m.newDMView.SetError(msg.Err.Error())
		return *m, nil
	}
	if msg.Chat.ID <= 0 {
		m.newDMView.SetError("Server returned an invalid conversation")
		return *m, nil
	}
	chats := m.client.Chats()
	found := false
	for i := range chats {
		if chats[i].ID == msg.Chat.ID {
			// Discovery supplies identity, not a replacement for live chat metadata.
			chats[i].RecipientID = msg.Chat.RecipientID
			chats[i].Name = msg.Chat.Name
			chats[i].Username = msg.Chat.Username
			found = true
			break
		}
	}
	if !found {
		chats = append([]client.Chat{msg.Chat}, chats...)
	}
	m.client.SetChats(chats)
	m.sidebarView.SelectChatByID(msg.Chat.ID)
	m.chatView.SetActiveChat(msg.Chat.ID)
	m.chatView.BlurInput()
	m.cancelDiscovery()
	m.modal = ModalNone
	c, id := m.client, msg.Chat.ID
	focus := func() tea.Msg {
		_ = c.FocusConv(id)
		_ = c.MarkRead(id, 0)
		return nil
	}
	if len(c.Messages(id)) == 0 && m.chatView.FetchStatus(id) != chat.StatusLoading {
		m.chatView.SetFetchStatus(id, chat.StatusLoading)
		return *m, tea.Batch(focus, fetchMessagesCmd(c, id))
	}
	return *m, focus
}

func markReadCmd(c client.Client, chatID, messageID int64) tea.Cmd {
	return func() tea.Msg {
		_ = c.MarkRead(chatID, messageID)
		return nil
	}
}
