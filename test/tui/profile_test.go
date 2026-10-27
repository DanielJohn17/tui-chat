package test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

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
	assert.Contains(t, view, "Sync:", "Profile view must display Sync storage notice")
}

func TestProfileBioPersistence(t *testing.T) {
	model := authenticatedApp()

	// 1. Open profile edit page via 'p'
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	assert.Contains(t, model.View(), "PAGE: PROFILE EDIT")

	// 2. Press Tab twice to focus Bio field (0=Name, 1=Username, 2=Bio)
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})

	// Enter new bio text
	for _, r := range "Building Go TUIs 💻" {
		model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	// 3. Save profile (Tab 4 times: Privacy -> Copy -> Regen -> Save)
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		msg := cmd()
		model, _ = model.Update(msg)
	}

	// 4. Re-open profile and verify bio is preserved
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	view := model.View()
	assert.Contains(t, view, "Building Go TUIs")
}

func TestProfilePrivacyToggleAndInviteRegen(t *testing.T) {
	model := authenticatedApp()

	// Open profile edit page via 'p'
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	view := model.View()
	assert.Contains(t, view, "Privacy:")
	assert.Contains(t, view, "[●] Public (Searchable)")
	assert.Contains(t, view, "tui://invite/inv_alice_test_token_123")

	// Tab to Privacy field (0=Name, 1=Username, 2=Bio, 3=Privacy)
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})

	// Toggle privacy via Space
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeySpace})
	view = model.View()
	assert.Contains(t, view, "[●] Invite-Only")

	// Tab to Copy button (focus 4)
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	view = model.View()
	assert.Contains(t, view, "Link copied")

	// Tab to Regen button (focus 5)
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	view = model.View()
	assert.Contains(t, view, "tui://invite/inv_test_regenerated")
	assert.Contains(t, view, "New invite link generated")

	// Tab to Save button (focus 6)
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		msg := cmd()
		model, _ = model.Update(msg)
	}
	assert.Contains(t, model.View(), "LINETALK")
}

func TestProfileCopyInviteLinkHotkey(t *testing.T) {
	model := authenticatedApp()

	// Open profile edit page via 'p'
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	// Navigate out of text fields (tab 3 times to Privacy)
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})

	// Press 'c' to copy invite link
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	view := model.View()
	assert.Contains(t, view, "Link copied")
}
