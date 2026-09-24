package test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestProfileEditPageFlow(t *testing.T) {
	model := authenticatedApp()

	// 1. Press 'p' to open Profile Edit page
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	assert.Contains(t, model.View(), "PROFILE EDIT")

	// 2. Press Esc to return to Chat
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	assert.Contains(t, model.View(), "LINETALK")

	// 3. Re-open Profile Edit page and Save
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	model, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		msg := cmd()
		model, _ = model.Update(msg)
	}
	assert.Contains(t, model.View(), "LINETALK")
}

func TestProfilePageMouseClicks(t *testing.T) {
	model := authenticatedApp()

	// Open profile edit page via 'p'
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	assert.Contains(t, model.View(), "PROFILE EDIT")

	// Terminal is 120x35. Card is 68 wide, 33 high. startX = 26, startY = 1.
	// Click Username field (startY + 13 = 14)
	model, _ = model.Update(tea.MouseMsg{
		X:      36,
		Y:      14,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})

	// Click Status / Bio field (startY + 18 = 19)
	model, _ = model.Update(tea.MouseMsg{
		X:      36,
		Y:      19,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})

	// Click Display Name field (startY + 8 = 9)
	model, _ = model.Update(tea.MouseMsg{
		X:      36,
		Y:      9,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})

	// Click Cancel button (startY + 27 = 28, X = startX + 45 = 71)
	model, cmd := model.Update(tea.MouseMsg{
		X:      71,
		Y:      28,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	if cmd != nil {
		msg := cmd()
		model, _ = model.Update(msg)
	}
	assert.Contains(t, model.View(), "LINETALK", "Clicking Cancel must return to chat view")

	// Re-open profile and click Save button (Y = 28, X = startX + 15 = 41)
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	model, cmdSave := model.Update(tea.MouseMsg{
		X:      41,
		Y:      28,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	if cmdSave != nil {
		msg := cmdSave()
		model, _ = model.Update(msg)
	}
	assert.Contains(t, model.View(), "LINETALK", "Clicking Save must save and return to chat view")
}

func TestProfileDateFormattingAndNoUserID(t *testing.T) {
	model := authenticatedApp()

	// Open profile edit page via 'p'
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	view := model.View()

	assert.Contains(t, view, "PAGE: PROFILE EDIT")
	assert.NotContains(t, view, "User ID:", "Profile view should not expose User ID")
	assert.Contains(t, view, "Member Since:", "Profile view must display Member Since")
	assert.Contains(t, view, "Status:", "Profile view must display Status")
}
