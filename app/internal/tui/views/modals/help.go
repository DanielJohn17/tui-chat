package modals

import (
	"github.com/DanielJohn17/tui-chat/app/internal/tui/theme"
	"github.com/charmbracelet/lipgloss"
)

func RenderHelp(width, height int) string {
	boxWidth := 52
	if width > 0 && width < 56 {
		boxWidth = width - 4
	}

	title := theme.StyleTitle.Render("◈ KEYBOARD COMMANDS ◈")
	closeHint := theme.StyleDim.Render("[ Esc / Enter / h: Close ]")

	renderRow := func(key, desc string) string {
		k := theme.StyleShortcut.Render(key)
		d := theme.StyleDim.Render(desc)
		spaces := boxWidth - 8 - lipgloss.Width(k) - lipgloss.Width(d)
		if spaces < 1 {
			spaces = 1
		}
		return lipgloss.JoinHorizontal(lipgloss.Center, k, lipgloss.NewStyle().Width(spaces).Render(""), d)
	}

	secNav := theme.StyleSubtitle.Render("NAVIGATION")
	r1 := renderRow("j / ↓", "Move down conversation list")
	r2 := renderRow("k / ↑", "Move up conversation list")
	r3 := renderRow("i / Enter", "Focus message input box")
	r4 := renderRow("Esc", "Unfocus input box / Close modal")

	secChat := theme.StyleSubtitle.Render("CHATTING")
	r5 := renderRow("Enter", "Send active message")
	r6 := renderRow("PgUp / PgDn", "Scroll message history")

	secActions := theme.StyleSubtitle.Render("ACTIONS & MODALS")
	r7 := renderRow("n", "Start new Direct Message")
	r8 := renderRow("p", "View user profile card")
	r9 := renderRow("h / ?", "Toggle this help modal")
	r10 := renderRow("q / Ctrl+C", "Quit application")

	content := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Width(boxWidth-4).Align(lipgloss.Center).Render(title),
		"",
		secNav,
		r1, r2, r3, r4,
		"",
		secChat,
		r5, r6,
		"",
		secActions,
		r7, r8, r9, r10,
		"",
		lipgloss.NewStyle().Width(boxWidth-4).Align(lipgloss.Center).Render(closeHint),
	)

	modal := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorCyan).
		Background(theme.ColorPanelBg).
		Padding(1, 2).
		Width(boxWidth).
		Render(content)

	if width > 0 && height > 0 {
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, modal)
	}
	return modal
}
