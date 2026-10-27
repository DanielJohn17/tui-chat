package profile

import (
	"strings"
	"time"

	"github.com/DanielJohn17/tui-chat/internal/tui/client"
	"github.com/DanielJohn17/tui-chat/internal/tui/theme"
	"github.com/atotto/clipboard"
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
	isSearchable  bool

	successMsg string
	errorMsg   string

	// Layout coordinates for mouse click detection
	lastBoxWidth int
	lastBoxX     int
	lastBoxY     int
}

func New(c client.Client) Model {
	prof := c.Profile()
	if prof.InviteToken == "" && c.IsAuthenticated() {
		if p, err := c.FetchProfile(); err == nil && p != nil {
			prof = *p
		}
	}

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

	bioVal := prof.Bio
	if bioVal == "" {
		bioVal = "Hacking in the terminal 🚀"
	}
	si := createInput(bioVal, "Set custom status / bio", 60)

	return Model{
		client:        c,
		focusIndex:    0,
		nameInput:     ni,
		usernameInput: ui,
		statusInput:   si,
		isSearchable:  prof.IsSearchable,
	}
}

func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m *Model) ReloadProfile() {
	prof := m.client.Profile()
	if prof.InviteToken == "" && m.client.IsAuthenticated() {
		if p, err := m.client.FetchProfile(); err == nil && p != nil {
			prof = *p
		}
	}
	m.nameInput.SetValue(prof.Name)
	m.usernameInput.SetValue(prof.Username)
	if prof.Bio != "" {
		m.statusInput.SetValue(prof.Bio)
	} else {
		m.statusInput.SetValue("Hacking in the terminal 🚀")
	}
	m.isSearchable = prof.IsSearchable
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
			cardWidth := 80
			if m.width > 0 && m.width < 84 {
				cardWidth = m.width - 4
			}
			if cardWidth < 40 {
				cardWidth = 40
			}
			cardW := cardWidth + 6
			cardH := 33
			if m.errorMsg != "" || m.successMsg != "" {
				cardH = 35
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
				// Privacy mode toggle (line 22)
				if relY == 22+alertOffset {
					m.focusIndex = 3
					m.updateFocus()
					m.isSearchable = !m.isSearchable
					return m, nil
				}
				// Invite link: Copy (left/middle) & Regenerate (right) (line 23)
				if relY == 23+alertOffset {
					if relX > cardW-18 {
						m.focusIndex = 5
						m.updateFocus()
						return m.regenerateToken()
					}
					m.focusIndex = 4
					m.updateFocus()
					return m.copyLink()
				}
				// Action buttons: Save Changes, Back to Chat, Logout (lines 26-29)
				if relY >= 26+alertOffset && relY <= 29+alertOffset {
					if relX < cardW/3 {
						m.focusIndex = 6
						return m.save()
					} else if relX < (cardW*3)/4 {
						m.focusIndex = 7
						return m, func() tea.Msg { return BackToChatMsg{} }
					} else {
						m.focusIndex = 8
						return m, func() tea.Msg { return LogoutMsg{} }
					}
				}
			}
		}

	case tea.KeyMsg:
		m.errorMsg = ""

		switch msg.String() {
		case "L", "ctrl+l":
			return m, func() tea.Msg { return LogoutMsg{} }

		case "l":
			if m.focusIndex >= 3 {
				return m, func() tea.Msg { return LogoutMsg{} }
			}

		case "c", "C":
			if m.focusIndex >= 3 {
				return m.copyLink()
			}

		case "r", "R":
			if m.focusIndex >= 3 {
				return m.regenerateToken()
			}

		case "tab", "down":
			m.focusIndex = (m.focusIndex + 1) % 9
			m.updateFocus()
			return m, nil

		case "shift+tab", "up":
			m.focusIndex = (m.focusIndex - 1 + 9) % 9
			m.updateFocus()
			return m, nil

		case " ":
			if m.focusIndex == 3 {
				m.isSearchable = !m.isSearchable
				return m, nil
			}
			if m.focusIndex == 4 {
				return m.copyLink()
			}
			if m.focusIndex == 5 {
				return m.regenerateToken()
			}

		case "ctrl+s":
			return m.save()

		case "enter":
			if m.focusIndex == 3 { // Privacy toggle
				m.isSearchable = !m.isSearchable
				return m, nil
			}
			if m.focusIndex == 4 { // Copy button
				return m.copyLink()
			}
			if m.focusIndex == 5 { // Regenerate button
				return m.regenerateToken()
			}
			if m.focusIndex == 7 { // Cancel button
				return m, func() tea.Msg { return BackToChatMsg{} }
			}
			if m.focusIndex == 8 { // Logout button
				return m, func() tea.Msg { return LogoutMsg{} }
			}
			// From input fields or Save button (6), submit
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

func (m Model) copyLink() (Model, tea.Cmd) {
	prof := m.client.Profile()
	tokenStr := prof.InviteToken
	if tokenStr == "" {
		m.errorMsg = "No invite token available to copy"
		return m, nil
	}
	inviteURL := "tui://invite/" + tokenStr
	_ = clipboard.WriteAll(inviteURL)
	m.successMsg = "✔ Link copied to clipboard: " + inviteURL
	m.errorMsg = ""
	return m, nil
}

func (m Model) regenerateToken() (Model, tea.Cmd) {
	prof, err := m.client.RegenerateInviteToken()
	if err != nil {
		m.errorMsg = "Regen failed: " + err.Error()
		return m, nil
	}
	if prof != nil {
		m.client.UpdateProfile(*prof)
	}
	m.successMsg = "✔ New invite link generated"
	m.errorMsg = ""
	return m, nil
}

func (m Model) save() (Model, tea.Cmd) {
	name := strings.TrimSpace(m.nameInput.Value())
	username := strings.TrimSpace(m.usernameInput.Value())
	bio := strings.TrimSpace(m.statusInput.Value())
	if bio == "" {
		bio = "Hacking in the terminal 🚀"
	}

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
	currentProf.Bio = bio
	currentProf.IsSearchable = m.isSearchable

	updatedProf, err := m.client.SaveProfile(name, bio, m.isSearchable)
	if err != nil {
		m.errorMsg = "Failed to save: " + err.Error()
		return m, nil
	}
	if updatedProf != nil {
		currentProf = *updatedProf
	} else {
		m.client.UpdateProfile(currentProf)
	}

	m.successMsg = "✔ Profile saved to server"
	m.errorMsg = ""

	return m, func() tea.Msg {
		return ProfileSavedMsg{Profile: currentProf}
	}
}

func (m Model) View() string {
	cardWidth := 80
	if m.width > 0 && m.width < 84 {
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
	subtitle := theme.StyleDim.Render("Update your profile display, bio, and privacy settings")

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

	// Section Dividers
	isDiscoveryFocused := m.focusIndex == 3 || m.focusIndex == 4 || m.focusIndex == 5
	divColor := theme.ColorBorderDim
	if isDiscoveryFocused {
		divColor = theme.ColorCyan
	}

	renderDivider := func(title string) string {
		if title == "" {
			return lipgloss.NewStyle().Foreground(divColor).Render(strings.Repeat("─", innerWidth))
		}
		prefix := "─── " + title + " "
		pW := lipgloss.Width(prefix)
		rem := innerWidth - pW
		if rem < 2 {
			rem = 2
		}
		return lipgloss.NewStyle().Foreground(divColor).Render(prefix + strings.Repeat("─", rem))
	}

	upperDivider := renderDivider("Discovery & Direct Messaging")

	// 3. Privacy Mode Toggle (Row 22)
	var privacyModeStr string
	if m.isSearchable {
		privacyModeStr = theme.StyleSuccess.Render("[●] Public (Searchable)") + "   " + theme.StyleDim.Render("[○] Invite-Only")
	} else {
		privacyModeStr = theme.StyleDim.Render("[○] Public (Searchable)") + "   " + lipgloss.NewStyle().Bold(true).Foreground(theme.ColorYellow).Render("[●] Invite-Only")
	}
	privacyPrefix := "  Privacy: "
	if m.focusIndex == 3 {
		privacyPrefix = lipgloss.NewStyle().Bold(true).Foreground(theme.ColorCyan).Render("▸ Privacy: ")
	} else {
		privacyPrefix = theme.StyleDim.Render("  Privacy: ")
	}
	privacyRow := lipgloss.JoinHorizontal(lipgloss.Center,
		privacyPrefix, privacyModeStr,
	)

	// 4. Personal Invite Link, Copy, & Regen buttons (Row 23)
	prof := m.client.Profile()
	tokenStr := prof.InviteToken
	if tokenStr == "" {
		tokenStr = "none"
	}
	inviteURL := "tui://invite/" + tokenStr

	copyBtnStyle := lipgloss.NewStyle().Foreground(theme.ColorGreen)
	if m.focusIndex == 4 {
		copyBtnStyle = lipgloss.NewStyle().Bold(true).Foreground(theme.ColorBlack).Background(theme.ColorGreen)
	}
	copyBtn := copyBtnStyle.Render("[Copy]")

	regenBtnStyle := lipgloss.NewStyle().Foreground(theme.ColorCyan)
	if m.focusIndex == 5 {
		regenBtnStyle = lipgloss.NewStyle().Bold(true).Foreground(theme.ColorBlack).Background(theme.ColorCyan)
	}
	regenBtn := regenBtnStyle.Render("[↻ Regen]")

	btnGroup := lipgloss.JoinHorizontal(lipgloss.Center, copyBtn, " ", regenBtn)

	invitePrefix := "  Invite:  "
	if m.focusIndex == 4 || m.focusIndex == 5 {
		invitePrefix = lipgloss.NewStyle().Bold(true).Foreground(theme.ColorCyan).Render("▸ Invite:  ")
	} else {
		invitePrefix = theme.StyleDim.Render("  Invite:  ")
	}

	maxLinkW := innerWidth - lipgloss.Width(invitePrefix) - lipgloss.Width(btnGroup) - 2
	if maxLinkW < 12 {
		maxLinkW = 12
	}
	truncatedLink := theme.Truncate(inviteURL, maxLinkW)
	inviteLinkView := lipgloss.NewStyle().Foreground(theme.ColorMagenta).Bold(true).Render(truncatedLink)

	// Right-align buttons cleanly
	linkSpacing := innerWidth - lipgloss.Width(invitePrefix) - lipgloss.Width(inviteLinkView) - lipgloss.Width(btnGroup)
	if linkSpacing < 1 {
		linkSpacing = 1
	}
	spacer := lipgloss.NewStyle().Width(linkSpacing).Render("")

	inviteRow := lipgloss.JoinHorizontal(lipgloss.Center,
		invitePrefix, inviteLinkView, spacer, btnGroup,
	)

	// 5. Account Details & Metadata Bar (Rows 24 & 25)
	lowerDivider := renderDivider("Account Status (Sync: Server)")

	created := formatFriendlyDate(prof.CreatedAt)
	sinceItem := theme.StyleDim.Render("Member Since: ") + lipgloss.NewStyle().Foreground(theme.ColorWhite).Render(created)
	statusItem := theme.StyleDim.Render("Status: ") + theme.StyleSuccess.Render("● Connected")
	bullet := theme.StyleDim.Render("   •   ")

	metaLine := lipgloss.JoinHorizontal(lipgloss.Center, sinceItem, bullet, statusItem)
	if lipgloss.Width(metaLine) > innerWidth {
		metaLine = theme.Truncate(metaLine, innerWidth)
	}
	metaBar := lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(metaLine)

	// 6. Action Buttons (Rows 26-28)
	saveStyle := lipgloss.NewStyle().
		Bold(true).
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1)
	if m.focusIndex == 6 {
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
	if m.focusIndex == 7 {
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
	if m.focusIndex == 8 {
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
	hint := theme.StyleDim.Render("Tab/↑↓: Navigate • Space: Toggle • C: Copy • R: Regen • Enter: Select • Esc: Back")
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
		upperDivider,
		privacyRow,
		inviteRow,
		lowerDivider,
		metaBar,
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
