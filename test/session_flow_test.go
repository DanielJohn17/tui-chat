package test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DanielJohn17/tui-chat/app/internal/tui"
	"github.com/DanielJohn17/tui-chat/app/internal/tui/client"
	"github.com/DanielJohn17/tui-chat/app/internal/tui/views/profile"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartupWithSavedSession(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tui-chat-test-session-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	sessionFile := filepath.Join(tempDir, "session.json")
	savedProfile := client.Profile{
		ID:        999,
		Name:      "Dev User",
		Username:  "devuser",
		Token:     "valid-session-token",
		CreatedAt: "2026-09-21T00:00:00Z",
	}
	err = client.SaveSession(sessionFile, savedProfile)
	require.NoError(t, err)

	// Simulate TUI startup
	c := newTestClient()
	loaded, err := client.LoadSession(sessionFile)
	require.NoError(t, err)
	require.NotNil(t, loaded)
	c.SetProfile(*loaded)
	c.SetSessionPath(sessionFile)

	assert.True(t, c.IsAuthenticated())

	app := tui.NewApp(c)
	model := tea.Model(app)
	model, _ = model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	// App should immediately display chat screen (e.g. LINETALK)
	view := model.View()
	assert.Contains(t, view, "LINETALK")
	assert.NotContains(t, view, "LOGIN")
}

func TestStartupWithoutSession(t *testing.T) {
	c := newTestClient()
	c.SetProfile(client.Profile{}) // unauthenticated

	assert.False(t, c.IsAuthenticated())

	app := tui.NewApp(c)
	model := tea.Model(app)
	model, _ = model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	// App should display auth / login screen
	view := model.View()
	assert.Contains(t, view, "LOGIN")
}

func TestSessionExpiryFallback(t *testing.T) {
	c := newTestClient()
	c.SetProfile(client.Profile{
		ID:       101,
		Username: "alice",
		Token:    "expired-token",
	})
	app := tui.NewApp(c)
	model := tea.Model(app)
	model, _ = model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	// Initial view is chat because token was present
	assert.Contains(t, model.View(), "LINETALK")

	// Simulate receiving 401 Unauthorized from fetchChatsCmd
	model, _ = model.Update(tui.ChatsLoadedMsg{
		Err: errors.New("Failed to fetch conversations (HTTP 401: unauthorized)"),
	})

	// Should fallback to Auth state with session expired message
	assert.False(t, c.IsAuthenticated())
	view := model.View()
	assert.Contains(t, view, "LOGIN")
	assert.Contains(t, view, "Session expired")
}

func TestLogoutFlow(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tui-chat-test-logout-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	sessionFile := filepath.Join(tempDir, "session.json")
	_ = client.SaveSession(sessionFile, client.Profile{
		ID:       101,
		Username: "alice",
		Token:    "alice-token",
	})

	c := newTestClient()
	c.SetProfile(client.Profile{ID: 101, Username: "alice", Token: "alice-token"})
	c.SetSessionPath(sessionFile)

	app := tui.NewApp(c)
	model := tea.Model(app)
	model, _ = model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	// Switch to profile view
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	assert.Contains(t, model.View(), "EDIT USER PROFILE")

	// Trigger Logout
	model, _ = model.Update(profile.LogoutMsg{})

	// Should transition to Auth state
	assert.False(t, c.IsAuthenticated())
	view := model.View()
	assert.Contains(t, view, "LOGIN")
	assert.Contains(t, view, "Logged out successfully")
}

func TestMultiWindowMessageSync(t *testing.T) {
	// Simulate Window 2 (sibling client) of Alice
	aliceProfile := client.Profile{
		ID:       101,
		Name:     "Alice",
		Username: "alice",
		Token:    "alice-token",
	}

	c := newTestClient()
	c.SetProfile(aliceProfile)

	app := tui.NewApp(c)
	model := tea.Model(app)
	model, _ = model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	// Populate initial chats
	model, _ = model.Update(tui.ChatsLoadedMsg{
		Chats: c.Chats(),
	})

	// Simulate Alice sending a message from Window 1 -> server echoes to Window 2 via WebSocket
	model, _ = model.Update(tui.WSIncomingMsg{
		Event: client.WSChatMessagePayload{
			ID:        9991,
			SenderID:  aliceProfile.ID, // Self (from sibling window)
			ConvID:    1,               // active chat (Bob Martin)
			Content:   "Hello from my second terminal window!",
			CreatedAt: "2026-09-21T21:00:00Z",
		},
	})

	// Verify message is visible in Window 2
	view := model.View()
	assert.Contains(t, view, "Hello from my second terminal window!")
	assert.Contains(t, view, "You")
}

func TestDevModeSessionIsolation(t *testing.T) {
	// Verify that in development mode, different SESSION names yield isolated session files
	alicePath := client.ResolveSessionPath("development", "alice")
	bobPath := client.ResolveSessionPath("development", "bob")
	defaultPath := client.ResolveSessionPath("development", "")

	assert.NotEqual(t, alicePath, bobPath)
	assert.NotEqual(t, alicePath, defaultPath)
	assert.Contains(t, alicePath, "dev_session_alice.json")
	assert.Contains(t, bobPath, "dev_session_bob.json")
	assert.Contains(t, defaultPath, "dev_session.json")
}
