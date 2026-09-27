package auth

import (
	"fmt"
	"strings"

	"github.com/DanielJohn17/tui-chat/internal/tui/client"
	"github.com/DanielJohn17/tui-chat/internal/tui/theme"
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

func (m *Model) SetErrorMessage(msg string) {
	m.errorMessage = CleanAuthErrorString(msg)
	m.loading = false
}

func (m Model) Mode() Mode {
	return m.mode
}

func (m Model) FocusIndex() int {
	return m.focusIndex
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
		m.errorMessage = CleanAuthError(msg.Err)
		return m, nil

	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			cardWidth := 56
			if m.width > 0 && m.width < 60 {
				cardWidth = m.width - 4
			}
			if cardWidth < 44 {
				cardWidth = 44
			}
			cardW := cardWidth + 6

			cardH := 25
			if m.mode == ModeRegister {
				cardH = 33
			}
			if m.errorMessage != "" {
				cardH += 4
			}

			startX := (m.width - cardW) / 2
			startY := (m.height - cardH) / 2
			if startX < 0 {
				startX = 0
			}
			if startY < 0 {
				startY = 0
			}

			relX := msg.X - startX
			relY := msg.Y - startY

			if relX >= 0 && relX < cardW {
				// 1. Tabs row (lines 6-7)
				if relY >= 6 && relY <= 7 {
					if relX < cardW/2 {
						m.mode = ModeLogin
					} else {
						m.mode = ModeRegister
					}
					m.focusIndex = 0
					m.updateFocus()
					return m, nil
				}

				errorOffset := 0
				if m.errorMessage != "" {
					errorOffset = 4
				}

				if m.mode == ModeLogin {
					// Username field (lines 9-12 + errorOffset)
					if relY >= 9+errorOffset && relY <= 12+errorOffset {
						m.focusIndex = 0
						m.updateFocus()
						return m, nil
					}
					// Password field (lines 14-17 + errorOffset)
					if relY >= 14+errorOffset && relY <= 17+errorOffset {
						m.focusIndex = 1
						m.updateFocus()
						return m, nil
					}
					// Submit button (lines 18-20 + errorOffset)
					if relY >= 18+errorOffset && relY <= 20+errorOffset {
						if !m.loading {
							return m, m.submit()
						}
						return m, nil
					}
				} else {
					// Register mode
					// Display Name field (lines 9-12 + errorOffset)
					if relY >= 9+errorOffset && relY <= 12+errorOffset {
						m.focusIndex = 0
						m.updateFocus()
						return m, nil
					}
					// Username field (lines 14-17 + errorOffset)
					if relY >= 14+errorOffset && relY <= 17+errorOffset {
						m.focusIndex = 1
						m.updateFocus()
						return m, nil
					}
					// Password field (lines 19-22 + errorOffset)
					if relY >= 19+errorOffset && relY <= 22+errorOffset {
						m.focusIndex = 2
						m.updateFocus()
						return m, nil
					}
					// Submit button (lines 25-28 + errorOffset)
					if relY >= 25+errorOffset && relY <= 28+errorOffset {
						if !m.loading {
							return m, m.submit()
						}
						return m, nil
					}
				}
			}
		}

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
		password := strings.TrimSpace(m.loginPassword.Value())

		if username == "" {
			m.errorMessage = "Please enter your username"
			return nil
		}
		if len(username) < 3 {
			m.errorMessage = "Username must be at least 3 characters"
			return nil
		}
		if len(username) > 20 {
			m.errorMessage = "Username must not exceed 20 characters"
			return nil
		}
		if password == "" {
			m.errorMessage = "Please enter your password"
			return nil
		}

		m.loading = true
		m.errorMessage = ""
		c := m.client

		return tea.Batch(
			m.spinner.Tick,
			func() tea.Msg {
				prof, err := c.Login(username, password)
				if err != nil {
					return AuthErrorMsg{Err: err}
				}
				return AuthSuccessMsg{Profile: *prof}
			},
		)
	}

	// Register mode
	name := strings.TrimSpace(m.regName.Value())
	username := strings.TrimSpace(m.regUsername.Value())
	password := strings.TrimSpace(m.regPassword.Value())

	if name == "" {
		m.errorMessage = "Please enter your display name"
		return nil
	}
	if len(name) < 3 {
		m.errorMessage = "Display name must be at least 3 characters"
		return nil
	}
	if len(name) > 50 {
		m.errorMessage = "Display name must not exceed 50 characters"
		return nil
	}
	if username == "" {
		m.errorMessage = "Please enter a username"
		return nil
	}
	if len(username) < 3 {
		m.errorMessage = "Username must be at least 3 characters"
		return nil
	}
	if len(username) > 20 {
		m.errorMessage = "Username must not exceed 20 characters"
		return nil
	}
	if password == "" {
		m.errorMessage = "Please enter a password"
		return nil
	}
	if len(password) < 6 {
		m.errorMessage = "Password must be at least 6 characters"
		return nil
	}
	if !strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		m.errorMessage = "Password must contain at least one capital letter (A-Z)"
		return nil
	}
	if !strings.ContainsAny(password, "0123456789") {
		m.errorMessage = "Password must contain at least one number (0-9)"
		return nil
	}

	m.loading = true
	m.errorMessage = ""
	c := m.client

	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			prof, err := c.Register(name, username, password)
			if err != nil {
				return AuthErrorMsg{Err: err}
			}
			return AuthSuccessMsg{Profile: *prof}
		},
	)
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
	header := theme.RenderLineTalkLogo(innerWidth)

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

		regPwd := m.regPassword.Value()
		hasUpper := strings.ContainsAny(regPwd, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
		hasNum := strings.ContainsAny(regPwd, "0123456789")
		hasMinLen := len(regPwd) >= 6

		renderHint := func(met bool, label string) string {
			if met {
				chk := lipgloss.NewStyle().Bold(true).Foreground(theme.ColorGreen).Render("[✔]")
				txt := lipgloss.NewStyle().Foreground(theme.ColorGreen).Render(" " + label)
				return chk + txt
			}
			chk := lipgloss.NewStyle().Foreground(theme.ColorDimText).Render("[ ]")
			txt := lipgloss.NewStyle().Foreground(theme.ColorDimText).Render(" " + label)
			return chk + txt
		}

		hintRow1 := lipgloss.JoinHorizontal(lipgloss.Left,
			renderHint(hasUpper, "Uppercase (A-Z)"),
			"   ",
			renderHint(hasNum, "Number (0-9)"),
		)
		hintRow2 := renderHint(hasMinLen, "Minimum 6 characters")
		checklist := lipgloss.JoinVertical(lipgloss.Left, hintRow1, hintRow2)

		formContent = lipgloss.JoinVertical(lipgloss.Left, nBox, "", uBox, "", pBox, "", checklist)
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
	var actionText string
	if m.loading {
		actionText = fmt.Sprintf("%s Authenticating with API server...", m.spinner.View())
	} else if m.mode == ModeRegister {
		actionText = "[ Enter: Create Account & Enter Chat ]"
	} else {
		actionText = "[ Enter: Login & Enter Chat ]"
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
		Background(theme.ColorDarkBg).
		Padding(1, 2).
		Width(cardWidth).
		Render(cardContent)

	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, card, lipgloss.WithWhitespaceBackground(theme.ColorDarkBg))
	}
	return card
}

