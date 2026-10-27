package auth

import (
	"fmt"
	"strings"

	"github.com/DanielJohn17/tui-chat/internal/tui/theme"
	"github.com/charmbracelet/lipgloss"
)

func renderTabs(mode Mode, innerWidth int) string {
	var tabLogin, tabRegister string
	if mode == ModeLogin {
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

	return lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(
		lipgloss.JoinHorizontal(lipgloss.Center, tabLogin, "  ", tabRegister),
	)
}

func renderField(label string, inputView string, focused bool, inputWidth int) string {
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

func renderRegisterChecklist(regPwd string) string {
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
	return lipgloss.JoinVertical(lipgloss.Left, hintRow1, hintRow2)
}

func renderErrorBanner(errorMessage string, inputWidth int) string {
	if errorMessage == "" {
		return ""
	}
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.ColorRed).
		Background(lipgloss.Color("#2A0812")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorRed).
		Padding(0, 1).
		Width(inputWidth).
		Render("✖ " + errorMessage)
}

func renderSubmitButton(loading bool, mode Mode, spinnerView string, innerWidth int) string {
	var actionText string
	if loading {
		actionText = fmt.Sprintf("%s Authenticating with API server...", spinnerView)
	} else if mode == ModeRegister {
		actionText = "[ Enter: Create Account & Enter Chat ]"
	} else {
		actionText = "[ Enter: Login & Enter Chat ]"
	}
	submitAction := theme.StyleSuccess.Render(actionText)
	return lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(submitAction)
}

func renderFooter(innerWidth int) string {
	hint1 := theme.StyleDim.Render("Tab: next field  •  Ctrl+T: switch mode")
	hint2 := theme.StyleDim.Render("? / Ctrl+H / F1: help  •  Esc: quit")
	return lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(hint1),
		lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(hint2),
	)
}
