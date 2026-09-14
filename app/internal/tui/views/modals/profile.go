package modals

import (
	"fmt"

	"github.com/DanielJohn17/tui-chat/app/internal/tui/client"
	"github.com/DanielJohn17/tui-chat/app/internal/tui/theme"
	"github.com/charmbracelet/lipgloss"
)

func RenderProfile(prof client.Profile, width, height int) string {
	boxWidth := 48
	if width > 0 && width < 52 {
		boxWidth = width - 4
	}

	title := theme.StyleTitle.Render("◈ USER PROFILE ◈")
	onlineBadge := theme.StyleSuccess.Render("● ONLINE")

	renderField := func(label, value string) string {
		l := theme.StyleDim.Render(label)
		v := lipgloss.NewStyle().Bold(true).Foreground(theme.ColorWhite).Render(value)
		spaces := boxWidth - 8 - lipgloss.Width(l) - lipgloss.Width(v)
		if spaces < 1 {
			spaces = 1
		}
		return lipgloss.JoinHorizontal(lipgloss.Center, l, lipgloss.NewStyle().Width(spaces).Render(""), v)
	}

	rName := renderField("Display Name:", prof.Name)
	rUser := renderField("Username:", "@"+prof.Username)
	rID := renderField("User ID:", fmt.Sprintf("#%d", prof.ID))

	tokenPreview := "None"
	if len(prof.Token) > 16 {
		tokenPreview = prof.Token[:8] + "..." + prof.Token[len(prof.Token)-8:]
	} else if prof.Token != "" {
		tokenPreview = prof.Token
	}
	rToken := renderField("Session Token:", tokenPreview)

	created := prof.CreatedAt
	if created == "" {
		created = "Active Session"
	}
	rCreated := renderField("Registered:", created)

	closeHint := theme.StyleDim.Render("[ Esc / Enter / p: Close ]")

	content := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Width(boxWidth-4).Align(lipgloss.Center).Render(title),
		lipgloss.NewStyle().Width(boxWidth-4).Align(lipgloss.Center).Render(onlineBadge),
		"",
		rName,
		rUser,
		rID,
		rToken,
		rCreated,
		"",
		lipgloss.NewStyle().Width(boxWidth-4).Align(lipgloss.Center).Render(closeHint),
	)

	modal := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorMagenta).
		Background(theme.ColorPanelBg).
		Padding(1, 2).
		Width(boxWidth).
		Render(content)

	if width > 0 && height > 0 {
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, modal)
	}
	return modal
}
