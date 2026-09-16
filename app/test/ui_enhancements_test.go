package test

import (
	"strings"
	"testing"

	"github.com/DanielJohn17/tui-chat/app/internal/tui"
	"github.com/DanielJohn17/tui-chat/app/internal/tui/client"
	"github.com/DanielJohn17/tui-chat/app/internal/tui/views/auth"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

func TestSidebarDimensionsAndNoOverflow(t *testing.T) {
	// Terminal 120x35
	m1 := authenticatedApp()
	v1 := m1.View()
	lines1 := strings.Split(v1, "\n")
	assert.Equal(t, 35, len(lines1), "Total rendered lines must exactly equal terminal height 35")
	for i, l := range lines1 {
		assert.Equal(t, 120, lipgloss.Width(l), "Line %d width must equal terminal width 120", i)
	}

	// Terminal 80x24
	c2 := client.NewMock()
	m2 := tea.Model(tui.NewApp(c2))
	m2, _ = m2.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m2, _ = m2.Update(auth.AuthSuccessMsg{Profile: c2.Profile()})
	v2 := m2.View()
	lines2 := strings.Split(v2, "\n")
	assert.Equal(t, 24, len(lines2), "Total rendered lines must exactly equal terminal height 24")
	for i, l := range lines2 {
		assert.Equal(t, 80, lipgloss.Width(l), "Line %d width must equal terminal width 80", i)
	}
}

func TestSidebarAllElementsClickable(t *testing.T) {
	// 1. Click [+n] button at top right of sidebar (Y: 1, X: 28) opens New DM modal
	m1 := authenticatedApp()
	mNewDM, _ := m1.Update(tea.MouseMsg{
		X:      28,
		Y:      1,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	assert.Contains(t, mNewDM.View(), "NEW DIRECT MESSAGE")

	// 2. Click bottom scroll indicator (line 26) advances selection down
	m2 := authenticatedApp()
	mScroll, _ := m2.Update(tea.MouseMsg{
		X:      15,
		Y:      26,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	viewScroll := mScroll.View()
	assert.NotContains(t, viewScroll, "PROFILE EDIT", "Clicking scroll indicator must NOT open profile")
	assert.Contains(t, viewScroll, "@sarah", "Selection must advance to Sarah Connor")

	// Scroll down further to push scrollOffset > 0
	for i := 0; i < 12; i++ {
		mScroll, _ = mScroll.Update(tea.MouseMsg{
			X:      15,
			Y:      26,
			Button: tea.MouseButtonLeft,
			Action: tea.MouseActionPress,
		})
	}
	assert.Contains(t, mScroll.View(), "more above", "Scrolling down past window must show top scroll indicator")

	// 3. Click top scroll indicator (line 3) scrolls back up
	mScrollUp, _ := mScroll.Update(tea.MouseMsg{
		X:      15,
		Y:      3,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	assert.Contains(t, mScrollUp.View(), "CONVERSATIONS")

	// 4. Click profile card (line 29: name/handle or line 30: hint) opens profile
	m3 := authenticatedApp()
	mProfile, _ := m3.Update(tea.MouseMsg{
		X:      15,
		Y:      29,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	assert.Contains(t, mProfile.View(), "PROFILE EDIT")

	// 5. Click on the rightmost column of an item (X: 30, Y: 8 for Charlie Zhang) selects item
	m4 := authenticatedApp()
	mRightEdge, _ := m4.Update(tea.MouseMsg{
		X:      30,
		Y:      8,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	assert.Contains(t, mRightEdge.View(), "@charlie")
}
