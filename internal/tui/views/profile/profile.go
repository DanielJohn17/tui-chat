package profile

import (
	"strings"
	"time"

	"github.com/DanielJohn17/tui-chat/app/internal/tui/client"
	"github.com/DanielJohn17/tui-chat/app/internal/tui/theme"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type BackToChatMsg struct{}

type LogoutMsg struct{}

type ProfileSavedMsg struct {
	Profile client.Profile
}

type Model struct {
	client     client.Client
	width      int
	height     int
	focusIndex int

	nameInput     textinput.Model
	usernameInput textinput.Model
	statusInput   textinput.Model

	successMsg string
	errorMsg   string

	// Layout coordinates for mouse click detection
	lastBoxWidth int
	lastBoxX     int
	lastBoxY     int
}

func New(c client.Client) Model {
	prof := c.Profile()

	createInput := func(val, placeholder string, charLimit int) textinput.Model {
		ti := textinput.New()
		ti.SetValue(val)
		ti.Placeholder = placeholder
		ti.Prompt = ""
		ti.CharLimit = charLimit
		ti.TextStyle = lipgloss.NewStyle().Foreground(theme.ColorWhite)
		ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(theme.ColorDimText)
		ti.Cursor.Style = lipgloss.NewStyle().Foreground(theme.ColorCyan)
		return ti
	}

	ni := createInput(prof.Name, "Enter display name", 40)
	ni.Focus()

	ui := createInput(prof.Username, "Enter username", 25)

	bioVal := "Hacking in the terminal 🚀"
	si := createInput(bioVal, "Set custom status / bio", 60)

	return Model{
		client:        c,
		focusIndex:    0,
		nameInput:     ni,
		usernameInput: ui,
		statusInput:   si,
	}
}

func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m *Model) ReloadProfile() {
	prof := m.client.Profile()
	m.nameInput.SetValue(prof.Name)
	m.usernameInput.SetValue(prof.Username)
	m.successMsg = ""
	m.errorMsg = ""
	m.focusIndex = 0
	m.updateFocus()
}

func (m *Model) updateFocus() {
	m.nameInput.Blur()
	m.usernameInput.Blur()
	m.statusInput.Blur()

	switch m.focusIndex {
	case 0:
		m.nameInput.Focus()
	case 1:
		m.usernameInput.Focus()
	case 2:
		m.statusInput.Focus()
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			// Check if clicked inside form fields or buttons
			// The card is centered vertically and horizontally
			cardWidth := 62
			if m.width > 0 && m.width < 66 {
				cardWidth = m.width - 4
			}
			if cardWidth < 40 {
				cardWidth = 40
			}
			cardW := cardWidth + 6
			cardH := 32
			if m.errorMsg != "" || m.successMsg != "" {
				cardH = 34
			}
			startX := (m.width - cardW) / 2
			startY := (m.height - cardH) / 2
			if startX < 0 {
				startX = 0
			}
			if startY < 0 {
				startY = 0
			}

			relY := msg.Y - startY
			relX := msg.X - startX

			if relX >= 0 && relX < cardW {
				alertOffset := 0
				if m.errorMsg != "" || m.successMsg != "" {
					alertOffset = 2
				}

				// Precise row positions inside card:
				// Name field (label at 7, box at 8-10)
				if relY >= 7+alertOffset && relY <= 10+alertOffset {
					m.focusIndex = 0
					m.updateFocus()
					return m, nil
				}
				// Username field (label at 12, box at 13-15)
				if relY >= 12+alertOffset && relY <= 15+alertOffset {
					m.focusIndex = 1
					m.updateFocus()
					return m, nil
				}
				// Status / Bio field (label at 17, box at 18-20)
				if relY >= 17+alertOffset && relY <= 20+alertOffset {
					m.focusIndex = 2
					m.updateFocus()
					return m, nil
				}
				// Action buttons: Save Changes, Back to Chat, Logout (lines 25-28)
				if relY >= 25+alertOffset && relY <= 28+alertOffset {
					if relX < cardW/3 {
						m.focusIndex = 3
						return m.save()
					} else if relX < (cardW*3)/4 {
						m.focusIndex = 4
						return m, func() tea.Msg { return BackToChatMsg{} }
					} else {
						m.focusIndex = 5
						return m, func() tea.Msg { return LogoutMsg{} }
					}
				}
			}
		}

	case tea.KeyMsg:
		m.errorMsg = ""

		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return BackToChatMsg{} }

		case "l", "ctrl+l":
			return m, func() tea.Msg { return LogoutMsg{} }

		case "tab", "down":
			m.focusIndex = (m.focusIndex + 1) % 6
			m.updateFocus()
			return m, nil

		case "shift+tab", "up":
			m.focusIndex = (m.focusIndex - 1 + 6) % 6
			m.updateFocus()
			return m, nil

		case "enter":
			if m.focusIndex == 4 { // Cancel button
				return m, func() tea.Msg { return BackToChatMsg{} }
			}
			if m.focusIndex == 5 { // Logout button
				return m, func() tea.Msg { return LogoutMsg{} }
			}
			// From any field or Save button, submit
			return m.save()
		}
	}

	var cmd tea.Cmd
	switch m.focusIndex {
	case 0:
		m.nameInput, cmd = m.nameInput.Update(msg)
	case 1:
		m.usernameInput, cmd = m.usernameInput.Update(msg)
	case 2:
		m.statusInput, cmd = m.statusInput.Update(msg)
	}

	return m, cmd
}

