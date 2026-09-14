package auth

import (
	"strings"

	"github.com/DanielJohn17/tui-chat/app/internal/tui/client"
	"github.com/DanielJohn17/tui-chat/app/internal/tui/theme"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Mode int

const (
	ModeLogin Mode = iota
	ModeRegister
)

type AuthSuccessMsg struct {
	Profile client.Profile
}

type AuthErrorMsg struct {
	Err error
}

type OpenHelpMsg struct{}

type Model struct {
	client     client.Client
	mode       Mode
	focusIndex int
	width      int
	height     int

	// Login inputs
	loginUsername textinput.Model
	loginPassword textinput.Model

	// Register inputs
	regName     textinput.Model
	regUsername textinput.Model
	regPassword textinput.Model

	// State
	loading      bool
	spinner      spinner.Model
	errorMessage string
}

func New(c client.Client) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.ColorCyan)

	createInput := func(placeholder string, isPassword bool, charLimit int) textinput.Model {
		ti := textinput.New()
		ti.Placeholder = placeholder
		ti.Prompt = ""
		ti.CharLimit = charLimit
		ti.Width = 44
		ti.TextStyle = lipgloss.NewStyle().Foreground(theme.ColorWhite)
		ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(theme.ColorDimText)
		ti.Cursor.Style = lipgloss.NewStyle().Foreground(theme.ColorCyan)
		if isPassword {
			ti.EchoMode = textinput.EchoPassword
			ti.EchoCharacter = '•'
		}
		return ti
	}

	// Login inputs
	lu := createInput("Enter username (or press Enter to pass)", false, 20)
	lu.Focus()
	lp := createInput("Enter password (optional in dev)", true, 50)

	// Register inputs
	rn := createInput("e.g. Alex Mercer", false, 50)
	ru := createInput("e.g. alexm", false, 20)
	rp := createInput("Enter password (optional in dev)", true, 50)

	return Model{
		client:        c,
		mode:          ModeLogin,
		focusIndex:    0,
		loginUsername: lu,
		loginPassword: lp,
		regName:       rn,
		regUsername:   ru,
		regPassword:   rp,
		spinner:       s,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case AuthErrorMsg:
		m.loading = false
		m.errorMessage = msg.Err.Error()
		return m, nil

	case tea.KeyMsg:
		m.errorMessage = "" // Clear error on new keypress

		switch msg.String() {
		case "esc":
			return m, tea.Quit

		case "f1", "ctrl+h":
			return m, func() tea.Msg { return OpenHelpMsg{} }

		case "tab", "down":
			maxIndex := 1
			if m.mode == ModeRegister {
				maxIndex = 2
			}
			m.focusIndex = (m.focusIndex + 1) % (maxIndex + 1)
			m.updateFocus()
			return m, nil

		case "shift+tab", "up":
			maxIndex := 1
			if m.mode == ModeRegister {
				maxIndex = 2
			}
			m.focusIndex = (m.focusIndex - 1 + maxIndex + 1) % (maxIndex + 1)
			m.updateFocus()
			return m, nil

		case "ctrl+t":
			if m.mode == ModeLogin {
				m.mode = ModeRegister
			} else {
				m.mode = ModeLogin
			}
			m.focusIndex = 0
			m.updateFocus()
			return m, nil

		case "enter":
			if m.loading {
				return m, nil
			}
			return m, m.submit()
		}
	}

	// Update active input
	var cmd tea.Cmd
	if m.mode == ModeLogin {
		if m.focusIndex == 0 {
			m.loginUsername, cmd = m.loginUsername.Update(msg)
		} else {
			m.loginPassword, cmd = m.loginPassword.Update(msg)
		}
	} else {
		switch m.focusIndex {
		case 0:
			m.regName, cmd = m.regName.Update(msg)
		case 1:
			m.regUsername, cmd = m.regUsername.Update(msg)
		case 2:
			m.regPassword, cmd = m.regPassword.Update(msg)
		}
	}

	return m, cmd
}

func (m *Model) updateFocus() {
	if m.mode == ModeLogin {
		if m.focusIndex == 0 {
			m.loginUsername.Focus()
			m.loginPassword.Blur()
		} else {
			m.loginUsername.Blur()
			m.loginPassword.Focus()
		}
	} else {
		m.regName.Blur()
		m.regUsername.Blur()
		m.regPassword.Blur()
		switch m.focusIndex {
		case 0:
			m.regName.Focus()
		case 1:
			m.regUsername.Focus()
		case 2:
			m.regPassword.Focus()
		}
	}
}

