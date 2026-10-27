package test

import (
	"strings"
	"testing"

	"github.com/DanielJohn17/tui-chat/internal/tui"
	"github.com/DanielJohn17/tui-chat/internal/tui/views/auth"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChatTabs_KeyboardAndMouseToggling(t *testing.T) {
	c := newTestClient()
	// Set dummy 32-byte key for Bob (102)
	c.SetPeerPublicKey(102, make([]byte, 32))

	app := tea.Model(tui.NewApp(c))
	app, _ = app.Update(tea.WindowSizeMsg{Width: 140, Height: 26})
	app, _ = app.Update(auth.AuthSuccessMsg{Profile: c.Profile()})

	// Initial view: Bob Martin is selected, TabCloud is active by default
	view := app.View()
	assert.Contains(t, view, "💬 Cloud")
	assert.Contains(t, view, "🔒 Secret")
	assert.NotContains(t, view, "⚠️ Secret Chat: Device-bound encryption", "Cloud mode must not show secret banner")

	// 1. Toggle to Secret tab using Tab key
	app, _ = app.Update(tea.KeyMsg{Type: tea.KeyTab})
	view = app.View()
	assert.Contains(t, view, "⚠️ Secret Chat: Device-bound encryption", "Secret mode must show device-bound banner")
	assert.Contains(t, view, "https://linetalk.vercel.app/docs", "Banner must contain docs link")

	// 2. Toggle back to Cloud tab using Ctrl+S key
	app, _ = app.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	view = app.View()
	assert.NotContains(t, view, "⚠️ Secret Chat: Device-bound encryption")

	// 3. Toggle to Secret tab using mouse click on Secret tab (row 2 in chat pane, X = 65)
	app, _ = app.Update(tea.MouseMsg{
		X:      65,
		Y:      2,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	view = app.View()
	assert.Contains(t, view, "⚠️ Secret Chat: Device-bound encryption")

	// 4. Toggle back to Cloud tab using mouse click on Cloud tab (row 2 in chat pane, X = 45)
	app, _ = app.Update(tea.MouseMsg{
		X:      45,
		Y:      2,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	view = app.View()
	assert.NotContains(t, view, "⚠️ Secret Chat: Device-bound encryption")
}

func TestChatTabs_MissingPeerKeyDisablesInput(t *testing.T) {
	c := newTestClient()
	// Do NOT set public key for Bob (102) -> peer has no key

	app := tea.Model(tui.NewApp(c))
	app, _ = app.Update(tea.WindowSizeMsg{Width: 90, Height: 26})
	app, _ = app.Update(auth.AuthSuccessMsg{Profile: c.Profile()})

	// Switch to Secret tab
	app, _ = app.Update(tea.KeyMsg{Type: tea.KeyTab})
	view := app.View()

	// Should show missing key banner
	assert.Contains(t, view, "🔒 Peer has not published an encryption key")
	assert.Contains(t, view, "Message input disabled")

	// Attempting to focus input with 'i' or click should not allow typing/sending
	app, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	app, _ = app.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// SendWS should not have added any message
	msgs := c.Messages(1)
	for _, m := range msgs {
		assert.NotEqual(t, "i", m.Text)
	}
}

func TestChatTabs_SidebarSecretBadge(t *testing.T) {
	c := newTestClient()
	chats := c.Chats()
	require.NotEmpty(t, chats)

	// Mark chat[0] as secret
	chats[0].IsSecret = true
	chats[0].LastMessage = "Top secret message"
	c.SetChats(chats)

	app := tea.Model(tui.NewApp(c))
	app, _ = app.Update(tea.WindowSizeMsg{Width: 90, Height: 26})
	app, _ = app.Update(auth.AuthSuccessMsg{Profile: c.Profile()})

	view := app.View()
	assert.Contains(t, view, "🔒 Top secret message")
}

func TestChatTabs_SecretModeZeroWrappingAcrossWidths(t *testing.T) {
	widths := []int{45, 50, 60, 70, 90, 110, 140}
	for _, w := range widths {
		c := newTestClient()
		c.SetPeerPublicKey(102, make([]byte, 32))
		app := tea.Model(tui.NewApp(c))
		app, _ = app.Update(tea.WindowSizeMsg{Width: w, Height: 24})
		app, _ = app.Update(auth.AuthSuccessMsg{Profile: c.Profile()})

		// Toggle to secret mode
		app, _ = app.Update(tea.KeyMsg{Type: tea.KeyTab})

		v := app.View()
		lines := strings.Split(v, "\n")
		assert.Equal(t, 24, len(lines), "Secret mode height must be 24 for width %d", w)
		for i, l := range lines {
			lineWidth := lipgloss.Width(l)
			assert.Equal(t, w, lineWidth, "Line %d width must equal terminal width %d (got %d)", i, w, lineWidth)
		}
	}
}
