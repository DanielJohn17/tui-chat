package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m *AppModel) handleAuthState(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "esc", "ctrl+c":
			return *m, tea.Quit
		case "f1", "ctrl+h", "?":
			m.modal = ModalHelp
			return *m, nil
		}
	}
	var cmd tea.Cmd
	m.authView, cmd = m.authView.Update(msg)
	return *m, cmd
}

func (m *AppModel) handleProfileState(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "esc" {
		m.state = StateChat
		return *m, nil
	}
	var cmd tea.Cmd
	m.profileView, cmd = m.profileView.Update(msg)
	return *m, cmd
}

func (m *AppModel) handleChatState(msg tea.Msg) (tea.Model, tea.Cmd) {
	sidebarWidth := 38
	if m.width < 100 {
		sidebarWidth = 30
	}
	mainHeight := m.height - 3

	// Mouse event handling
	if mouseMsg, ok := msg.(tea.MouseMsg); ok {
		if model, cmd, handled := m.handleChatStateMouse(mouseMsg, sidebarWidth, mainHeight); handled {
			return model, cmd
		}
	}

	// When input is focused, route keystrokes directly to chatView
	if m.chatView.IsInputFocused() {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "ctrl+h", "f1":
				m.modal = ModalHelp
				return *m, nil
			}
		}
		var cmd tea.Cmd
		m.chatView, cmd = m.chatView.Update(msg)
		if m.chatView.ActiveChatID() != 0 {
			m.sidebarView.SelectChatByID(m.chatView.ActiveChatID())
		}
		return *m, cmd
	}

	// When input is NOT focused, top-level navigation keys apply
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "q", "ctrl+c":
			return *m, tea.Quit

		case "j", "down":
			m.sidebarView.MoveDown()
			if sel := m.sidebarView.SelectedChat(); sel != nil {
				m.chatView.SetActiveChat(sel.ID)
				_ = m.client.FocusConv(sel.ID)
				_ = m.client.MarkRead(sel.ID, 0)
				zero := 0
				m.client.UpdateChatSnippet(sel.ID, "", "", 0, &zero)
				return *m, fetchMessagesCmd(m.client, sel.ID)
			}
			return *m, nil

		case "k", "up":
			m.sidebarView.MoveUp()
			if sel := m.sidebarView.SelectedChat(); sel != nil {
				m.chatView.SetActiveChat(sel.ID)
				_ = m.client.FocusConv(sel.ID)
				_ = m.client.MarkRead(sel.ID, 0)
				zero := 0
				m.client.UpdateChatSnippet(sel.ID, "", "", 0, &zero)
				return *m, fetchMessagesCmd(m.client, sel.ID)
			}
			return *m, nil

		case "r":
			return m.triggerManualReconnect()

		case "i", "enter":
			m.chatView.FocusInput()
			return *m, nil

		case "n":
			m.modal = ModalNewDM
			m.newDMView.Reset()
			return *m, nil

		case "p":
			m.state = StateProfile
			m.profileView.ReloadProfile()
			return *m, nil

		case "?", "ctrl+h", "f1":
			m.modal = ModalHelp
			return *m, nil

		case "pgup", "pgdown":
			var cmd tea.Cmd
			m.chatView, cmd = m.chatView.Update(msg)
			return *m, cmd
		}
	}

	return *m, nil
}
