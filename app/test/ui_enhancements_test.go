package test

import (
	"strings"
	"testing"

	"github.com/DanielJohn17/tui-chat/app/internal/tui"
	"github.com/DanielJohn17/tui-chat/app/internal/tui/client"
	"github.com/DanielJohn17/tui-chat/app/internal/tui/views/auth"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func authenticatedApp() tea.Model {
	c := client.NewMock()
	model := tea.Model(tui.NewApp(c))
	model, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 35})
	model, _ = model.Update(auth.AuthSuccessMsg{
		Profile: c.Profile(),
	})
	return model
}

func TestProfileEditPageFlow(t *testing.T) {
	model := authenticatedApp()

	// 1. Press 'p' to open Profile Edit page
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	assert.Contains(t, model.View(), "PROFILE EDIT")

	// 2. Press Esc to return to Chat
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	assert.Contains(t, model.View(), "CONVERSATIONS")

	// 3. Re-open Profile Edit page and Save
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	model, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		msg := cmd()
		model, _ = model.Update(msg)
	}
	assert.Contains(t, model.View(), "CONVERSATIONS")
}

func TestSidebarBottomUserInfoAndScrolling(t *testing.T) {
	model := authenticatedApp()
	view := model.View()

	// Verify user handle appears in the bottom sidebar
	assert.Contains(t, view, "@alice")

	// Verify mock contacts (at least 10+ in list)
	c := client.NewMock()
	assert.GreaterOrEqual(t, len(c.Chats()), 10)

	// Verify messages in Chat 1 is at least 20+
	assert.GreaterOrEqual(t, len(c.Messages(1)), 20)

	// Test mouse wheel down on sidebar scrolls down
	model, _ = model.Update(tea.MouseMsg{
		X:      10,
		Y:      10,
		Button: tea.MouseButtonWheelDown,
		Type:   tea.MouseWheelDown,
	})
	assert.Contains(t, model.View(), "CONVERSATIONS")
}

func TestHelpKeybindings(t *testing.T) {
	model := authenticatedApp()

	// 1. 'h' should NOT open Help Desk anymore
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	assert.NotContains(t, model.View(), "HELP DESK")

	// 2. '?' MUST open Help Desk
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	assert.Contains(t, model.View(), "HELP DESK")

	// 3. Esc closes Help Desk
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	assert.NotContains(t, model.View(), "HELP DESK")

	// 4. Ctrl+H MUST open Help Desk
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyCtrlH})
	assert.Contains(t, model.View(), "HELP DESK")
}

func TestMouseClickActions(t *testing.T) {
	model := authenticatedApp()

	initView := model.View()
	lines := strings.Split(initView, "\n")
	for idx, line := range lines {
		if idx < 20 {
			t.Logf("LINE %d: %s", idx, line)
		}
	}

	// 1. Mouse click on second chat (Sarah Connor, line 6)
	model, _ = model.Update(tea.MouseMsg{
		X:      10,
		Y:      6,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	assert.Contains(t, model.View(), "@sarah")

	// 2. Mouse click on bottom profile row in sidebar opens profile edit page
	model, _ = model.Update(tea.MouseMsg{
		X:      10,
		Y:      31,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	assert.Contains(t, model.View(), "PROFILE EDIT")
}

func TestSidebarClickWhenItemExpands(t *testing.T) {
	model := authenticatedApp()

	// 1. Click on Charlie Zhang (line 8)
	model, _ = model.Update(tea.MouseMsg{
		X:      10,
		Y:      8,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	assert.Contains(t, model.View(), "@charlie")

	// 2. Click on Alex Rivera (line 10)
	model, _ = model.Update(tea.MouseMsg{
		X:      10,
		Y:      10,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	assert.Contains(t, model.View(), "@alex")

	// 3. Click back on Bob Martin (line 3)
	model, _ = model.Update(tea.MouseMsg{
		X:      10,
		Y:      3,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	assert.Contains(t, model.View(), "@bob")
}
