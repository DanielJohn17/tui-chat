package test

import (
	"testing"

	"github.com/DanielJohn17/tui-chat/app/internal/tui/theme"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
)

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

func TestCtrlHInInputModeOpensHelp(t *testing.T) {
	model := authenticatedApp()

	// 1. Focus input mode via 'i'
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	assert.Contains(t, model.View(), "INPUT MODE")

	// 2. Press Ctrl+H while in input mode
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyCtrlH})
	assert.Contains(t, model.View(), "HELP DESK", "Ctrl+H in input mode must open Help Desk modal")

	// 3. Press Esc to close Help Desk modal
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	assert.NotContains(t, model.View(), "HELP DESK", "Esc must close Help Desk modal")

	// 4. Press F1 while in input mode
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyF1})
	assert.Contains(t, model.View(), "HELP DESK", "F1 in input mode must open Help Desk modal")
}

func TestUTF8EmojiTruncation(t *testing.T) {
	msgWithEmoji := "Sounds great! See you then 👍"
	// Should not corrupt emoji bytes or produce \uFFFD replacement char
	res := theme.Truncate(msgWithEmoji, 30)
	assert.Equal(t, msgWithEmoji, res)
	assert.NotContains(t, res, "\uFFFD")

	truncated := theme.Truncate(msgWithEmoji, 20)
	assert.NotContains(t, truncated, "\uFFFD")
	assert.True(t, lipgloss.Width(truncated) <= 20)
}
