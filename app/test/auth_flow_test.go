package test

import (
	"testing"

	"github.com/DanielJohn17/tui-chat/app/internal/tui"
	"github.com/DanielJohn17/tui-chat/app/internal/tui/client"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthPassThrough(t *testing.T) {
	c := client.NewMock()
	model := tea.Model(tui.NewApp(c))
	model, _ = model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	// 1. Test Enter passes straight to chat app without API call
	newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd, "expected command from Enter")

	// Execute cmd
	msg := cmd()
	newModel, _ = newModel.Update(msg)
	assert.Contains(t, newModel.View(), "CONVERSATIONS")
}

func TestAuthHelpModal(t *testing.T) {
	c := client.NewMock()
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
	c := client.NewMock()
	model := tea.Model(tui.NewApp(c))
	model, _ = model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	// 4. Test Esc in auth quits the app
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	require.NotNil(t, cmd, "expected quit cmd on Esc")
}
