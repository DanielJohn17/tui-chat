package test

import (
	"strings"
	"testing"

	"github.com/DanielJohn17/tui-chat/internal/tui"
	"github.com/DanielJohn17/tui-chat/internal/tui/client"
	"github.com/DanielJohn17/tui-chat/internal/tui/views/auth"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestStatusBarLiveWithoutWS(t *testing.T) {
	model := authenticatedApp()
	view := model.View()

	// Must contain "● LIVE"
	assert.Contains(t, view, "LIVE")
	// Must NOT contain "(WS)"
	assert.NotContains(t, view, "(WS)")
}

func TestReconnectionFlowAndState(t *testing.T) {
	c := newTestClient()
	model := tea.Model(tui.NewApp(c))
	model, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 35})
	model, _ = model.Update(auth.AuthSuccessMsg{Profile: c.Profile()})

	// Initial status is LIVE
	assert.Contains(t, model.View(), "LIVE")

	// Simulate WebSocket connection drop
	c.SetWSConnected(false)
	model, cmd := model.Update(tui.WSIncomingMsg{
		Event: client.WSErrorPayload{
			Type:  "disconnected",
			Error: "websocket connection lost",
		},
	})
	assert.NotNil(t, cmd, "Should schedule a reconnect tick command")

	// View should reflect reconnecting state with retry seconds / connecting
	viewOffline := model.View()
	assert.True(t, strings.Contains(viewOffline, "RETRYING") || strings.Contains(viewOffline, "CONNECTING"))

	// Manual retry key 'r' in navigation mode
	model, retryCmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	assert.NotNil(t, retryCmd, "Pressing 'r' should return a command to reconnect")

	// Simulate successful reconnect result
	c.SetWSConnected(true)
	model, _ = model.Update(tui.WSConnectResultMsg{Err: nil})

	// View should be restored to LIVE
	assert.Contains(t, model.View(), "LIVE")
}

func TestStatusBarRetryCountdownAndLoadingAnimation(t *testing.T) {
	c := newTestClient()
	model := tea.Model(tui.NewApp(c))
	model, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 35})
	model, _ = model.Update(auth.AuthSuccessMsg{Profile: c.Profile()})

	// Simulate connect failure to enter Reconnecting state with countdown
	c.SetWSConnected(false)
	model, _ = model.Update(tui.WSConnectResultMsg{Err: assert.AnError})

	view := model.View()
	// Should show RETRYING with seconds countdown and spinner frame
	assert.Contains(t, view, "RETRYING")

	// Advance spinner frame
	model, _ = model.Update(tui.SpinnerTickMsg{})
	viewSpinner := model.View()
	assert.Contains(t, viewSpinner, "RETRYING")
}

func TestSpinnerStopsTickingWhenConnected(t *testing.T) {
	c := newTestClient()
	model := tea.Model(tui.NewApp(c))
	model, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 35})
	model, _ = model.Update(auth.AuthSuccessMsg{Profile: c.Profile()})

	// Successful connect result sets StatusConnected
	c.SetWSConnected(true)
	model, _ = model.Update(tui.WSConnectResultMsg{Err: nil})

	// When connected, handling SpinnerTickMsg must return nil Cmd (stops the ticker loop)
	model, cmd := model.Update(tui.SpinnerTickMsg{})
	assert.Nil(t, cmd, "SpinnerTickMsg must not schedule another tick command once connected")
}
