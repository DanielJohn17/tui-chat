package modals

import (
	"fmt"
	"strings"

	"github.com/DanielJohn17/tui-chat/internal/tui/client"
	"github.com/DanielJohn17/tui-chat/internal/tui/theme"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type NewDMModel struct {
	input    textinput.Model
	results  []client.UserSummary
	selected int
	loading  bool
	opening  bool
	searched bool
	errorMsg string
}

func NewDM() NewDMModel {
	ti := textinput.New()
	ti.Placeholder = "Search by @username or paste invite link..."
	ti.CharLimit = 64
	ti.Prompt = "@"
	ti.PromptStyle = lipgloss.NewStyle().Foreground(theme.ColorMagenta).Bold(true)
	ti.TextStyle = lipgloss.NewStyle().Foreground(theme.ColorWhite)
	return NewDMModel{input: ti}
}

func (m *NewDMModel) SetSize(w, h int) {
	m.input.Width = max(1, min(70, w-10))
}

func (m *NewDMModel) Reset() {
	m.input.Reset()
	m.input.Focus()
	m.SetLoading(false)
}

func (m NewDMModel) Value() string { return strings.TrimSpace(m.input.Value()) }
func (m NewDMModel) Opening() bool { return m.opening }

func (m *NewDMModel) SetLoading(loading bool) {
	m.results = nil
	m.selected = 0
	m.loading = loading
	m.opening = false
	m.searched = false
	m.errorMsg = ""
}

func (m *NewDMModel) SetResults(users []client.UserSummary, err error) {
	m.loading = false
	m.searched = true
	m.results = users
	m.selected = 0
	if err != nil {
		m.results = nil
		m.SetError(err.Error())
	}
}

func (m *NewDMModel) SetError(err string) {
	m.errorMsg = err
	m.loading = false
	m.opening = false
}

func (m *NewDMModel) StartOpening() {
	m.opening = true
	m.errorMsg = ""
}

func (m NewDMModel) Selected() (client.UserSummary, bool) {
	if m.loading || m.selected < 0 || m.selected >= len(m.results) {
		return client.UserSummary{}, false
	}
	return m.results[m.selected], true
}

func (m NewDMModel) Update(msg tea.Msg) (NewDMModel, tea.Cmd) {
	if m.opening {
		return m, nil
	}
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "up":
			m.selected = max(0, m.selected-1)
			return m, nil
		case "down":
			m.selected = max(0, min(len(m.results)-1, m.selected+1))
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m NewDMModel) View(width, height int) string {
	if width < 1 || height < 1 {
		return ""
	}
	border := width >= 6 && height >= 7
	targetW := 74
	if width < targetW {
		targetW = width
	}
	if targetW < 1 {
		targetW = 1
	}

	innerWidth := targetW
	availableHeight := height
	if border {
		innerWidth -= 4
		availableHeight -= 2
	}
	if innerWidth < 1 {
		innerWidth = 1
	}

	// Status text logic
	status := "Type at least 2 characters"
	switch {
	case m.opening:
		status = "Opening conversation..."
	case m.errorMsg != "":
		status = "Error: " + m.errorMsg
	case m.loading:
		status = "Searching..."
	case m.searched && len(m.results) == 0:
		status = "No results found matching query"
	case len(m.results) > 0:
		status = fmt.Sprintf("Select a user (%d found)", len(m.results))
	}

	statusStyle := theme.StyleDim
	if m.errorMsg != "" {
		statusStyle = theme.StyleError
	} else if m.loading {
		statusStyle = lipgloss.NewStyle().Foreground(theme.ColorYellow)
	} else if len(m.results) > 0 {
		statusStyle = theme.StyleSubtitle
	}

	var lines []string

	// Compact mode for small terminals
	if availableHeight < 12 {
		m.input.Width = max(1, innerWidth-2)
		input := m.input.View()
		if lipgloss.Width(input) > innerWidth {
			input = theme.Truncate(theme.StripANSI(input), innerWidth)
		}
		lines = []string{
			theme.StyleTitle.Render(theme.Truncate("NEW DIRECT MESSAGE", innerWidth)),
			input,
			statusStyle.Render(theme.Truncate(theme.SanitizeLine(status), innerWidth)),
		}

		visible := max(0, availableHeight-4)
		start := max(0, m.selected-visible+1)
		for i := start; i < len(m.results) && i < start+visible; i++ {
			user := m.results[i]
			prefix := "  "
			style := lipgloss.NewStyle().Foreground(theme.ColorCyan)
			if i == m.selected {
				prefix = "> "
				style = style.Foreground(theme.ColorMagenta).Bold(true)
			}
			item := prefix + theme.SanitizeLine(user.Name) + " @" + theme.SanitizeLine(user.Username)
			if user.InviteToken != "" {
				item += " [Invite]"
			}
			if user.Bio != "" {
				item += " • " + theme.SanitizeLine(user.Bio)
			}
			lines = append(lines, style.Render(theme.Truncate(item, innerWidth)))
		}
		lines = append(lines, theme.StyleDim.Render(theme.Truncate("Up/Down: Select  Enter: Open  Esc: Cancel", innerWidth)))
	} else {
		// Rich inviting mode for standard terminals
		titleHeader := lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(
			theme.StyleTitle.Render("◈ NEW DIRECT MESSAGE ◈"),
		)
		lines = append(lines, titleHeader, "")

		// Bordered search input container (Width innerWidth-2 so outer width with borders equals innerWidth)
		m.input.Width = max(1, innerWidth-6)
		searchBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.ColorCyan).
			Width(innerWidth-2).
			Padding(0, 1).
			Render(m.input.View())
		lines = append(lines, strings.Split(searchBox, "\n")...)
		lines = append(lines, "")

		// Status banner
		lines = append(lines, statusStyle.Render(theme.Truncate(status, innerWidth)))

		// Friendly tips when search has not yet yielded results
		if !m.searched && !m.loading && len(m.results) == 0 && availableHeight >= 16 {
			lines = append(lines,
				"",
				theme.StyleDim.Render("💡 How to connect:"),
				theme.StyleDim.Render("  • Type @username to search public users in directory"),
				theme.StyleDim.Render("  • Paste an invite link (tui://invite/...) to connect with private users"),
			)
		}

		// Results list
		if len(m.results) > 0 {
			lines = append(lines, "")
			overhead := len(lines) + 2 // reserve space for footer & separator
			visible := max(1, availableHeight-overhead)
			start := max(0, m.selected-visible+1)

			for i := start; i < len(m.results) && i < start+visible; i++ {
				user := m.results[i]
				var prefix string
				nameStyle := lipgloss.NewStyle().Foreground(theme.ColorWhite)
				handleStyle := lipgloss.NewStyle().Foreground(theme.ColorCyan)

				if i == m.selected {
					prefix = lipgloss.NewStyle().Bold(true).Foreground(theme.ColorMagenta).Render("▶ ")
					nameStyle = nameStyle.Bold(true)
					handleStyle = handleStyle.Bold(true)
				} else {
					prefix = "  "
				}

				nameView := nameStyle.Render(theme.SanitizeLine(user.Name))
				handleView := handleStyle.Render("@" + theme.SanitizeLine(user.Username))

				var badgeView string
				if user.InviteToken != "" {
					badgeView = " " + lipgloss.NewStyle().Bold(true).Foreground(theme.ColorBlack).Background(theme.ColorGreen).Render(" INVITE ")
				}

				var bioView string
				if user.Bio != "" {
					bioView = " " + theme.StyleDim.Render("• "+theme.SanitizeLine(user.Bio))
				}

				itemLine := prefix + nameView + " " + handleView + badgeView + bioView
				lines = append(lines, theme.Truncate(itemLine, innerWidth))
			}
		}

		lines = append(lines,
			"",
			theme.StyleDim.Render(theme.Truncate("↑↓: Navigate  •  Enter: Open Chat  •  Esc: Cancel", innerWidth)),
		)
	}

	if len(lines) > availableHeight {
		lines = lines[:availableHeight]
	}

	boxStyle := lipgloss.NewStyle().Width(innerWidth)
	if border {
		boxStyle = boxStyle.
			Width(innerWidth+2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.ColorCyan).
			Padding(0, 1)
	}
	box := boxStyle.Render(strings.Join(lines, "\n"))
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}
