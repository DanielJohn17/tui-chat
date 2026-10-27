package chat

import (
	"github.com/DanielJohn17/tui-chat/internal/tui/client"
	"github.com/DanielJohn17/tui-chat/internal/tui/theme"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	client       client.Client
	activeChatID int64
	activeTab    ChatTab
	peerHasKey   bool
	viewport     viewport.Model
	input        textinput.Model
	isFocused    bool
	width        int
	height       int

	convStatus   map[int64]FetchStatus
	convErrors   map[int64]string
	spinnerFrame string
}

func New(c client.Client) Model {
	vp := viewport.New(60, 18)
	vp.SetContent("Select a conversation to begin chatting.")

	ti := textinput.New()
	ti.Placeholder = "Type a message... (Press Enter to send, Esc to unfocus)"
	ti.Prompt = "❯ "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(theme.ColorCyan).Bold(true)
	ti.TextStyle = lipgloss.NewStyle().Foreground(theme.ColorWhite)
	ti.CharLimit = 500

	return Model{
		client:       c,
		activeTab:    TabCloud,
		peerHasKey:   true,
		viewport:     vp,
		input:        ti,
		isFocused:    false,
		width:        60,
		height:       24,
		convStatus:   make(map[int64]FetchStatus),
		convErrors:   make(map[int64]string),
		spinnerFrame: "⠋",
	}
}

func (m *Model) SetFetchStatus(chatID int64, status FetchStatus) {
	if m.convStatus == nil {
		m.convStatus = make(map[int64]FetchStatus)
	}
	m.convStatus[chatID] = status
	if status != StatusFailed && m.convErrors != nil {
		delete(m.convErrors, chatID)
	}
	m.RefreshMessages()
}

func (m *Model) SetFetchError(chatID int64, err string) {
	if m.convStatus == nil {
		m.convStatus = make(map[int64]FetchStatus)
	}
	if m.convErrors == nil {
		m.convErrors = make(map[int64]string)
	}
	m.convStatus[chatID] = StatusFailed
	m.convErrors[chatID] = err
	m.RefreshMessages()
}

func (m Model) FetchStatus(chatID int64) FetchStatus {
	if m.convStatus == nil {
		return StatusIdle
	}
	return m.convStatus[chatID]
}

func (m Model) FetchError(chatID int64) string {
	if m.convErrors == nil {
		return ""
	}
	return m.convErrors[chatID]
}

func (m *Model) SetSpinnerFrame(frame string) {
	m.spinnerFrame = frame
	if m.activeChatID != 0 && m.convStatus != nil && m.convStatus[m.activeChatID] == StatusLoading {
		messages := m.client.Messages(m.activeChatID)
		if len(messages) == 0 {
			m.RefreshMessages()
		}
	}
}

func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h

	vpWidth := w - 4
	if vpWidth < 1 {
		vpWidth = 1
	}
	m.viewport.Width = vpWidth
	m.updateViewportHeight()

	inputW := vpWidth - 4
	if inputW < 1 {
		inputW = 1
	}
	m.input.Width = inputW
	m.RefreshMessages()
}

func (m Model) ActiveTab() ChatTab {
	return m.activeTab
}

func (m *Model) SetActiveTab(tab ChatTab) {
	m.activeTab = tab
	m.updateViewportHeight()
	m.updateInputPlaceholder()
}

func (m Model) PeerHasKey() bool {
	return m.peerHasKey
}

func (m *Model) SetPeerKeyStatus(hasKey bool) {
	m.peerHasKey = hasKey
	if m.activeTab == TabSecret && !hasKey {
		m.BlurInput()
	}
	m.updateInputPlaceholder()
}

func (m *Model) SetActiveChat(chatID int64) {
	if m.activeChatID != chatID {
		m.activeChatID = chatID
		m.RefreshMessages()
		m.viewport.GotoBottom()
	}
}

func (m Model) ActiveChatID() int64 {
	return m.activeChatID
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.MouseMsg:
		if msg.Button == tea.MouseButtonWheelUp || msg.Type == tea.MouseWheelUp {
			m.viewport.LineUp(3)
			return m, nil
		}
		if msg.Button == tea.MouseButtonWheelDown || msg.Type == tea.MouseWheelDown {
			m.viewport.LineDown(3)
			return m, nil
		}
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			// Check if click was in the bottom input area (last 4 lines of chat pane)
			if msg.Y >= m.height-5 {
				m.FocusInput()
				return m, nil
			}
		}
	}

	if m.isFocused {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if newM, keyCmd, handled := m.handleInputKey(keyMsg); handled {
				return newM, keyCmd
			}
		}
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+s", "tab":
				return m, func() tea.Msg { return ToggleTabMsg{} }
			}
		}
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View(activeChat *client.Chat) string {
	contentWidth := m.width - 4
	if contentWidth < 1 {
		contentWidth = 1
	}

	header := renderHeader(activeChat, m.isFocused, contentWidth)
	tabBar := renderTabBar(m.activeTab, contentWidth)
	warnBanner := renderSecretWarningBanner(m.activeTab, m.peerHasKey, contentWidth)
	vpBox := m.viewport.View()
	inputRendered := renderInputBox(m.input.View(), m.isFocused, contentWidth)

	boxWidth := m.width - 2
	if boxWidth < 1 {
		boxWidth = 1
	}
	boxHeight := m.height - 2
	if boxHeight < 1 {
		boxHeight = 1
	}

	var paneContent string
	if m.activeTab == TabSecret {
		paneContent = lipgloss.JoinVertical(lipgloss.Left,
			header,
			tabBar,
			warnBanner,
			"",
			vpBox,
			"",
			inputRendered,
		)
	} else {
		paneContent = lipgloss.JoinVertical(lipgloss.Left,
			header,
			tabBar,
			"",
			vpBox,
			"",
			inputRendered,
		)
	}

	pane := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorBorderDim).
		Width(boxWidth).
		Height(boxHeight).
		Padding(0, 1).
		Render(paneContent)

	return pane
}
