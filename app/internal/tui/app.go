package tui

import (
	"fmt"

	"github.com/DanielJohn17/tui-chat/app/internal/tui/client"
	"github.com/DanielJohn17/tui-chat/app/internal/tui/views/auth"
	"github.com/DanielJohn17/tui-chat/app/internal/tui/views/chat"
	"github.com/DanielJohn17/tui-chat/app/internal/tui/views/modals"
	"github.com/DanielJohn17/tui-chat/app/internal/tui/views/profile"
	"github.com/DanielJohn17/tui-chat/app/internal/tui/views/sidebar"
	"github.com/DanielJohn17/tui-chat/app/internal/tui/views/statusbar"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type AppState int

const (
	StateAuth AppState = iota
	StateChat
	StateProfile
)

type ModalType int

const (
	ModalNone ModalType = iota
	ModalHelp
	ModalNewDM
)

type AppModel struct {
	client client.Client
	state  AppState
	modal  ModalType
	width  int
	height int

	authView    auth.Model
	sidebarView *sidebar.Model
	chatView    chat.Model
	profileView profile.Model
	newDMView   modals.NewDMModel
}

func NewApp(c client.Client) AppModel {
	initialState := StateAuth
	if c.IsAuthenticated() {
		initialState = StateChat
	}

	authV := auth.New(c)
	sidebarV := sidebar.New(c)
	chatV := chat.New(c)
	profileV := profile.New(c)
	newDMV := modals.NewDM()

	return AppModel{
		client:      c,
		state:       initialState,
		modal:       ModalNone,
		authView:    authV,
		sidebarView: &sidebarV,
		chatView:    chatV,
		profileView: profileV,
		newDMView:   newDMV,
	}
}

func (m AppModel) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		m.authView.Init(),
	)
}

