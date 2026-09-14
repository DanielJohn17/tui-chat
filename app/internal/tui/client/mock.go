package client

import (
	"fmt"
	"time"
)

type mockClient struct {
	chats    []Chat
	messages map[int64][]Message
	profile  Profile
	msgID    int64
}

func NewMock() Client {
	now := time.Now()
	t := func(minutesAgo int) string {
		return now.Add(-time.Duration(minutesAgo) * time.Minute).Format("15:04")
	}

	return &mockClient{
		chats: []Chat{
			{
				ID:          1,
				RecipientID: 102,
				Name:        "Bob Martin",
				Username:    "bob",
				LastMessage: "Fixed it, thanks for the review!",
				Time:        t(2),
				Unread:      1,
				Online:      true,
			},
			{
				ID:          2,
				RecipientID: 103,
				Name:        "Sarah Connor",
				Username:    "sarah",
				LastMessage: "Deployment to staging was successful 🚀",
				Time:        t(14),
				Unread:      0,
				Online:      true,
			},
			{
				ID:          3,
				RecipientID: 104,
				Name:        "Charlie Zhang",
				Username:    "charlie",
				LastMessage: "Check out the new database migration script",
				Time:        t(45),
				Unread:      2,
				Online:      false,
			},
			{
				ID:          4,
				RecipientID: 105,
				Name:        "Alex Rivera",
				Username:    "alex",
				LastMessage: "Are we still on for the design sync at 4?",
				Time:        t(120),
				Unread:      0,
				Online:      true,
			},
			{
				ID:          5,
				RecipientID: 999,
				Name:        "System Bot",
				Username:    "system",
				LastMessage: "Welcome to TUI Chat Terminal Messenger!",
				Time:        "Yesterday",
				Unread:      0,
				Online:      true,
			},
		},
		messages: map[int64][]Message{
			1: {
				{ID: 1001, SenderID: 999, Sender: "system", ConvID: 1, Text: "Direct message thread established with Bob Martin (@bob)", Timestamp: t(30), Self: false},
				{ID: 1002, SenderID: 102, Sender: "bob", ConvID: 1, Text: "Hey Alice! PR #3 is ready for your review.", Timestamp: t(25), Self: false},
				{ID: 1003, SenderID: 101, Sender: "alice", ConvID: 1, Text: "On it, checking the websocket hub logic now.", Timestamp: t(20), Self: true},
				{ID: 1004, SenderID: 101, Sender: "alice", ConvID: 1, Text: "Looks super clean! Just verify the readPump error handling.", Timestamp: t(10), Self: true},
				{ID: 1005, SenderID: 102, Sender: "bob", ConvID: 1, Text: "Fixed it, thanks for the review!", Timestamp: t(2), Self: false},
			},
			2: {
				{ID: 2001, SenderID: 103, Sender: "sarah", ConvID: 2, Text: "Hey Alice, pushing the new container image now.", Timestamp: t(40), Self: false},
				{ID: 2002, SenderID: 101, Sender: "alice", ConvID: 2, Text: "Great! Let me know when health checks pass.", Timestamp: t(35), Self: true},
				{ID: 2003, SenderID: 103, Sender: "sarah", ConvID: 2, Text: "Deployment to staging was successful 🚀", Timestamp: t(14), Self: false},
			},
			3: {
				{ID: 3001, SenderID: 104, Sender: "charlie", ConvID: 3, Text: "Check out the new database migration script", Timestamp: t(45), Self: false},
				{ID: 3002, SenderID: 104, Sender: "charlie", ConvID: 3, Text: "Added foreign keys with CASCADE delete on participants.", Timestamp: t(44), Self: false},
			},
			4: {
				{ID: 4001, SenderID: 105, Sender: "alex", ConvID: 4, Text: "Are we still on for the design sync at 4?", Timestamp: t(120), Self: false},
				{ID: 4002, SenderID: 101, Sender: "alice", ConvID: 4, Text: "Yes! Excited to see the new vibrant terminal palettes.", Timestamp: t(115), Self: true},
			},
			5: {
				{ID: 5001, SenderID: 999, Sender: "system", ConvID: 5, Text: "Welcome to TUI Chat Terminal Messenger! Direct messaging & real-time sockets enabled.", Timestamp: "Yesterday", Self: false},
			},
		},
		profile: Profile{
			ID:        101,
			Name:      "Alice Walker",
			Username:  "alice",
			Password:  "SecretPass123",
			Token:     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6MTAxLCJ1c2VybmFtZSI6ImFsaWNlIn0",
			CreatedAt: "2026-09-01 10:00:00",
		},
		msgID: 6000,
	}
}

func (m *mockClient) Chats() []Chat {
	return m.chats
}

func (m *mockClient) Messages(chatID int64) []Message {
	return m.messages[chatID]
}

func (m *mockClient) Profile() Profile {
	return m.profile
}

func (m *mockClient) UpdateProfile(p Profile) {
	m.profile = p
}

func (m *mockClient) Send(chatID int64, text string) {
	m.msgID++
	now := time.Now().Format("15:04")
	msg := Message{
		ID:        m.msgID,
		SenderID:  m.profile.ID,
		Sender:    m.profile.Username,
		ConvID:    chatID,
		Text:      text,
		Timestamp: now,
		Self:      true,
	}
	m.messages[chatID] = append(m.messages[chatID], msg)

	// Update last message in chat
	for i := range m.chats {
		if m.chats[i].ID == chatID {
			m.chats[i].LastMessage = text
			m.chats[i].Time = now
			break
		}
	}
}

func (m *mockClient) AddChat(name, username string) Chat {
	newID := int64(len(m.chats) + 1)
	newRecipientID := int64(200 + len(m.chats))
	chat := Chat{
		ID:          newID,
		RecipientID: newRecipientID,
		Name:        name,
		Username:    username,
		LastMessage: "Conversation started",
		Time:        "Just now",
		Unread:      0,
		Online:      true,
	}
	m.chats = append(m.chats, chat)
	m.messages[newID] = []Message{
		{
			ID:        m.msgID + 1,
			SenderID:  999,
			Sender:    "system",
			ConvID:    newID,
			Text:      fmt.Sprintf("Direct conversation created with %s (@%s)", name, username),
			Timestamp: time.Now().Format("15:04"),
			Self:      false,
		},
	}
	return chat
}
func (m *mockClient) Login(username, password string) (*Profile, error) {
	m.profile = Profile{
		ID:        1,
		Name:      username,
		Username:  username,
		Token:     "mock-jwt-token-authenticated",
		CreatedAt: time.Now().Format("2006-01-02"),
	}
	return &m.profile, nil
}

func (m *mockClient) Register(name, username, password string) (*Profile, error) {
	m.profile = Profile{
		ID:        1,
		Name:      name,
		Username:  username,
		Token:     "mock-jwt-token-registered",
		CreatedAt: time.Now().Format("2006-01-02"),
	}
	return &m.profile, nil
}

func (m *mockClient) IsAuthenticated() bool {
	return m.profile.Token != ""
}

func (m *mockClient) SetProfile(p Profile) {
	m.profile = p
}