// CleanAuthError translates backend validator and server errors into human-friendly, actionable messages.
func CleanAuthError(err error) string {
	if err == nil {
		return ""
	}
	return CleanAuthErrorString(err.Error())
}

// CleanAuthErrorString cleans raw error message strings.
func CleanAuthErrorString(msg string) string {
	trimmed := strings.TrimSpace(msg)
	if trimmed == "" {
		return ""
	}

	lower := strings.ToLower(trimmed)

	// Specific backend auth service errors
	if strings.Contains(lower, "incorrect username or password") {
		return "Incorrect username or password. Please try again."
	}
	if strings.Contains(lower, "user already exists") {
		return "Username is already taken. Please choose another username."
	}

	// Go validator/v10 tag errors
	if strings.Contains(trimmed, "containsany") {
		return "Password must contain at least one uppercase letter (A-Z) and one number (0-9)."
	}
	if strings.Contains(trimmed, "validation error") {
		if strings.Contains(trimmed, "Password") {
			return "Password must be at least 6 characters and contain uppercase letters and numbers."
		}
		if strings.Contains(trimmed, "Username") {
			return "Username must be between 3 and 20 characters."
		}
		if strings.Contains(trimmed, "Name") {
			return "Display name must be between 3 and 50 characters."
		}
		return "Please verify that all fields are filled out correctly."
	}
	if strings.Contains(trimmed, "failed on the 'min' tag") {
		if strings.Contains(trimmed, "Username") {
			return "Username must be at least 3 characters."
		}
		if strings.Contains(trimmed, "Name") {
			return "Display name must be at least 3 characters."
		}
		if strings.Contains(trimmed, "Password") {
			return "Password must be at least 6 characters."
		}
		return "One or more fields do not meet the minimum length requirement."
	}
	if strings.Contains(trimmed, "failed on the 'max' tag") {
		return "One or more fields exceed the maximum length."
	}
	if strings.Contains(trimmed, "failed on the 'required' tag") {
		return "Please fill in all required fields."
	}

	// Server / network issues
	if strings.Contains(lower, "offline") || strings.Contains(lower, "unreachable") || strings.Contains(lower, "connection refused") {
		return "Chat server is currently offline or unreachable."
	}
	if strings.Contains(lower, "timeout") || strings.Contains(lower, "timed out") {
		return "Request timed out. Please try again."
	}
	if strings.Contains(lower, "http 500") || strings.Contains(lower, "internal server error") {
		return "Server error occurred. Please try again later."
	}
	if strings.Contains(lower, "http 401") || strings.Contains(lower, "unauthorized") {
		return "Invalid credentials. Please verify your username and password."
	}

	// Check if this looks like an internal Go or SQL error
	if strings.Contains(trimmed, "Key:") || strings.Contains(trimmed, "Error:") || strings.Contains(lower, "sql") || strings.Contains(lower, "syntax") {
		return "Invalid input provided. Please verify your information and try again."
	}

	return trimmed
}