func (m Model) save() (Model, tea.Cmd) {
	name := strings.TrimSpace(m.nameInput.Value())
	username := strings.TrimSpace(m.usernameInput.Value())

	if name == "" {
		m.errorMsg = "Display Name cannot be empty"
		m.focusIndex = 0
		m.updateFocus()
		return m, nil
	}
	if len(username) < 2 {
		m.errorMsg = "Username must be at least 2 characters"
		m.focusIndex = 1
		m.updateFocus()
		return m, nil
	}

	currentProf := m.client.Profile()
	currentProf.Name = name
	currentProf.Username = username
	m.client.UpdateProfile(currentProf)

	m.successMsg = "✔ Profile changes saved successfully!"
	m.errorMsg = ""

	return m, func() tea.Msg {
		return ProfileSavedMsg{Profile: currentProf}
	}
}

func (m Model) View() string {
	cardWidth := 62
	if m.width > 0 && m.width < 66 {
		cardWidth = m.width - 4
	}
	if cardWidth < 40 {
		cardWidth = 40
	}

	innerWidth := cardWidth - 6
	inputWidth := innerWidth - 4

	m.nameInput.Width = inputWidth - 2
	m.usernameInput.Width = inputWidth - 2
	m.statusInput.Width = inputWidth - 2

	// 1. Page Header & Subtitle
	badge := theme.StyleBadge.Render("PAGE: PROFILE EDIT")
	title := theme.StyleTitle.Render("◈ EDIT USER PROFILE & IDENTITY ◈")
	subtitle := theme.StyleDim.Render("Update your terminal account identity and broadcast status")

	header := lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(badge),
		"",
		lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(title),
		lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(subtitle),
	)

	// 2. Field Renderer
	renderField := func(label string, inputView string, isFocused bool) string {
		lblStyle := theme.StyleDim
		borderCol := theme.ColorBorderDim
		if isFocused {
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

	nameBox := renderField("Display Name", m.nameInput.View(), m.focusIndex == 0)
	userBox := renderField("Username", m.usernameInput.View(), m.focusIndex == 1)
	statusBox := renderField("Status / Bio", m.statusInput.View(), m.focusIndex == 2)

	// 3. Read-Only Account Details
	prof := m.client.Profile()
	renderDetail := func(lbl, val string) string {
		left := theme.StyleDim.Render(lbl)
		right := lipgloss.NewStyle().Foreground(theme.ColorCyan).Render(val)
		spaces := innerWidth - 4 - lipgloss.Width(left) - lipgloss.Width(right)
		if spaces < 1 {
			spaces = 1
		}
		return lipgloss.JoinHorizontal(lipgloss.Center, left, lipgloss.NewStyle().Width(spaces).Render(""), right)
	}

	created := formatFriendlyDate(prof.CreatedAt)

	infoBox := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(theme.ColorBorderDim).
		Padding(0, 1).
		Render(lipgloss.JoinVertical(lipgloss.Left,
			renderDetail("Member Since:", created),
			renderDetail("Status:", "● Connected & Authenticated"),
		))

	// 4. Buttons
	saveStyle := lipgloss.NewStyle().
		Bold(true).
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1)
	if m.focusIndex == 3 {
		saveStyle = saveStyle.
			Foreground(theme.ColorBlack).
			Background(theme.ColorGreen).
			BorderForeground(theme.ColorGreen)
	} else {
		saveStyle = saveStyle.
			Foreground(theme.ColorGreen).
			BorderForeground(theme.ColorGreen)
	}

	cancelStyle := lipgloss.NewStyle().
		Bold(true).
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1)
	if m.focusIndex == 4 {
		cancelStyle = cancelStyle.
			Foreground(theme.ColorBlack).
			Background(theme.ColorCyan).
			BorderForeground(theme.ColorCyan)
	} else {
		cancelStyle = cancelStyle.
			Foreground(theme.ColorDimText).
			BorderForeground(theme.ColorBorderDim)
	}

	logoutStyle := lipgloss.NewStyle().
		Bold(true).
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1)
	if m.focusIndex == 5 {
		logoutStyle = logoutStyle.
			Foreground(theme.ColorWhite).
			Background(theme.ColorRed).
			BorderForeground(theme.ColorRed)
	} else {
		logoutStyle = logoutStyle.
			Foreground(theme.ColorRed).
			BorderForeground(theme.ColorBorderDim)
	}

	saveBtn := saveStyle.Render("✔ Save")
	cancelBtn := cancelStyle.Render("✖ Back [Esc]")
	logoutBtn := logoutStyle.Render("⏻ Logout [L]")

	buttonsRow := lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(
		lipgloss.JoinHorizontal(lipgloss.Center, saveBtn, "  ", cancelBtn, "  ", logoutBtn),
	)

	// Feedback Messages
	var alertBox string
	if m.errorMsg != "" {
		alertBox = lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.ColorRed).
			Width(innerWidth).
			Align(lipgloss.Center).
			Render("✖ " + m.errorMsg)
	} else if m.successMsg != "" {
		alertBox = lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.ColorGreen).
			Width(innerWidth).
			Align(lipgloss.Center).
			Render("✔ " + m.successMsg)
	}

	// Hotkey hints footer
	hint := theme.StyleDim.Render("Tab / ↑↓: Navigate  •  Enter: Select  •  L: Logout  •  Esc: Back")
	footer := lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(hint)

	items := []string{
		header,
		"",
	}
	if alertBox != "" {
		items = append(items, alertBox, "")
	}
	items = append(items,
		nameBox,
		"",
		userBox,
		"",
		statusBox,
		"",
		infoBox,
		"",
		buttonsRow,
		"",
		footer,
	)

	cardContent := lipgloss.JoinVertical(lipgloss.Left, items...)

	card := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorMagenta).
		Padding(1, 2).
		Width(cardWidth).
		Render(cardContent)

	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, card)
	}
	return card
}

func formatFriendlyDate(tStr string) string {
	if tStr == "" {
		return "Active Session"
	}
	layouts := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05.999999999Z07:00",
		"2006-01-02 15:04:05.999999999-07:00",
		"2006-01-02 15:04:05-07:00",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, l := range layouts {
		if t, err := time.Parse(l, tStr); err == nil {
			return t.Local().Format("Jan 02, 2006 15:04")
		}
	}
	return tStr
}
