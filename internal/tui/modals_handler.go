package tui

import (
	"time"

	"github.com/DanielJohn17/tui-chat/app/internal/tui/client"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *AppModel) handleModalUpdate(msg tea.Msg) (tea.Model, tea.Cmd, bool) {
	if m.modal == ModalNone {
		return nil, nil, false
	}

	switch m.modal {
	case ModalHelp:
		model, cmd := m.handleHelpModalUpdate(msg)
		return model, cmd, true
	case ModalNewDM:
		model, cmd := m.handleNewDMModalUpdate(msg)
		return model, cmd, true
	}

	return nil, nil, false
}

func (m *AppModel) handleHelpModalUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	if mouseMsg, ok := msg.(tea.MouseMsg); ok {
		if mouseMsg.Action == tea.MouseActionPress && mouseMsg.Button == tea.MouseButtonLeft {
			m.modal = ModalNone
			return *m, nil
		}
	}
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "esc", "enter", "?", "q", "f1", "ctrl+h":
			m.modal = ModalNone
			return *m, nil
		}
	}
	return *m, nil
}

func (m *AppModel) handleNewDMModalUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	if mouseMsg, ok := msg.(tea.MouseMsg); ok {
		if mouseMsg.Action == tea.MouseActionPress && mouseMsg.Button == tea.MouseButtonLeft {
			boxW := 48
			boxH := 10
			startX := (m.width - boxW) / 2
			startY := (m.height - boxH) / 2
			relY := mouseMsg.Y - startY
			relX := mouseMsg.X - startX

			if relY >= 5 && relY <= 8 && relX >= 0 && relX < boxW {
				if relX < boxW/2 {
					return m.submitNewDM()
				}
				m.modal = ModalNone
				return *m, nil
			}
		}
	}
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "esc":
			m.modal = ModalNone
			return *m, nil
		case "enter":
			return m.submitNewDM()
		}
	}

	var cmd tea.Cmd
	m.newDMView, cmd = m.newDMView.Update(msg)
	return *m, cmd
}

func (m *AppModel) submitNewDM() (tea.Model, tea.Cmd) {
	username := m.newDMView.Value()
	if len(username) < 3 {
		m.newDMView.SetError("Username must be at least 3 characters")
		return *m, nil
	}
	chats := m.client.Chats()
	newChat := client.Chat{
		ID:          time.Now().UnixNano(),
		RecipientID: 0,
		Name:        username,
		Username:    username,
		Time:        "now",
		Unread:      0,
		Online:      false,
	}
	m.client.SetChats(append([]client.Chat{newChat}, chats...))
	m.sidebarView.SelectIndex(0)
	m.chatView.SetActiveChat(newChat.ID)
	m.modal = ModalNone
	return *m, nil
}
