package test

import (
	"testing"

	"github.com/DanielJohn17/tui-chat/internal/tui"
	"github.com/DanielJohn17/tui-chat/internal/tui/views/auth"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthPassThrough(t *testing.T) {
	c := newTestClient()
	model := tea.Model(tui.NewApp(c))
	model, _ = model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	// Enter without inputs shows validation error
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Contains(t, model.View(), "Please enter your username")

	// Simulate successful login
	newModel, _ := model.Update(auth.AuthSuccessMsg{Profile: c.Profile()})
	assert.Contains(t, newModel.View(), "LINETALK")
}

func TestAuthHelpModal(t *testing.T) {
	c := newTestClient()
	model := tea.Model(tui.NewApp(c))
	model, _ = model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	// 2. Test F1 opens Help Desk modal
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyF1})
	assert.Contains(t, newModel.View(), "HELP DESK")

	// 3. Test Esc closes Help Desk modal and returns to auth card
	newModel, _ = newModel.Update(tea.KeyMsg{Type: tea.KeyEsc})
	assert.Contains(t, newModel.View(), "LOGIN")
}

func TestAuthEscQuits(t *testing.T) {
	c := newTestClient()
	model := tea.Model(tui.NewApp(c))
	model, _ = model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	// 4. Test Esc in auth quits the app
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	require.NotNil(t, cmd, "expected quit cmd on Esc")
}

func TestAuthMouseClicks(t *testing.T) {
	c := newTestClient()
	model := tea.Model(tui.NewApp(c))
	model, _ = model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	// 1. Click Register tab (X: 59, Y: 8)
	model, _ = model.Update(tea.MouseMsg{
		X:      59,
		Y:      8,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	assert.Contains(t, model.View(), "Display Name", "Clicking Register tab must switch to register form")

	// 2. Click Username field in Register mode (X: 39, Y: 14)
	model, _ = model.Update(tea.MouseMsg{
		X:      39,
		Y:      14,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})

	// 3. Click Password field in Register mode (X: 39, Y: 19)
	model, _ = model.Update(tea.MouseMsg{
		X:      39,
		Y:      19,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})

	// 4. Click Login tab to switch back (X: 29, Y: 6)
	model, _ = model.Update(tea.MouseMsg{
		X:      29,
		Y:      6,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	assert.NotContains(t, model.View(), "Display Name", "Clicking Login tab must return to login form")

	// 5. Click Password field in Login mode (X: 39, Y: 16)
	model, _ = model.Update(tea.MouseMsg{
		X:      39,
		Y:      16,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
}