func (m *AppModel) handleResize(w, h int) {
	m.width = w
	m.height = h

	m.authView.SetSize(w, h)
	m.profileView.SetSize(w, h)

	mainHeight := h - 3
	if mainHeight < 10 {
		mainHeight = 10
	}

	sidebarWidth := 32
	if w < 90 {
		sidebarWidth = 28
	}
	chatWidth := w - sidebarWidth
	if chatWidth < 30 {
		chatWidth = 30
	}

	m.sidebarView.SetSize(sidebarWidth, mainHeight)
	m.chatView.SetSize(chatWidth, mainHeight)
	m.newDMView.SetSize(w, h)
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.handleResize(msg.Width, msg.Height)
		return m, nil

	case auth.AuthSuccessMsg:
		m.state = StateChat
		m.client.SetProfile(msg.Profile)
		m.profileView.ReloadProfile()
		if chat := m.sidebarView.SelectedChat(); chat != nil {
			m.chatView.SetActiveChat(chat.ID)
		}
		return m, nil

	case auth.OpenHelpMsg:
		m.modal = ModalHelp
		return m, nil

	case profile.BackToChatMsg:
		m.state = StateChat
		return m, nil

	case profile.ProfileSavedMsg:
		m.client.SetProfile(msg.Profile)
		m.state = StateChat
		return m, nil
	}

	// 1. If a modal is open, modal captures input
	if m.modal != ModalNone {
		switch m.modal {
		case ModalHelp:
			if mouseMsg, ok := msg.(tea.MouseMsg); ok {
				if mouseMsg.Action == tea.MouseActionPress && mouseMsg.Button == tea.MouseButtonLeft {
					// Clicking anywhere dismisses help
					m.modal = ModalNone
					return m, nil
				}
			}
			if keyMsg, ok := msg.(tea.KeyMsg); ok {
				switch keyMsg.String() {
				case "esc", "enter", "?", "q", "f1", "ctrl+h":
					m.modal = ModalNone
					return m, nil
				}
			}
			return m, nil

		case ModalNewDM:
			if mouseMsg, ok := msg.(tea.MouseMsg); ok {
				if mouseMsg.Action == tea.MouseActionPress && mouseMsg.Button == tea.MouseButtonLeft {
					// Check if clicked buttons inside NewDM modal
					boxW := 48
					boxH := 10
					startX := (m.width - boxW) / 2
					startY := (m.height - boxH) / 2
					relY := mouseMsg.Y - startY
					relX := mouseMsg.X - startX

					// Connect vs Cancel buttons are around relY == 6 or 7
					if relY >= 5 && relY <= 8 && relX >= 0 && relX < boxW {
						if relX < boxW/2 {
							// Connect
							username := m.newDMView.Value()
							if len(username) < 3 {
								m.newDMView.SetError("Username must be at least 3 characters")
								return m, nil
							}
							newChat := m.client.AddChat(username, username)
							m.sidebarView.MoveDown()
							m.chatView.SetActiveChat(newChat.ID)
							m.modal = ModalNone
							return m, nil
						} else {
							// Cancel
							m.modal = ModalNone
							return m, nil
						}
					}
				}
			}
			if keyMsg, ok := msg.(tea.KeyMsg); ok {
				switch keyMsg.String() {
				case "esc":
					m.modal = ModalNone
					return m, nil
				case "enter":
					username := m.newDMView.Value()
					if len(username) < 3 {
						m.newDMView.SetError("Username must be at least 3 characters")
						return m, nil
					}
					newChat := m.client.AddChat(username, username)
					m.sidebarView.MoveDown()
					m.chatView.SetActiveChat(newChat.ID)
					m.modal = ModalNone
					return m, nil
				}
			}
			var cmd tea.Cmd
			m.newDMView, cmd = m.newDMView.Update(msg)
			return m, cmd
		}
	}

	// 2. State-specific updates
	switch m.state {
	case StateAuth:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "esc", "ctrl+c":
				return m, tea.Quit
			case "f1", "ctrl+h", "?":
				m.modal = ModalHelp
				return m, nil
			}
		}
		var cmd tea.Cmd
		m.authView, cmd = m.authView.Update(msg)
		return m, cmd

	case StateProfile:
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "esc" {
			m.state = StateChat
			return m, nil
		}
		var cmd tea.Cmd
		m.profileView, cmd = m.profileView.Update(msg)
		return m, cmd

	case StateChat:
		sidebarWidth := 32
		if m.width < 90 {
			sidebarWidth = 28
		}
		mainHeight := m.height - 3

		// Mouse event handling across chat, sidebar, and statusbar
		if mouseMsg, ok := msg.(tea.MouseMsg); ok {
			// Mouse Wheel handling
			if mouseMsg.Button == tea.MouseButtonWheelUp || mouseMsg.Type == tea.MouseWheelUp {
				if mouseMsg.X < sidebarWidth {
					m.sidebarView.MoveUp()
					if sel := m.sidebarView.SelectedChat(); sel != nil {
						m.chatView.SetActiveChat(sel.ID)
					}
				} else {
					m.chatView.ScrollUp(3)
				}
				return m, nil
			}
			if mouseMsg.Button == tea.MouseButtonWheelDown || mouseMsg.Type == tea.MouseWheelDown {
				if mouseMsg.X < sidebarWidth {
					m.sidebarView.MoveDown()
					if sel := m.sidebarView.SelectedChat(); sel != nil {
						m.chatView.SetActiveChat(sel.ID)
					}
				} else {
					m.chatView.ScrollDown(3)
				}
				return m, nil
			}

			// Mouse Left Click handling
			isClick := (mouseMsg.Action == tea.MouseActionPress && mouseMsg.Button == tea.MouseButtonLeft) ||
				(mouseMsg.Type == tea.MouseLeft && mouseMsg.Action != tea.MouseActionMotion)

			if isClick {
				// 1. Click in status bar at bottom
				if mouseMsg.Y >= mainHeight {
					// Approximate sections across the status bar
					// shortcuts: "j/k: select • i: write • n: new DM • p: profile • ?: help • q: quit"
					// Click near right edge -> quit
					if mouseMsg.X >= m.width-12 {
						return m, tea.Quit
					}
					// Click in shortcuts area
					normX := float64(mouseMsg.X) / float64(m.width)
					if normX >= 0.40 && normX < 0.50 {
						// i: write
						m.chatView.FocusInput()
						return m, nil
					} else if normX >= 0.50 && normX < 0.60 {
						// n: new DM
						m.modal = ModalNewDM
						m.newDMView.Reset()
						return m, nil
					} else if normX >= 0.60 && normX < 0.72 {
						// p: profile
						m.state = StateProfile
						m.profileView.ReloadProfile()
						return m, nil
					} else if normX >= 0.72 && normX < 0.85 {
						// ?: help
						m.modal = ModalHelp
						return m, nil
					}
					return m, nil
				}

				// 2. Click in sidebar
				if mouseMsg.X < sidebarWidth && mouseMsg.Y < mainHeight {
					clickedChat, clickedNewDM, clickedProfile := m.sidebarView.HandleClick(mouseMsg.X, mouseMsg.Y)
					if clickedNewDM {
						m.modal = ModalNewDM
						m.newDMView.Reset()
						return m, nil
					}
					if clickedProfile {
						m.state = StateProfile
						m.profileView.ReloadProfile()
						return m, nil
					}
					if clickedChat != nil {
						m.chatView.SetActiveChat(clickedChat.ID)
						m.chatView.BlurInput()
						return m, nil
					}
					return m, nil
				}

				// 3. Click in chat pane
				if mouseMsg.X >= sidebarWidth && mouseMsg.Y < mainHeight {
					// Click in bottom input area -> focus
					if mouseMsg.Y >= mainHeight-5 {
						m.chatView.FocusInput()
						return m, nil
					}
					// Click in header -> unfocus if focused
					if mouseMsg.Y <= 2 && m.chatView.IsInputFocused() {
						m.chatView.BlurInput()
						return m, nil
					}
				}
			}
		}

		// When input is focused, route keystrokes directly to chatView
		if m.chatView.IsInputFocused() {
			var cmd tea.Cmd
			m.chatView, cmd = m.chatView.Update(msg)
			return m, cmd
		}

		// When input is NOT focused, top-level navigation keys apply
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit

			case "j", "down":
				m.sidebarView.MoveDown()
				if sel := m.sidebarView.SelectedChat(); sel != nil {
					m.chatView.SetActiveChat(sel.ID)
				}
				return m, nil

			case "k", "up":
				m.sidebarView.MoveUp()
				if sel := m.sidebarView.SelectedChat(); sel != nil {
					m.chatView.SetActiveChat(sel.ID)
				}
				return m, nil

			case "i", "enter":
				m.chatView.FocusInput()
				return m, nil

			case "n":
				m.modal = ModalNewDM
				m.newDMView.Reset()
				return m, nil

			case "p":
				m.state = StateProfile
				m.profileView.ReloadProfile()
				return m, nil

			case "?", "ctrl+h", "f1":
				m.modal = ModalHelp
				return m, nil

			case "pgup", "pgdown":
				var cmd tea.Cmd
				m.chatView, cmd = m.chatView.Update(msg)
				return m, cmd
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m AppModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing terminal..."
	}

	// Active modal overlays (renders over either Auth or Chat state)
	if m.modal != ModalNone {
		switch m.modal {
		case ModalHelp:
			return modals.RenderHelp(m.width, m.height)
		case ModalNewDM:
			return m.newDMView.View(m.width, m.height)
		}
	}

	// State views
	switch m.state {
	case StateAuth:
		return m.authView.View()

	case StateProfile:
		return m.profileView.View()

	case StateChat:
		activeChat := m.sidebarView.SelectedChat()
		mainRow := lipgloss.JoinHorizontal(lipgloss.Top,
			m.sidebarView.View(),
			m.chatView.View(activeChat),
		)
		statusBar := statusbar.Render(activeChat, m.chatView.IsInputFocused(), m.width)
		return lipgloss.JoinVertical(lipgloss.Left, mainRow, statusBar)
	}

	return fmt.Sprintf("Unknown state: %d", m.state)
}
