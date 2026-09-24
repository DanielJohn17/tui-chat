package modals

import (
	"strings"

	"github.com/DanielJohn17/tui-chat/internal/tui/theme"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type NewDMModel struct {
	input    textinput.Model
	width    int
	height   int
	errorMsg string
}

func NewDM() NewDMModel {
	ti := textinput.New()
	ti.Placeholder = "e.g. alice"
	ti.CharLimit = 20
	ti.Focus()
	ti.Prompt = "@"
	ti.PromptStyle = lipgloss.NewStyle().Foreground(theme.ColorMagenta).Bold(true)
	ti.TextStyle = lipgloss.NewStyle().Foreground(theme.ColorWhite)

	return NewDMModel{
		input: ti,
	}
}

func (m *NewDMModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m *NewDMModel) Reset() {
	m.input.Reset()
	m.input.Focus()
	m.errorMsg = ""
}

func (m NewDMModel) Value() string {
	return strings.TrimSpace(m.input.Value())
}

func (m *NewDMModel) SetError(err string) {
	m.errorMsg = err
}

func (m NewDMModel) Update(msg tea.Msg) (NewDMModel, tea.Cmd) {
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m NewDMModel) View(width, height int) string {
	boxWidth := 48
	if width > 0 && width < 52 {
		boxWidth = width - 4
	}
	if boxWidth < 36 {
		boxWidth = 36
	}

	// Inner content width inside padding(1, 2) and border(1, 1) is boxWidth - 6
	innerWidth := boxWidth - 6

	title := theme.StyleTitle.Render("◈ NEW DIRECT MESSAGE ◈")
	desc := theme.StyleDim.Render("Enter username to initiate conversation:")

	// Input box inside rounded border and padding(0, 1)
	inputBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorCyan).
		Width(innerWidth-4).
		Padding(0, 1).
		Render(m.input.View())

	var errBanner string
	if m.errorMsg != "" {
		errBanner = theme.StyleError.Render("✖ " + m.errorMsg)
	}

	actions := lipgloss.JoinHorizontal(lipgloss.Center,
		theme.StyleSuccess.Render("[Enter: Connect]"),
		"   ",
		theme.StyleDim.Render("[Esc: Cancel]"),
	)

	items := []string{
		lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(title),
		"",
		desc,
		inputBox,
	}
	if errBanner != "" {
		items = append(items, errBanner)
	}
	items = append(items, "", lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(actions))

	content := lipgloss.JoinVertical(lipgloss.Left, items...)

	modal := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorCyan).
		Padding(1, 2).
		Width(boxWidth).
		Render(content)

	if width > 0 && height > 0 {
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, modal)
	}
	return modal
}
