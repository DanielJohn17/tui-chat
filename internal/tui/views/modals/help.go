package modals

import (
	"github.com/DanielJohn17/tui-chat/internal/tui/theme"
	"github.com/charmbracelet/lipgloss"
)

func RenderHelp(width, height int) string {
	boxWidth := 58
	if width > 0 && width < 62 {
		boxWidth = width - 4
	}
	if boxWidth < 40 {
		boxWidth = 40
	}

	innerWidth := boxWidth - 6

	title := theme.StyleTitle.Render("◈ ") + theme.RenderLineTalkWordmark() + theme.StyleTitle.Render(" HELP DESK ◈")
	closeHint := theme.StyleDim.Render("[ Esc / Enter / ? / Ctrl+H / q: Close ]")

	renderRow := func(key, desc string) string {
		k := theme.StyleShortcut.Render(key)
		d := theme.StyleDim.Render(desc)
		spaces := innerWidth - lipgloss.Width(k) - lipgloss.Width(d)
		if spaces < 1 {
			spaces = 1
		}
		return lipgloss.JoinHorizontal(lipgloss.Center, k, lipgloss.NewStyle().Width(spaces).Render(""), d)
	}

	secAuth := theme.StyleSubtitle.Render("AUTHENTICATION & ACCESS")
	a1 := renderRow("Enter", "Login / Register (passes straight in)")
	a2 := renderRow("Ctrl+T", "Switch between Login and Register")
	a3 := renderRow("Tab / ↑↓", "Cycle through input fields")
	a4 := renderRow("? / Ctrl+H / F1", "Open this Help Desk anytime")

	secNav := theme.StyleSubtitle.Render("CHAT NAVIGATION & MOUSE")
	r1 := renderRow("j / ↓ / Wheel", "Move down / scroll conversations")
	r2 := renderRow("k / ↑ / Wheel", "Move up / scroll conversations")
	r3 := renderRow("i / Enter / Click", "Focus message input box")
	r4 := renderRow("Esc", "Unfocus input / Return to sidebar")

	secChat := theme.StyleSubtitle.Render("MESSAGING")
	r5 := renderRow("Enter", "Send active message")
	r6 := renderRow("PgUp / PgDn / Wheel", "Scroll chat history")

	secActions := theme.StyleSubtitle.Render("ACTIONS & MODALS")
	r7 := renderRow("n / Click [+n]", "Start new Direct Message")
	r8 := renderRow("p / Click Profile", "Open Profile Edit page")
	r9 := renderRow("? / Ctrl+H / F1", "Toggle Help Desk")
	r10 := renderRow("q / Ctrl+C", "Quit application")

	content := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(title),
		"",
		secAuth,
		a1, a2, a3, a4,
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
		lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(closeHint),
	)

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
