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
	assert.Contains(t, model.View(), "LINETALK")
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
	assert.Contains(t, mScrollUp.View(), "LINETALK")

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

func TestSidebarFocusRetentionOnInboundMessage(t *testing.T) {
	c := newTestClient()
	model := tea.Model(tui.NewApp(c))
	model, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 35})
	model, _ = model.Update(auth.AuthSuccessMsg{Profile: c.Profile()})
	chats, _ := c.FetchChats()
	model, _ = model.Update(tui.ChatsLoadedMsg{Chats: chats})

	// Initial active chat is Bob Martin (ID 1) at index 0
	assert.Equal(t, int64(1), c.Chats()[0].ID)
	assert.Contains(t, model.View(), "@bob")

	// Sarah Connor (ID 2) sends an inbound message
	model, _ = model.Update(tui.WSIncomingMsg{
		Event: client.WSChatMessagePayload{
			ID:        9999,
			SenderID:  103,
			ConvID:    2,
			Content:   "New message from Sarah while you were chatting with Bob",
			CreatedAt: "now",
		},
	})

	// Sarah (ID 2) is now bumped to index 0 of the chats slice
	assert.Equal(t, int64(2), c.Chats()[0].ID, "Sarah should move to top of chats slice")
	assert.Equal(t, int64(1), c.Chats()[1].ID, "Bob should move to second row in chats slice")

	// Active conversation in chat view MUST remain Bob (@bob)
	view := model.View()
	assert.Contains(t, view, "@bob", "Active chat view must still display Bob")

	// The sidebar focus arrow must follow Bob to row 2 instead of pointing to Sarah
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	for _, r := range "Still talking to you Bob!" {
		model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Bob must now be bumped back to index 0 with the sent message
	assert.Equal(t, int64(1), c.Chats()[0].ID)
	assert.Equal(t, "Still talking to you Bob!", c.Chats()[0].LastMessage)
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
