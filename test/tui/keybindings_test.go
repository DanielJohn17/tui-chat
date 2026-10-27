package test

import (
	"testing"

	"github.com/DanielJohn17/tui-chat/internal/tui/theme"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
)

func TestHelpKeybindings(t *testing.T) {
	model := authenticatedApp()

	// 1. 'h' should NOT open Help Desk anymore
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	assert.NotContains(t, model.View(), "HELP DESK")

	// 2. '?' and Ctrl+H open Help Desk; Esc closes it
	for _, openKey := range []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune{'?'}},
		{Type: tea.KeyCtrlH},
	} {
		model, _ = model.Update(openKey)
		assert.Contains(t, model.View(), "HELP DESK")
		model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
		assert.NotContains(t, model.View(), "HELP DESK")
	}

	// 3. Test input mode: Ctrl+H and F1 open Help Desk while typing
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	assert.Contains(t, model.View(), "INPUT MODE")

	for _, openKey := range []tea.KeyMsg{
		{Type: tea.KeyCtrlH},
		{Type: tea.KeyF1},
	} {
		model, _ = model.Update(openKey)
		assert.Contains(t, model.View(), "HELP DESK")
		model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
		assert.NotContains(t, model.View(), "HELP DESK")
	}
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
