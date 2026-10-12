package modals

import (
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
	ti.Placeholder = "Search username"
	ti.CharLimit = 20
	ti.Prompt = "@"
	ti.PromptStyle = lipgloss.NewStyle().Foreground(theme.ColorMagenta).Bold(true)
	ti.TextStyle = lipgloss.NewStyle().Foreground(theme.ColorWhite)
	return NewDMModel{input: ti}
}

func (m *NewDMModel) SetSize(w, h int) {
	m.input.Width = max(1, min(56, w)-5)
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
	// Width includes horizontal padding; borders add two more cells.
	border := width >= 6 && height >= 7
	innerWidth := min(56, width)
	availableHeight := height
	if border {
		innerWidth -= 4
		availableHeight -= 2
	}
	m.input.Width = max(1, innerWidth-1)
	input := m.input.View()
	if lipgloss.Width(input) > innerWidth {
		input = theme.Truncate(theme.StripANSI(input), innerWidth)
	}
	lines := []string{
		theme.StyleTitle.Render(theme.Truncate("NEW DIRECT MESSAGE", innerWidth)),
		input,
	}
	status := "Type at least 2 characters"
	switch {
	case m.opening:
		status = "Opening conversation..."
	case m.errorMsg != "":
		status = "Error: " + m.errorMsg
	case m.loading:
		status = "Searching..."
	case m.searched && len(m.results) == 0:
		status = "No results"
	case len(m.results) > 0:
		status = "Select a user"
	}
	style := theme.StyleDim
	if m.errorMsg != "" {
		style = theme.StyleError
	}
	lines = append(lines, style.Render(theme.Truncate(theme.SanitizeLine(status), innerWidth)))
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
		lines = append(lines, style.Render(theme.Truncate(prefix+theme.SanitizeLine(user.Name)+" @"+theme.SanitizeLine(user.Username), innerWidth)))
	}
	lines = append(lines, theme.StyleDim.Render(theme.Truncate("Up/Down: Select  Enter: Open  Esc: Cancel", innerWidth)))
	if len(lines) > availableHeight {
		lines = lines[:availableHeight]
	}
	boxStyle := lipgloss.NewStyle().Width(innerWidth)
	if border {
		boxStyle = boxStyle.Width(innerWidth+2).Border(lipgloss.RoundedBorder()).BorderForeground(theme.ColorCyan).Padding(0, 1)
	}
	box := boxStyle.Render(strings.Join(lines, "\n"))
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}
