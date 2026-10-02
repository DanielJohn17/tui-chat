package tui

import (
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
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "esc":
			m.cancelDiscovery()
			m.modal = ModalNone
			return *m, nil
		case "ctrl+c":
			m.cancelDiscovery()
			return *m, tea.Quit
		case "enter":
			return m.submitNewDM()
		}
	}

	before := m.newDMView.Value()
	var cmd tea.Cmd
	m.newDMView, cmd = m.newDMView.Update(msg)
	if m.newDMView.Value() != before {
		return *m, tea.Batch(cmd, m.scheduleDiscovery())
	}
	return *m, cmd
}
