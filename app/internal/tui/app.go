package tui

import (
	"fmt"

	"github.com/DanielJohn17/tui-chat/app/internal/tui/client"
	"github.com/DanielJohn17/tui-chat/app/internal/tui/views/auth"
	"github.com/DanielJohn17/tui-chat/app/internal/tui/views/chat"
	"github.com/DanielJohn17/tui-chat/app/internal/tui/views/modals"
	"github.com/DanielJohn17/tui-chat/app/internal/tui/views/sidebar"
	"github.com/DanielJohn17/tui-chat/app/internal/tui/views/statusbar"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type AppState int

const (
	StateAuth AppState = iota
	StateChat
)

type ModalType int

const (
	ModalNone ModalType = iota
	ModalHelp
	ModalNewDM
	ModalProfile
)

type AppModel struct {
	client  client.Client
	state   AppState
	modal   ModalType
	width   int
	height  int

	authView    auth.Model
	sidebarView sidebar.Model
	chatView    chat.Model
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
	newDMV := modals.NewDM()

	return AppModel{
		client:      c,
		state:       initialState,
		modal:       ModalNone,
		authView:    authV,
		sidebarView: sidebarV,
		chatView:    chatV,
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
		if chat := m.sidebarView.SelectedChat(); chat != nil {
			m.chatView.SetActiveChat(chat.ID)
		}
		return m, nil

	case auth.OpenHelpMsg:
		m.modal = ModalHelp
		return m, nil
	}

	// 1. If a modal is open, modal captures input
	if m.modal != ModalNone {
		switch m.modal {
		case ModalHelp:
			if keyMsg, ok := msg.(tea.KeyMsg); ok {
				switch keyMsg.String() {
				case "esc", "enter", "h", "q", "f1", "ctrl+h":
					m.modal = ModalNone
					return m, nil
				}
			}
			return m, nil

		case ModalProfile:
			if keyMsg, ok := msg.(tea.KeyMsg); ok {
				switch keyMsg.String() {
				case "esc", "enter", "p", "q":
					m.modal = ModalNone
					return m, nil
				}
			}
			return m, nil

		case ModalNewDM:
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
			case "f1", "ctrl+h":
				m.modal = ModalHelp
				return m, nil
			}
		}
		var cmd tea.Cmd
		m.authView, cmd = m.authView.Update(msg)
		return m, cmd

	case StateChat:
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
				m.modal = ModalProfile
				return m, nil

			case "h", "?", "f1":
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
		case ModalProfile:
			return modals.RenderProfile(m.client.Profile(), m.width, m.height)
		case ModalNewDM:
			return m.newDMView.View(m.width, m.height)
		}
	}

	// State views
	switch m.state {
	case StateAuth:
		return m.authView.View()

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
