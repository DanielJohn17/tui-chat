package tui

import (
	"github.com/DanielJohn17/tui-chat/internal/tui/views/chat"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *AppModel) handleChatStateMouse(mouseMsg tea.MouseMsg, sidebarWidth, mainHeight int) (tea.Model, tea.Cmd, bool) {
	// 1. Mouse Wheel handling
	if mouseMsg.Button == tea.MouseButtonWheelUp || mouseMsg.Type == tea.MouseWheelUp {
		if mouseMsg.X < sidebarWidth {
			m.sidebarView.MoveUp()
			if sel := m.sidebarView.SelectedChat(); sel != nil {
				return *m, m.selectChat(sel), true
			}
		} else {
			m.chatView.ScrollUp(3)
		}
		return *m, nil, true
	}

	if mouseMsg.Button == tea.MouseButtonWheelDown || mouseMsg.Type == tea.MouseWheelDown {
		if mouseMsg.X < sidebarWidth {
			m.sidebarView.MoveDown()
			if sel := m.sidebarView.SelectedChat(); sel != nil {
				return *m, m.selectChat(sel), true
			}
		} else {
			m.chatView.ScrollDown(3)
		}
		return *m, nil, true
	}

	// 2. Mouse Left Click handling
	isClick := (mouseMsg.Action == tea.MouseActionPress && mouseMsg.Button == tea.MouseButtonLeft) ||
		(mouseMsg.Type == tea.MouseLeft && mouseMsg.Action != tea.MouseActionMotion)

	if !isClick {
		return nil, nil, false
	}

	// Click in status bar at bottom
	if mouseMsg.Y >= mainHeight {
		if mouseMsg.X >= m.width-12 {
			_ = m.client.CloseWS()
			return *m, tea.Quit, true
		}
		normX := float64(mouseMsg.X) / float64(m.width)
		if normX >= 0.40 && normX < 0.50 {
			m.chatView.FocusInput()
			return *m, nil, true
		} else if normX >= 0.50 && normX < 0.60 {
			m.openDiscovery()
			return *m, nil, true
		} else if normX >= 0.60 && normX < 0.72 {
			m.state = StateProfile
			m.profileView.ReloadProfile()
			return *m, nil, true
		} else if normX >= 0.72 && normX < 0.85 {
			m.modal = ModalHelp
			return *m, nil, true
		}
		return *m, nil, true
	}

	// Click in sidebar
	if mouseMsg.X < sidebarWidth && mouseMsg.Y < mainHeight {
		clickedChat, clickedNewDM, clickedProfile := m.sidebarView.HandleClick(mouseMsg.X, mouseMsg.Y)
		if clickedNewDM {
			m.openDiscovery()
			return *m, nil, true
		}
		if clickedProfile {
			m.state = StateProfile
			m.profileView.ReloadProfile()
			return *m, nil, true
		}
		if clickedChat != nil {
			m.chatView.BlurInput()
			cmd := m.selectChat(clickedChat)
			return *m, cmd, true
		}
		return *m, nil, true
	}

	// Click in chat pane
	if mouseMsg.X >= sidebarWidth && mouseMsg.Y < mainHeight {
		activeID := m.chatView.ActiveChatID()
		if activeID != 0 && m.chatView.FetchStatus(activeID) == chat.StatusFailed {
			m.chatView.SetFetchStatus(activeID, chat.StatusLoading)
			return *m, fetchMessagesCmd(m.client, activeID), true
		}
		// Click on tab bar (row 2 in chat pane)
		if mouseMsg.Y == 2 {
			xRel := mouseMsg.X - sidebarWidth
			sel := m.sidebarView.SelectedChat()
			if sel != nil {
				if xRel >= 1 && xRel <= 15 {
					model, cmd := m.switchChatTab(sel, chat.TabCloud)
					return model, cmd, true
				} else if xRel >= 16 && xRel <= 32 {
					model, cmd := m.switchChatTab(sel, chat.TabSecret)
					return model, cmd, true
				}
			}
		}
		if mouseMsg.Y >= mainHeight-5 {
			m.chatView.FocusInput()
			return *m, nil, true
		}
		if mouseMsg.Y <= 2 && m.chatView.IsInputFocused() {
			m.chatView.BlurInput()
			return *m, nil, true
		}
	}

	return nil, nil, false
}
