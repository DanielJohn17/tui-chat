package tui

import (
	"fmt"
	"time"

	"github.com/DanielJohn17/tui-chat/internal/tui/client"
	"github.com/DanielJohn17/tui-chat/internal/tui/views/auth"
	"github.com/DanielJohn17/tui-chat/internal/tui/views/chat"
	"github.com/DanielJohn17/tui-chat/internal/tui/views/modals"
	"github.com/DanielJohn17/tui-chat/internal/tui/views/profile"
	"github.com/DanielJohn17/tui-chat/internal/tui/views/sidebar"
	"github.com/DanielJohn17/tui-chat/internal/tui/views/statusbar"
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

type ChatsLoadedMsg struct {
	Chats []client.Chat
	Err   error
}

type MessagesLoadedMsg struct {
	ChatID   int64
	Messages []client.Message
	Err      error
}

type WSIncomingMsg struct {
	Event any
}

type WSConnectResultMsg struct {
	Err error
}

type ReconnectTickMsg struct{}

type RetryCountdownTickMsg struct{}

type SpinnerTickMsg struct{}

type ClearNotificationMsg struct {
	ID int64
}

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type BulkMessagesLoadedMsg struct {
	MessagesByConv map[int64][]client.Message
	Err            error
}

func fetchBulkMessagesCmd(c client.Client) tea.Cmd {
	return func() tea.Msg {
		res, err := c.FetchBulkMessages()
		return BulkMessagesLoadedMsg{MessagesByConv: res, Err: err}
	}
}

func fetchChatsCmd(c client.Client) tea.Cmd {
	return func() tea.Msg {
		chats, err := c.FetchChats()
		return ChatsLoadedMsg{Chats: chats, Err: err}
	}
}

func fetchMessagesCmd(c client.Client, chatID int64) tea.Cmd {
	return func() tea.Msg {
		msgs, err := c.FetchMessages(chatID)
		return MessagesLoadedMsg{ChatID: chatID, Messages: msgs, Err: err}
	}
}

func connectWSCmd(c client.Client, eventsChan chan any) tea.Cmd {
	return func() tea.Msg {
		err := c.ConnectWS(eventsChan)
		return WSConnectResultMsg{Err: err}
	}
}

type AppModel struct {
	client       client.Client
	state        AppState
	modal        ModalType
	width        int
	height       int
	wsEventsChan chan any

	connStatus       statusbar.ConnectionStatus
	reconnectDelay   time.Duration
	retrySecondsLeft int
	spinnerIdx       int

	notificationText string
	notificationID   int64

	authView    auth.Model
	sidebarView *sidebar.Model
	chatView    chat.Model
	profileView profile.Model
	newDMView   modals.NewDMModel
}

func NewApp(c client.Client) AppModel {
	initialState := StateAuth
	connStatus := statusbar.StatusOffline
	if c.IsAuthenticated() {
		initialState = StateChat
		connStatus = statusbar.StatusConnecting
	}

	authV := auth.New(c)
	sidebarV := sidebar.New(c)
	chatV := chat.New(c)
	profileV := profile.New(c)
	newDMV := modals.NewDM()
	wsChan := make(chan any, 64)

	return AppModel{
		client:           c,
		state:            initialState,
		modal:            ModalNone,
		wsEventsChan:     wsChan,
		connStatus:       connStatus,
		reconnectDelay:   time.Second,
		retrySecondsLeft: 0,
		spinnerIdx:       0,
		notificationText: "",
		notificationID:   0,
		authView:         authV,
		sidebarView:      &sidebarV,
		chatView:         chatV,
		profileView:      profileV,
		newDMView:        newDMV,
	}
}

func (m AppModel) listenWSCmd() tea.Cmd {
	return func() tea.Msg {
		if m.wsEventsChan == nil {
			return nil
		}
		evt, ok := <-m.wsEventsChan
		if !ok {
			return nil
		}
		return WSIncomingMsg{Event: evt}
	}
}

func (m AppModel) Init() tea.Cmd {
	cmds := []tea.Cmd{
		tea.EnterAltScreen,
		m.authView.Init(),
	}
	if m.client.IsAuthenticated() {
		m.connStatus = statusbar.StatusConnecting
		cmds = append(cmds,
			fetchChatsCmd(m.client),
			connectWSCmd(m.client, m.wsEventsChan),
			m.listenWSCmd(),
			tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg { return SpinnerTickMsg{} }),
		)
	}
	return tea.Batch(cmds...)
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

	sidebarWidth := 38
	if w < 100 {
		sidebarWidth = 30
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
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.handleResize(msg.Width, msg.Height)
		return m, nil

	case auth.AuthSuccessMsg:
		m.state = StateChat
		m.client.SetProfile(msg.Profile)
		m.profileView.ReloadProfile()
		m.connStatus = statusbar.StatusConnecting
		m.reconnectDelay = time.Second
		m.retrySecondsLeft = 0
		return m, tea.Batch(
			fetchChatsCmd(m.client),
			connectWSCmd(m.client, m.wsEventsChan),
			m.listenWSCmd(),
			tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg { return SpinnerTickMsg{} }),
		)

	case WSConnectResultMsg:
		return m.handleWSConnectResult(msg)

	case RetryCountdownTickMsg:
		return m.handleRetryCountdownTick()

	case SpinnerTickMsg:
		return m.handleSpinnerTick()

	case ReconnectTickMsg:
		return m.handleReconnectTick()

	case ChatsLoadedMsg:
		return m.handleChatsLoaded(msg)

	case MessagesLoadedMsg:
		return m.handleMessagesLoaded(msg)

	case BulkMessagesLoadedMsg:
		return m.handleBulkMessagesLoaded(msg)

	case WSIncomingMsg:
		return m.handleWSIncoming(msg)

	case ClearNotificationMsg:
		if msg.ID == m.notificationID {
			m.notificationText = ""
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

	case profile.LogoutMsg:
		_ = m.client.Logout()
		m.state = StateAuth
		m.authView.SetErrorMessage("Logged out successfully.")
		return m, nil
	}

	// 1. Modals capture input when open
	if model, cmd, handled := m.handleModalUpdate(msg); handled {
		return model, cmd
	}

	// 2. State-specific updates
	switch m.state {
	case StateAuth:
		return m.handleAuthState(msg)

	case StateProfile:
		return m.handleProfileState(msg)

	case StateChat:
		return m.handleChatState(msg)
	}

	return m, nil
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
		currentConnStatus := m.connStatus
		if m.client.IsWSConnected() {
			currentConnStatus = statusbar.StatusConnected
		} else if currentConnStatus == statusbar.StatusConnected {
			currentConnStatus = statusbar.StatusOffline
		}
		spinnerFrame := spinnerFrames[m.spinnerIdx%len(spinnerFrames)]
		statusBar := statusbar.Render(activeChat, m.chatView.IsInputFocused(), currentConnStatus, m.retrySecondsLeft, spinnerFrame, m.notificationText, m.width)
		return lipgloss.JoinVertical(lipgloss.Left, mainRow, statusBar)
	}

	return fmt.Sprintf("Unknown state: %d", m.state)
}