func (m *Model) submit() tea.Cmd {
	if m.mode == ModeLogin {
		username := strings.TrimSpace(m.loginUsername.Value())
		if username == "" {
			username = "alex"
		}

		// Pass to the chat app directly without requiring an API call
		return func() tea.Msg {
			return AuthSuccessMsg{
				Profile: client.Profile{
					ID:        1,
					Name:      username,
					Username:  username,
					Token:     "mock-jwt-token-authenticated",
					CreatedAt: "Today",
				},
			}
		}
	}

	// Register mode - pass straight to the chat app
	name := strings.TrimSpace(m.regName.Value())
	if name == "" {
		name = "Alex Mercer"
	}
	username := strings.TrimSpace(m.regUsername.Value())
	if username == "" {
		username = "alexm"
	}

	return func() tea.Msg {
		return AuthSuccessMsg{
			Profile: client.Profile{
				ID:        1,
				Name:      name,
				Username:  username,
				Token:     "mock-jwt-token-registered",
				CreatedAt: "Today",
			},
		}
	}
}

func (m Model) View() string {
	cardWidth := 56
	if m.width > 0 && m.width < 60 {
		cardWidth = m.width - 4
	}
	if cardWidth < 44 {
		cardWidth = 44
	}

	innerWidth := cardWidth - 6  // inside padding(1, 2)
	inputWidth := innerWidth - 4 // inside input box margin/border

	m.loginUsername.Width = inputWidth - 2
	m.loginPassword.Width = inputWidth - 2
	m.regName.Width = inputWidth - 2
	m.regUsername.Width = inputWidth - 2
	m.regPassword.Width = inputWidth - 2

	// Title / Brand
	title := theme.StyleTitle.Render("◈ TUI CHAT SYSTEM ◈")
	subtitle := theme.StyleDim.Render("Secure Terminal Communication")
	header := lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(title),
		lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(subtitle),
	)

	// Mode selector tabs
	var tabLogin, tabRegister string
	if m.mode == ModeLogin {
		tabLogin = lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.ColorBlack).
			Background(theme.ColorCyan).
			Padding(0, 2).
			Render("1. LOGIN")
		tabRegister = lipgloss.NewStyle().
			Foreground(theme.ColorDimText).
			Padding(0, 2).
			Render("2. REGISTER")
	} else {
		tabLogin = lipgloss.NewStyle().
			Foreground(theme.ColorDimText).
			Padding(0, 2).
			Render("1. LOGIN")
		tabRegister = lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.ColorBlack).
			Background(theme.ColorCyan).
			Padding(0, 2).
			Render("2. REGISTER")
	}

	tabs := lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(
		lipgloss.JoinHorizontal(lipgloss.Center, tabLogin, "  ", tabRegister),
	)

	// Render input helper
	renderField := func(label string, inputView string, focused bool) string {
		lblStyle := theme.StyleDim
		borderCol := theme.ColorBorderDim
		if focused {
			lblStyle = theme.StyleSubtitle
			borderCol = theme.ColorCyan
		}

		box := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderCol).
			Width(inputWidth).
			Padding(0, 1).
			Render(inputView)

		return lipgloss.JoinVertical(lipgloss.Left,
			lblStyle.Render(label),
			box,
		)
	}

	// Form inputs
	var formContent string
	if m.mode == ModeLogin {
		uBox := renderField("Username", m.loginUsername.View(), m.focusIndex == 0)
		pBox := renderField("Password", m.loginPassword.View(), m.focusIndex == 1)
		formContent = lipgloss.JoinVertical(lipgloss.Left, uBox, "", pBox)
	} else {
		nBox := renderField("Display Name", m.regName.View(), m.focusIndex == 0)
		uBox := renderField("Username", m.regUsername.View(), m.focusIndex == 1)
		pBox := renderField("Password", m.regPassword.View(), m.focusIndex == 2)
		formContent = lipgloss.JoinVertical(lipgloss.Left, nBox, "", uBox, "", pBox)
	}

	// Error banner if any
	var errorBanner string
	if m.errorMessage != "" {
		errorBanner = lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.ColorRed).
			Background(lipgloss.Color("#2A0812")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.ColorRed).
			Padding(0, 1).
			Width(inputWidth).
			Render("✖ " + m.errorMessage)
	}

	// Submit button
	actionText := "[ Enter: Login & Enter Chat ]"
	if m.mode == ModeRegister {
		actionText = "[ Enter: Create Account & Enter Chat ]"
	}
	submitAction := theme.StyleSuccess.Render(actionText)
	submitActionCentered := lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(submitAction)

	// Clean 2-line footer helper with Help Desk and Quit shortcuts
	hint1 := theme.StyleDim.Render("Tab: next field  •  Ctrl+T: switch mode")
	hint2 := theme.StyleDim.Render("? / Ctrl+H / F1: help  •  Esc: quit")
	footer := lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(hint1),
		lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(hint2),
	)

	// Assemble card body cleanly without erratic line wrapping
	bodyItems := []string{
		header,
		"",
		tabs,
		"",
	}
	if errorBanner != "" {
		bodyItems = append(bodyItems, errorBanner, "")
	}
	bodyItems = append(bodyItems,
		formContent,
		"",
		submitActionCentered,
		"",
		footer,
	)

	cardContent := lipgloss.JoinVertical(lipgloss.Left, bodyItems...)

	card := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorCyan).
		Padding(1, 2).
		Width(cardWidth).
		Render(cardContent)

	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, card)
	}
	return card
}
