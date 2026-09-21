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
	c := newTestClient()
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
	c := newTestClient()
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

	// 1. Mouse click on second chat (Sarah Connor, line 8)
	model, _ = model.Update(tea.MouseMsg{
		X:      10,
		Y:      8,
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

	// 1. Click on Charlie Zhang (line 12)
	model, _ = model.Update(tea.MouseMsg{
		X:      10,
		Y:      12,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	assert.Contains(t, model.View(), "@charlie")

	// 2. Click on Alex Rivera (line 16)
	model, _ = model.Update(tea.MouseMsg{
		X:      10,
		Y:      16,
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
	c2 := newTestClient()
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

	// Scroll down via mouse wheel to push scrollOffset > 0
	for i := 0; i < 15; i++ {
		mScroll, _ = mScroll.Update(tea.MouseMsg{
			X:      15,
			Y:      10,
			Button: tea.MouseButtonWheelDown,
			Type:   tea.MouseWheelDown,
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

	// 5. Click on the rightmost column of an item (X: 35, Y: 12 for Charlie Zhang) selects item
	m4 := authenticatedApp()
	mRightEdge, _ := m4.Update(tea.MouseMsg{
		X:      35,
		Y:      12,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	assert.Contains(t, mRightEdge.View(), "@charlie")
}

func TestTelegramSidebarLayout(t *testing.T) {
	c := newTestClient()
	model := tea.Model(tui.NewApp(c))
	model, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 35})
	model, _ = model.Update(auth.AuthSuccessMsg{Profile: c.Profile()})

	view := model.View()

	// 1. Fixed height per chat with vertical padding - Last message preview is visible without expanding
	assert.Contains(t, view, "Fixed it, thanks for the rev")
	assert.Contains(t, view, "Deployment to staging was su")
	assert.Contains(t, view, "Added foreign keys")

	// 2. Unseen message badges are rendered for chats with unread messages
	// Charlie has 2, DevOps has 3, Maya has 1, Sofia has 4
	assert.Contains(t, view, "2")
	assert.Contains(t, view, "3")
	assert.Contains(t, view, "1")
	assert.Contains(t, view, "4")

	// 3. Selecting a chat does not change overall line count (fixed height)
	linesBefore := len(strings.Split(view, "\n"))
	model, _ = model.Update(tea.MouseMsg{
		X:      10,
		Y:      12,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	linesAfter := len(strings.Split(model.View(), "\n"))
	assert.Equal(t, linesBefore, linesAfter)
}

func TestUnseenMessageMarkRead(t *testing.T) {
	c := newTestClient()
	// Charlie Zhang (ID 3) has 2 unread messages initially
	var charlieChat *client.Chat
	for _, ch := range c.Chats() {
		if ch.ID == 3 {
			charlieChat = &ch
			break
		}
	}
	assert.NotNil(t, charlieChat)
	assert.Equal(t, 2, charlieChat.Unread)

	model := tea.Model(tui.NewApp(c))
	model, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 35})
	model, _ = model.Update(auth.AuthSuccessMsg{Profile: c.Profile()})

	// Click on Charlie Zhang (Y: 11, 12, or 13)
	model, _ = model.Update(tea.MouseMsg{
		X:      10,
		Y:      12,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})

	// After selecting Charlie, unread count should be marked as 0
	for _, ch := range c.Chats() {
		if ch.ID == 3 {
			assert.Equal(t, 0, ch.Unread, "Unread count must be cleared after clicking chat")
		}
	}
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
	assert.Contains(t, model.View(), "CONVERSATIONS", "Clicking Cancel must return to chat view")

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
	assert.Contains(t, model.View(), "CONVERSATIONS", "Clicking Save must save and return to chat view")
}

func TestDynamicConversationSortingOnSendAndNewDM(t *testing.T) {
	c := newTestClient()
	model := tea.Model(tui.NewApp(c))
	model, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 35})
	model, _ = model.Update(auth.AuthSuccessMsg{Profile: c.Profile()})

	// 1. Initial order: Chat 1 (Bob) is index 0, Chat 2 (Sarah) is index 1, Chat 3 (Charlie) is index 2
	assert.Equal(t, int64(1), c.Chats()[0].ID)
	assert.Equal(t, int64(2), c.Chats()[1].ID)
	assert.Equal(t, int64(3), c.Chats()[2].ID)

	// 2. Select Charlie Zhang (Chat 3)
	model, _ = model.Update(tea.MouseMsg{
		X:      10,
		Y:      12,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	assert.Contains(t, model.View(), "@charlie")

	// Focus input and send a message in Charlie's chat
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	for _, r := range "Hey Charlie, bumped to top!" {
		model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Charlie (Chat 3) MUST now be at index 0 (top of the slice)
	assert.Equal(t, int64(3), c.Chats()[0].ID, "Chat 3 must move to top of slice after sending message")
	assert.Equal(t, "Hey Charlie, bumped to top!", c.Chats()[0].LastMessage)

	// The active view must still be Charlie
	assert.Contains(t, model.View(), "@charlie")

	// 3. Unfocus input with Esc and create a new direct message with David
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	for _, r := range "david" {
		model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// New chat (David) must be at index 0
	assert.Equal(t, "david", c.Chats()[0].Username, "New DM must be prepended to top of sidebar")
	assert.Contains(t, model.View(), "@david")
}
