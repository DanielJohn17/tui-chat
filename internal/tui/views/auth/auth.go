package auth

import (
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
		return m.handleMouse(msg)

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
	tabs := renderTabs(m.mode, innerWidth)

	// Form inputs
	var formContent string
	if m.mode == ModeLogin {
		uBox := renderField("Username", m.loginUsername.View(), m.focusIndex == 0, inputWidth)
		pBox := renderField("Password", m.loginPassword.View(), m.focusIndex == 1, inputWidth)
		formContent = lipgloss.JoinVertical(lipgloss.Left, uBox, "", pBox)
	} else {
		nBox := renderField("Display Name", m.regName.View(), m.focusIndex == 0, inputWidth)
		uBox := renderField("Username", m.regUsername.View(), m.focusIndex == 1, inputWidth)
		pBox := renderField("Password", m.regPassword.View(), m.focusIndex == 2, inputWidth)
		checklist := renderRegisterChecklist(m.regPassword.Value())
		formContent = lipgloss.JoinVertical(lipgloss.Left, nBox, "", uBox, "", pBox, "", checklist)
	}

	// Error banner if any
	errorBanner := renderErrorBanner(m.errorMessage, inputWidth)

	// Submit button
	submitActionCentered := renderSubmitButton(m.loading, m.mode, m.spinner.View(), innerWidth)

	// Footer helper with Help Desk and Quit shortcuts
	footer := renderFooter(innerWidth)

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
