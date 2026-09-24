package chat

import (
	"fmt"
	"strings"

	"github.com/DanielJohn17/tui-chat/internal/tui/client"
	"github.com/DanielJohn17/tui-chat/internal/tui/theme"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type FetchStatus int

const (
	StatusIdle FetchStatus = iota
	StatusLoading
	StatusSuccess
	StatusFailed
)

type Model struct {
	client       client.Client
	activeChatID int64
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
	if vpWidth < 20 {
		vpWidth = 20
	}
	// Reserve 2 lines for header, 3 lines for input box, 1 blank line before input, 2 lines for outer borders
	vpHeight := h - 8
	if vpHeight < 4 {
		vpHeight = 4
	}

	m.viewport.Width = vpWidth
	m.viewport.Height = vpHeight
	m.input.Width = vpWidth - 4
	m.RefreshMessages()
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

func (m *Model) FocusInput() {
	m.isFocused = true
	m.input.Focus()
}

func (m *Model) BlurInput() {
	m.isFocused = false
	m.input.Blur()
}

func (m Model) IsInputFocused() bool {
	return m.isFocused
}

func (m *Model) RefreshMessages() {
	if m.activeChatID == 0 {
		m.viewport.SetContent("No conversation selected.")
		return
	}

	status := StatusIdle
	if m.convStatus != nil {
		status = m.convStatus[m.activeChatID]
	}

	messages := m.client.Messages(m.activeChatID)

	// Loading state when messages are not yet in memory
	if status == StatusLoading && len(messages) == 0 {
		sp := m.spinnerFrame
		if sp == "" {
			sp = "⠋"
		}
		loadingText := lipgloss.JoinHorizontal(lipgloss.Center,
			lipgloss.NewStyle().Foreground(theme.ColorYellow).Bold(true).Render(sp),
			" ",
			theme.StyleDim.Render("Loading conversation messages..."),
		)
		m.viewport.SetContent("\n\n  " + loadingText)
		return
	}

	// Error / Failure state when fetching failed
	if status == StatusFailed && len(messages) == 0 {
		errDetail := m.convErrors[m.activeChatID]
		errMsg := "Failed to load messages"
		if errDetail != "" {
			errMsg = fmt.Sprintf("Failed to load messages: %s", errDetail)
		}
		errTitle := lipgloss.NewStyle().Foreground(theme.ColorRed).Bold(true).Render("✖ " + errMsg)
		retryHint := theme.StyleDim.Render("Press [ r ] to retry  •  Click to reload")
		errBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.ColorRed).
			Padding(1, 2).
			Render(lipgloss.JoinVertical(lipgloss.Center, errTitle, "", retryHint))
		m.viewport.SetContent("\n  " + errBox)
		return
	}

	if len(messages) == 0 {
		m.viewport.SetContent(theme.StyleDim.Render("\n  No messages yet. Send a message to break the ice!"))
		return
	}

	var sb strings.Builder
	contentWidth := m.viewport.Width - 2
	if contentWidth < 20 {
		contentWidth = 20
	}

	for _, msg := range messages {
		var senderTag, row string
		timeTag := theme.StyleDim.Render(msg.Timestamp)

		if msg.Self {
			senderTag = lipgloss.NewStyle().Bold(true).Foreground(theme.ColorCyan).Render("You")
			header := fmt.Sprintf("%s  %s", senderTag, timeTag)
			body := lipgloss.NewStyle().
				Foreground(theme.ColorWhite).
				Width(contentWidth).
				Render(msg.Text)
			row = fmt.Sprintf("  %s\n  %s\n", header, body)
		} else {
			senderTag = lipgloss.NewStyle().Bold(true).Foreground(theme.ColorMagenta).Render(msg.Sender)
			header := fmt.Sprintf("%s  %s", senderTag, timeTag)
			body := lipgloss.NewStyle().
				Foreground(theme.ColorWhite).
				Width(contentWidth).
				Render(msg.Text)
			row = fmt.Sprintf("  %s\n  %s\n", header, body)
		}
		sb.WriteString(row + "\n")
	}

	m.viewport.SetContent(sb.String())
	m.viewport.GotoBottom()
}

func (m *Model) ScrollUp(lines int) {
	m.viewport.LineUp(lines)
}

func (m *Model) ScrollDown(lines int) {
	m.viewport.LineDown(lines)
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
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				m.BlurInput()
				return m, nil
			case "enter":
				text := strings.TrimSpace(m.input.Value())
				if text != "" && m.activeChatID != 0 {
					_ = m.client.SendWS(m.activeChatID, text)
					nowStr := "now"
					m.client.AppendMessage(client.Message{
						ConvID:    m.activeChatID,
						SenderID:  m.client.Profile().ID,
						Sender:    "You",
						Text:      text,
						Timestamp: nowStr,
						Self:      true,
					})
					m.client.UpdateChatSnippet(m.activeChatID, text, nowStr, 0, nil)
					m.input.Reset()
					m.RefreshMessages()
					m.viewport.GotoBottom()
				}
				return m, nil
			}
		}
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View(activeChat *client.Chat) string {
	contentWidth := m.width - 4
	if contentWidth < 20 {
		contentWidth = 20
	}

	// 1. Chat Header
	var header string
	if activeChat != nil {
		var dot string
		if activeChat.Online {
			dot = theme.StyleSuccess.Render("●")
		} else {
			dot = theme.StyleDim.Render("○")
		}

		leftTitle := lipgloss.JoinHorizontal(lipgloss.Center,
			dot, " ",
			theme.StyleSubtitle.Render(activeChat.Name), " ",
			lipgloss.NewStyle().Foreground(theme.ColorMagenta).Render("@"+activeChat.Username),
		)

		var focusHint string
		if m.isFocused {
			focusHint = theme.StyleDim.Render("[Esc: unfocus input]")
		} else {
			focusHint = theme.StyleDim.Render("[i / Enter: focus input]")
		}

		hSpaces := contentWidth - lipgloss.Width(leftTitle) - lipgloss.Width(focusHint)
		if hSpaces < 1 {
			hSpaces = 1
		}
		header = lipgloss.JoinHorizontal(lipgloss.Center, leftTitle, lipgloss.NewStyle().Width(hSpaces).Render(""), focusHint)
	} else {
		header = theme.StyleDim.Render("No active conversation")
	}

	// 2. Viewport
	vpBox := m.viewport.View()

	// 3. Input Box
	inputBorder := theme.ColorBorderDim
	if m.isFocused {
		inputBorder = theme.ColorCyan
	}
	inputBoxWidth := contentWidth - 2
	if inputBoxWidth < 10 {
		inputBoxWidth = 10
	}
	inputRendered := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(inputBorder).
		Width(inputBoxWidth).
		Padding(0, 1).
		Render(m.input.View())

	// Assemble Pane
	boxWidth := m.width - 2
	if boxWidth < 10 {
		boxWidth = 10
	}
	boxHeight := m.height - 2
	if boxHeight < 6 {
		boxHeight = 6
	}

	pane := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorBorderDim).
		Width(boxWidth).
		Height(boxHeight).
		Padding(0, 1).
		Render(lipgloss.JoinVertical(lipgloss.Left,
			header,
			"",
			vpBox,
			"",
			inputRendered,
		))

	return pane
}
