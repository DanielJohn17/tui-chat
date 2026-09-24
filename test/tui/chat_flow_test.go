package test

import (
	"testing"

	"github.com/DanielJohn17/tui-chat/internal/tui"
	"github.com/DanielJohn17/tui-chat/internal/tui/client"
	"github.com/DanielJohn17/tui-chat/internal/tui/views/auth"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

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

func TestChatNotificationReceived(t *testing.T) {
	c := newTestClient()
	model := tea.Model(tui.NewApp(c))
	model, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 35})
	model, _ = model.Update(auth.AuthSuccessMsg{Profile: c.Profile()})
	chats, _ := c.FetchChats()
	model, _ = model.Update(tui.ChatsLoadedMsg{Chats: chats})

	// Initial active chat is Bob (ID 1)
	assert.Contains(t, model.View(), "@bob")

	// Inbound chat_notification event from Sarah Connor (ID 2)
	model, cmd := model.Update(tui.WSIncomingMsg{
		Event: client.WSChatNotificationPayload{
			ConvID:      2,
			SenderID:    103,
			SenderName:  "Sarah Connor",
			Content:     "Urgent: Check production logs!",
			UnreadCount: 5,
			CreatedAt:   "now",
		},
	})
	assert.NotNil(t, cmd, "Notification should return a command for auto-clear timer and bell")

	// 1. Status bar should display the notification banner
	view := model.View()
	assert.Contains(t, view, "Sarah Connor", "Notification should show sender name in status bar")
	assert.Contains(t, view, "Urgent: Check production logs!", "Notification should show message preview in status bar")

	// 2. Chat 2 should be at index 0 in chats with unread count 5
	assert.Equal(t, int64(2), c.Chats()[0].ID)
	assert.Equal(t, 5, c.Chats()[0].Unread)
	assert.Equal(t, "Urgent: Check production logs!", c.Chats()[0].LastMessage)

	// 3. Message should be appended to conversation 2 cache
	msgs := c.Messages(2)
	assert.GreaterOrEqual(t, len(msgs), 1)
	assert.Equal(t, "Urgent: Check production logs!", msgs[len(msgs)-1].Text)

	// 4. ClearNotificationMsg should clear the status bar notification banner
	model, _ = model.Update(tui.ClearNotificationMsg{ID: 1})
	viewCleared := model.View()
	assert.NotContains(t, viewCleared, "Urgent: Check production logs!")
}

func TestChatNotificationNewConversation(t *testing.T) {
	c := newTestClient()
	model := tea.Model(tui.NewApp(c))
	model, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 35})
	model, _ = model.Update(auth.AuthSuccessMsg{Profile: c.Profile()})
	chats, _ := c.FetchChats()
	model, _ = model.Update(tui.ChatsLoadedMsg{Chats: chats})

	// Inbound chat_notification for a completely brand new conversation (ConvID 999)
	model, _ = model.Update(tui.WSIncomingMsg{
		Event: client.WSChatNotificationPayload{
			ConvID:      999,
			SenderID:    888,
			SenderName:  "Grace Hopper",
			Content:     "Compilers are working!",
			UnreadCount: 1,
			CreatedAt:   "now",
		},
	})

	// 1. Brand new chat 999 must be created and placed at index 0 of sidebar
	assert.Equal(t, int64(999), c.Chats()[0].ID)
	assert.Equal(t, "Grace Hopper", c.Chats()[0].Name)
	assert.Equal(t, 1, c.Chats()[0].Unread)
	assert.Equal(t, "Compilers are working!", c.Chats()[0].LastMessage)

	// 2. Status bar notification should be displayed
	view := model.View()
	assert.Contains(t, view, "Grace Hopper")
	assert.Contains(t, view, "Compilers are working!")

	// 3. Message should be in message cache for ConvID 999
	msgs := c.Messages(999)
	assert.Len(t, msgs, 1)
	assert.Equal(t, "Compilers are working!", msgs[0].Text)
	assert.Equal(t, "Grace Hopper", msgs[0].Sender)
}

func TestBackgroundMessagesLoadingAndLoadingState(t *testing.T) {
	c := newTestClient()
	// Clear messages for chat 2 initially to test loading state
	c.messages[2] = nil
	c.SetChats([]client.Chat{
		{ID: 1, Name: "Bob", Username: "bob"},
		{ID: 2, Name: "Sarah", Username: "sarah"},
	})

	model := tea.Model(tui.NewApp(c))
	model, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 35})
	model, _ = model.Update(auth.AuthSuccessMsg{Profile: c.Profile()})

	// Dispatch ChatsLoadedMsg with chats
	model, cmd := model.Update(tui.ChatsLoadedMsg{
		Chats: []client.Chat{
			{ID: 1, Name: "Bob", Username: "bob"},
			{ID: 2, Name: "Sarah", Username: "sarah"},
		},
	})
	assert.NotNil(t, cmd, "ChatsLoadedMsg should batch-dispatch background prefetch commands")

	// Switch to conversation 2 while messages are loading
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	viewLoading := model.View()
	assert.Contains(t, viewLoading, "Loading conversation messages...")

	// Deliver bulk messages loaded
	c.messages[2] = []client.Message{
		{ID: 501, ConvID: 2, SenderID: 103, Sender: "Sarah", Text: "Hey there from Sarah!", Timestamp: "12:00"},
	}
	model, _ = model.Update(tui.BulkMessagesLoadedMsg{
		MessagesByConv: map[int64][]client.Message{
			2: c.messages[2],
		},
	})

	viewLoaded := model.View()
	assert.Contains(t, viewLoaded, "Hey there from Sarah!")
}

func TestConversationMessageLoadingFailureAndRetry(t *testing.T) {
	c := newTestClient()
	c.messages = make(map[int64][]client.Message)

	model := tea.Model(tui.NewApp(c))
	model, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 35})
	model, _ = model.Update(auth.AuthSuccessMsg{Profile: c.Profile()})

	model, _ = model.Update(tui.ChatsLoadedMsg{
		Chats: []client.Chat{
			{ID: 1, Name: "Bob", Username: "bob"},
		},
	})

	// Simulate failure loading messages for chat 1
	model, _ = model.Update(tui.MessagesLoadedMsg{
		ChatID: 1,
		Err:    assert.AnError,
	})

	viewFailed := model.View()
	assert.Contains(t, viewFailed, "Failed to load messages")
	assert.Contains(t, viewFailed, "Press [ r ] to retry")

	// 1. Press 'r' to retry
	model, retryCmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	assert.NotNil(t, retryCmd, "Pressing 'r' should return a command to re-fetch messages")
	assert.Contains(t, model.View(), "Loading conversation messages...")

	// Simulate failure again
	model, _ = model.Update(tui.MessagesLoadedMsg{
		ChatID: 1,
		Err:    assert.AnError,
	})

	// 2. Click on chat pane (X: 50, Y: 10) to retry
	model, clickRetryCmd := model.Update(tea.MouseMsg{
		X:      50,
		Y:      10,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	assert.NotNil(t, clickRetryCmd, "Clicking chat pane on error should return retry command")
	assert.Contains(t, model.View(), "Loading conversation messages...")
}
