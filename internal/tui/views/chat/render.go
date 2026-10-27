package chat

import (
	"github.com/DanielJohn17/tui-chat/internal/tui/client"
	"github.com/DanielJohn17/tui-chat/internal/tui/theme"
	"github.com/charmbracelet/lipgloss"
)

func renderHeader(activeChat *client.Chat, isFocused bool, contentWidth int) string {
	if activeChat == nil {
		return theme.StyleDim.Render("No active conversation")
	}

	var dot string
	if activeChat.Online {
		dot = theme.StyleSuccess.Render("●")
	} else {
		dot = theme.StyleDim.Render("○")
	}

	nameW := 14
	if contentWidth < 30 {
		nameW = 8
	}
	var bioSnippet string
	if activeChat.Bio != "" && contentWidth >= 50 {
		maxBioW := contentWidth - 35
		if maxBioW > 25 {
			maxBioW = 25
		}
		if maxBioW >= 8 {
			cleanBio := theme.Truncate(theme.SanitizeLine(activeChat.Bio), maxBioW)
			bioSnippet = theme.StyleDim.Render("• " + cleanBio)
		}
	}

	leftParts := []string{
		dot, " ",
		theme.StyleSubtitle.Render(theme.Truncate(theme.SanitizeLine(activeChat.Name), nameW)), " ",
		lipgloss.NewStyle().Foreground(theme.ColorMagenta).Render("@" + theme.Truncate(theme.SanitizeLine(activeChat.Username), nameW)),
	}
	if bioSnippet != "" {
		leftParts = append(leftParts, " ", bioSnippet)
	}
	leftTitle := lipgloss.JoinHorizontal(lipgloss.Center, leftParts...)

	if lipgloss.Width(leftTitle) > contentWidth {
		leftTitle = lipgloss.JoinHorizontal(lipgloss.Center,
			dot, " ",
			theme.StyleSubtitle.Render(theme.Truncate(theme.SanitizeLine(activeChat.Name), nameW)), " ",
			lipgloss.NewStyle().Foreground(theme.ColorMagenta).Render("@"+theme.Truncate(theme.SanitizeLine(activeChat.Username), nameW)),
		)
		if lipgloss.Width(leftTitle) > contentWidth {
			leftTitle = theme.Truncate(theme.SanitizeLine(activeChat.Name), contentWidth)
		}
	}

	var focusHint string
	if contentWidth >= 35 {
		if isFocused {
			focusHint = theme.StyleDim.Render("[Esc: unfocus]")
		} else {
			focusHint = theme.StyleDim.Render("[i: focus]")
		}
	}

	hSpaces := contentWidth - lipgloss.Width(leftTitle) - lipgloss.Width(focusHint)
	if hSpaces < 1 {
		hSpaces = 1
	}
	if focusHint != "" && lipgloss.Width(leftTitle)+lipgloss.Width(focusHint)+1 <= contentWidth {
		return lipgloss.JoinHorizontal(lipgloss.Center, leftTitle, lipgloss.NewStyle().Width(hSpaces).Render(""), focusHint)
	}
	return leftTitle
}

func renderTabBar(activeTab ChatTab, contentWidth int) string {
	var tabCloudStr, tabSecretStr string
	if contentWidth < 28 {
		if activeTab == TabCloud {
			tabCloudStr = lipgloss.NewStyle().Bold(true).Foreground(theme.ColorCyan).Render("[ 💬 ]")
			tabSecretStr = lipgloss.NewStyle().Foreground(theme.ColorDimText).Render("[ 🔒 ]")
		} else {
			tabCloudStr = lipgloss.NewStyle().Foreground(theme.ColorDimText).Render("[ 💬 ]")
			tabSecretStr = lipgloss.NewStyle().Bold(true).Foreground(theme.ColorMagenta).Render("[ 🔒 ]")
		}
	} else {
		if activeTab == TabCloud {
			tabCloudStr = lipgloss.NewStyle().Bold(true).Foreground(theme.ColorCyan).Render("[ 💬 Cloud ]")
			tabSecretStr = lipgloss.NewStyle().Foreground(theme.ColorDimText).Render("[ 🔒 Secret ]")
		} else {
			tabCloudStr = lipgloss.NewStyle().Foreground(theme.ColorDimText).Render("[ 💬 Cloud ]")
			tabSecretStr = lipgloss.NewStyle().Bold(true).Foreground(theme.ColorMagenta).Render("[ 🔒 Secret ]")
		}
	}
	tabBar := lipgloss.JoinHorizontal(lipgloss.Center, " ", tabCloudStr, "  ", tabSecretStr)
	if lipgloss.Width(tabBar) > contentWidth {
		tabBar = theme.Truncate(tabBar, contentWidth)
	}
	return tabBar
}

func renderSecretWarningBanner(activeTab ChatTab, peerHasKey bool, contentWidth int) string {
	if activeTab != TabSecret {
		return ""
	}

	if !peerHasKey {
		rawBanner := "🔒 Peer has not published an encryption key"
		return "  " + lipgloss.NewStyle().Bold(true).Foreground(theme.ColorRed).Render(theme.Truncate(rawBanner, contentWidth-2))
	}

	avail := contentWidth - 2
	var rawBanner string
	if avail >= 90 {
		rawBanner = "⚠️ Secret Chat: Device-bound encryption. Key migration: https://linetalk.vercel.app/docs"
	} else if avail >= 73 {
		rawBanner = "⚠️ Secret Chat: Device-bound. Docs: https://linetalk.vercel.app/docs"
	} else if avail >= 57 {
		rawBanner = "⚠️ Secret Chat: Device-bound (linetalk.vercel.app/docs)"
	} else if avail >= 30 {
		rawBanner = "⚠️ Secret Chat: Device-bound"
	} else {
		rawBanner = "⚠️ Secret Chat"
	}
	return "  " + lipgloss.NewStyle().Foreground(theme.ColorYellow).Render(theme.Truncate(rawBanner, avail))
}

func renderInputBox(inputView string, isFocused bool, contentWidth int) string {
	inputBorder := theme.ColorBorderDim
	if isFocused {
		inputBorder = theme.ColorCyan
	}
	inputBoxWidth := contentWidth - 2
	if inputBoxWidth < 1 {
		inputBoxWidth = 1
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(inputBorder).
		Width(inputBoxWidth).
		Padding(0, 1).
		Render(inputView)
}
