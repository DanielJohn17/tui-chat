package auth

import (
	"strings"
	"unicode"

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

	// Login inputs
	lu := textinput.New()
	lu.Placeholder = "Enter your username"
	lu.CharLimit = 20
	lu.Focus()
	lu.Prompt = "  "
	lu.TextStyle = lipgloss.NewStyle().Foreground(theme.ColorWhite)

	lp := textinput.New()
	lp.Placeholder = "Enter your password"
	lp.EchoMode = textinput.EchoPassword
	lp.EchoCharacter = '•'
	lp.CharLimit = 50
	lp.Prompt = "  "
	lp.TextStyle = lipgloss.NewStyle().Foreground(theme.ColorWhite)

	// Register inputs
	rn := textinput.New()
	rn.Placeholder = "e.g. Alex Mercer"
	rn.CharLimit = 50
	rn.Prompt = "  "
	rn.TextStyle = lipgloss.NewStyle().Foreground(theme.ColorWhite)

	ru := textinput.New()
	ru.Placeholder = "e.g. alexm (3-20 chars)"
	ru.CharLimit = 20
	ru.Prompt = "  "
	ru.TextStyle = lipgloss.NewStyle().Foreground(theme.ColorWhite)

	rp := textinput.New()
	rp.Placeholder = "Must include A-Z and 0-9"
	rp.EchoMode = textinput.EchoPassword
	rp.EchoCharacter = '•'
	rp.CharLimit = 50
	rp.Prompt = "  "
	rp.TextStyle = lipgloss.NewStyle().Foreground(theme.ColorWhite)

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
		case "tab", "shift+tab", "up", "down":
			isDown := msg.String() == "tab" || msg.String() == "down"
			maxIndex := 1 // Login has 2 fields (0, 1)
			if m.mode == ModeRegister {
				maxIndex = 2 // Register has 3 fields (0, 1, 2)
			}

			if isDown {
				m.focusIndex = (m.focusIndex + 1) % (maxIndex + 1)
			} else {
				m.focusIndex = (m.focusIndex - 1 + maxIndex + 1) % (maxIndex + 1)
			}
			m.updateFocus()
			return m, nil

		case "ctrl+t":
			// Toggle between login and register
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

func hasUpperAndDigit(s string) (bool, bool) {
	var hasUpper, hasDigit bool
	for _, r := range s {
		if unicode.IsUpper(r) {
			hasUpper = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}
	return hasUpper, hasDigit
}

func (m *Model) submit() tea.Cmd {
	if m.mode == ModeLogin {
		username := strings.TrimSpace(m.loginUsername.Value())
		password := m.loginPassword.Value()

		if len(username) < 3 || len(username) > 20 {
			m.errorMessage = "Username must be between 3 and 20 characters."
			return nil
		}
		hasUp, hasDig := hasUpperAndDigit(password)
		if !hasUp || !hasDig {
			m.errorMessage = "Password requires at least one uppercase letter (A-Z) and one number (0-9)."
			return nil
		}

		m.loading = true
		return tea.Batch(m.spinner.Tick, func() tea.Msg {
			prof, err := m.client.Login(username, password)
			if err != nil {
				return AuthErrorMsg{Err: err}
			}
			return AuthSuccessMsg{Profile: *prof}
		})
	}

	// Register validation
	name := strings.TrimSpace(m.regName.Value())
	username := strings.TrimSpace(m.regUsername.Value())
	password := m.regPassword.Value()

	if len(name) < 3 || len(name) > 50 {
		m.errorMessage = "Display name must be between 3 and 50 characters."
		return nil
	}
	if len(username) < 3 || len(username) > 20 {
		m.errorMessage = "Username must be between 3 and 20 characters."
		return nil
	}
	hasUp, hasDig := hasUpperAndDigit(password)
	if !hasUp || !hasDig {
		m.errorMessage = "Password requires at least one uppercase letter (A-Z) and one number (0-9)."
		return nil
	}

	m.loading = true
	return tea.Batch(m.spinner.Tick, func() tea.Msg {
		prof, err := m.client.Register(name, username, password)
		if err != nil {
			return AuthErrorMsg{Err: err}
		}
		return AuthSuccessMsg{Profile: *prof}
	})
}

func (m Model) View() string {
	boxWidth := 54
	if m.width > 0 && m.width < 58 {
		boxWidth = m.width - 4
	}

	// Title / Brand
	header := lipgloss.JoinVertical(lipgloss.Center,
		theme.StyleTitle.Render("◈ TUI CHAT SYSTEM ◈"),
		theme.StyleDim.Render("Secure Terminal Communication"),
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

	tabs := lipgloss.JoinHorizontal(lipgloss.Center, tabLogin, "  ", tabRegister)

	// Form inputs
	var formContent string
	if m.mode == ModeLogin {
		uBox := m.renderInputBox("Username", m.loginUsername.View(), m.focusIndex == 0, boxWidth-8)
		pBox := m.renderInputBox("Password", m.loginPassword.View(), m.focusIndex == 1, boxWidth-8)
		formContent = lipgloss.JoinVertical(lipgloss.Left, uBox, "", pBox)
	} else {
		nBox := m.renderInputBox("Display Name", m.regName.View(), m.focusIndex == 0, boxWidth-8)
		uBox := m.renderInputBox("Username (3-20 chars)", m.regUsername.View(), m.focusIndex == 1, boxWidth-8)
		pBox := m.renderInputBox("Password (A-Z, 0-9)", m.regPassword.View(), m.focusIndex == 2, boxWidth-8)
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
			Width(boxWidth - 6).
			Render("✖ " + m.errorMessage)
	}

	// Submit button or loading indicator
	var submitAction string
	if m.loading {
		submitAction = lipgloss.JoinHorizontal(lipgloss.Center,
			m.spinner.View(),
			"  ",
			theme.StyleSubtitle.Render("Authenticating with server..."),
		)
	} else {
		actionText := "[ Enter: Login ]"
		if m.mode == ModeRegister {
			actionText = "[ Enter: Create Account ]"
		}
		submitAction = theme.StyleSuccess.Render(actionText)
	}

	// Hotkeys helper footer
	footer := theme.StyleDim.Render("Tab: next field • Ctrl+T: switch login/register • Esc: quit")

	// Assemble card body
	bodyItems := []string{
		header,
		"",
		tabs,
		"",
	}
	if errorBanner != "" {
		bodyItems = append(bodyItems, errorBanner, "")
	}
	bodyItems = append(bodyItems, formContent, "", lipgloss.NewStyle().Width(boxWidth-6).Align(lipgloss.Center).Render(submitAction), "", footer)

	cardContent := lipgloss.JoinVertical(lipgloss.Center, bodyItems...)

	card := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorCyan).
		Background(theme.ColorPanelBg).
		Padding(1, 2).
		Width(boxWidth).
		Render(cardContent)

	// Center on screen
	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, card)
	}
	return card
}

func (m Model) renderInputBox(label, inputView string, focused bool, width int) string {
	lblStyle := theme.StyleDim
	borderCol := theme.ColorBorderDim
	if focused {
		lblStyle = theme.StyleSubtitle
		borderCol = theme.ColorCyan
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderCol).
		Width(width).
		Padding(0, 1).
		Render(inputView)

	return lipgloss.JoinVertical(lipgloss.Left,
		lblStyle.Render(label),
		box,
	)
}
