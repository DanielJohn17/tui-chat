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
				LastMessage: "Added foreign keys with CASCADE delete on participants.",
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
				RecipientID: 106,
				Name:        "DevOps Bot",
				Username:    "devops-bot",
				LastMessage: "CI Pipeline #842 passed in 1m 42s [main]",
				Time:        t(140),
				Unread:      3,
				Online:      true,
			},
			{
				ID:          6,
				RecipientID: 107,
				Name:        "Elena Rostova",
				Username:    "elena",
				LastMessage: "The E2E test suite reported zero flakiness today.",
				Time:        t(190),
				Unread:      0,
				Online:      true,
			},
			{
				ID:          7,
				RecipientID: 108,
				Name:        "Marcus Vance",
				Username:    "marcus",
				LastMessage: "Security scan clean: 0 CVEs detected in Go dependencies.",
				Time:        t(240),
				Unread:      0,
				Online:      false,
			},
			{
				ID:          8,
				RecipientID: 109,
				Name:        "Maya Lin",
				Username:    "maya",
				LastMessage: "Updated the roadmap tickets for the v2 release.",
				Time:        t(310),
				Unread:      1,
				Online:      true,
			},
			{
				ID:          9,
				RecipientID: 110,
				Name:        "David K.",
				Username:    "davidk",
				LastMessage: "Postgres query latency dropped 45% after re-indexing.",
				Time:        t(380),
				Unread:      0,
				Online:      false,
			},
			{
				ID:          10,
				RecipientID: 111,
				Name:        "Liam Chen",
				Username:    "liam",
				LastMessage: "Terraform apply completed in eu-central-1.",
				Time:        t(450),
				Unread:      0,
				Online:      true,
			},
			{
				ID:          11,
				RecipientID: 112,
				Name:        "Sofia Patel",
				Username:    "sofia",
				LastMessage: "Terminal animations feeling butter-smooth now!",
				Time:        t(520),
				Unread:      0,
				Online:      true,
			},
			{
				ID:          12,
				RecipientID: 113,
				Name:        "Lucas Gray",
				Username:    "lucas",
				LastMessage: "Let's review the websocket multiplexing benchmarks.",
				Time:        t(600),
				Unread:      0,
				Online:      false,
			},
			{
				ID:          13,
				RecipientID: 114,
				Name:        "Rachel Green",
				Username:    "rachel",
				LastMessage: "Docs updated with keyboard shortcuts reference.",
				Time:        "Yesterday",
				Unread:      0,
				Online:      true,
			},
			{
				ID:          14,
				RecipientID: 115,
				Name:        "Viktor Hale",
				Username:    "viktor",
				LastMessage: "All prometheus metrics reporting 99.99% uptime.",
				Time:        "Yesterday",
				Unread:      0,
				Online:      false,
			},
			{
				ID:          15,
				RecipientID: 999,
				Name:        "System Bot",
				Username:    "system",
				LastMessage: "Welcome to TUI Chat Terminal Messenger!",
				Time:        "3d ago",
				Unread:      0,
				Online:      true,
			},
		},
		messages: map[int64][]Message{
			1: {
				{ID: 1001, SenderID: 999, Sender: "system", ConvID: 1, Text: "Direct message thread established with Bob Martin (@bob)", Timestamp: t(180), Self: false},
				{ID: 1002, SenderID: 102, Sender: "bob", ConvID: 1, Text: "Hey Alice! Hope your morning is going great.", Timestamp: t(170), Self: false},
				{ID: 1003, SenderID: 101, Sender: "alice", ConvID: 1, Text: "Morning Bob! Reviewing the sprint backlog right now.", Timestamp: t(165), Self: true},
				{ID: 1004, SenderID: 102, Sender: "bob", ConvID: 1, Text: "Did you check the new bubbletea terminal renderer?", Timestamp: t(160), Self: false},
				{ID: 1005, SenderID: 101, Sender: "alice", ConvID: 1, Text: "Yes, the frame rate and ANSI handling are fantastic.", Timestamp: t(155), Self: true},
				{ID: 1006, SenderID: 102, Sender: "bob", ConvID: 1, Text: "PR #3 is ready for review: refactored websocket client & reconnection loop.", Timestamp: t(145), Self: false},
				{ID: 1007, SenderID: 101, Sender: "alice", ConvID: 1, Text: "Awesome, pulling the branch to test locally.", Timestamp: t(140), Self: true},
				{ID: 1008, SenderID: 102, Sender: "bob", ConvID: 1, Text: "Take a look at internal/api/ws/client.go lines 80-120 in particular.", Timestamp: t(130), Self: false},
				{ID: 1009, SenderID: 101, Sender: "alice", ConvID: 1, Text: "Checking now. The exponential backoff on reconnect looks very resilient.", Timestamp: t(120), Self: true},
				{ID: 1010, SenderID: 102, Sender: "bob", ConvID: 1, Text: "We also need to make sure the ping/pong tickers don't leak goroutines.", Timestamp: t(110), Self: false},
				{ID: 1011, SenderID: 101, Sender: "alice", ConvID: 1, Text: "Good catch. Added a defer ticker.Stop() and close channel check.", Timestamp: t(95), Self: true},
				{ID: 1012, SenderID: 102, Sender: "bob", ConvID: 1, Text: "Perfect! How does mouse scrolling feel on your terminal?", Timestamp: t(85), Self: false},
				{ID: 1013, SenderID: 101, Sender: "alice", ConvID: 1, Text: "Smooth as silk! Wheel motion triggers viewport updates directly.", Timestamp: t(75), Self: true},
				{ID: 1014, SenderID: 102, Sender: "bob", ConvID: 1, Text: "What about the sidebar list? Can we scroll conversations too?", Timestamp: t(65), Self: false},
				{ID: 1015, SenderID: 101, Sender: "alice", ConvID: 1, Text: "Yes, mouse wheel over the sidebar scrolls through contacts now.", Timestamp: t(55), Self: true},
				{ID: 1016, SenderID: 102, Sender: "bob", ConvID: 1, Text: "And user profile pinned at the bottom?", Timestamp: t(45), Self: false},
				{ID: 1017, SenderID: 101, Sender: "alice", ConvID: 1, Text: "Pinned cleanly with a nice divider line at the bottom!", Timestamp: t(35), Self: true},
				{ID: 1018, SenderID: 102, Sender: "bob", ConvID: 1, Text: "Have you tested the profile edit screen yet?", Timestamp: t(25), Self: false},
				{ID: 1019, SenderID: 101, Sender: "alice", ConvID: 1, Text: "Just opened it with 'p' - display name, username and bio inputs all work.", Timestamp: t(20), Self: true},
				{ID: 1020, SenderID: 102, Sender: "bob", ConvID: 1, Text: "That's huge! Let me push a commit for the migration script.", Timestamp: t(15), Self: false},
				{ID: 1021, SenderID: 101, Sender: "alice", ConvID: 1, Text: "Go ahead, CI pipeline is ready to run the tests.", Timestamp: t(10), Self: true},
				{ID: 1022, SenderID: 102, Sender: "bob", ConvID: 1, Text: "Commit pushed: 7a82fc1 'feat: add cascade delete constraints'", Timestamp: t(5), Self: false},
				{ID: 1023, SenderID: 101, Sender: "alice", ConvID: 1, Text: "Review approved and merged! Great work Bob 🚀", Timestamp: t(3), Self: true},
				{ID: 1024, SenderID: 102, Sender: "bob", ConvID: 1, Text: "Fixed it, thanks for the review!", Timestamp: t(2), Self: false},
			},
			2: {
				{ID: 2001, SenderID: 103, Sender: "sarah", ConvID: 2, Text: "Hey Alice, pushing the new container image now.", Timestamp: t(40), Self: false},
				{ID: 2002, SenderID: 101, Sender: "alice", ConvID: 2, Text: "Great! Let me know when health checks pass.", Timestamp: t(35), Self: true},
				{ID: 2003, SenderID: 103, Sender: "sarah", ConvID: 2, Text: "Deployment to staging was successful 🚀", Timestamp: t(14), Self: false},
			},
			3: {
				{ID: 3001, SenderID: 104, Sender: "charlie", ConvID: 3, Text: "Check out the new database migration script", Timestamp: t(180), Self: false},
				{ID: 3002, SenderID: 104, Sender: "charlie", ConvID: 3, Text: "Added foreign keys with CASCADE delete on participants.", Timestamp: t(170), Self: false},
				{ID: 3003, SenderID: 101, Sender: "alice", ConvID: 3, Text: "Does this affect the soft-delete queries in sqlc?", Timestamp: t(160), Self: true},
				{ID: 3004, SenderID: 104, Sender: "charlie", ConvID: 3, Text: "No, soft delete keeps deleted_at intact. Hard cascade only applies on user purge.", Timestamp: t(150), Self: false},
				{ID: 3005, SenderID: 101, Sender: "alice", ConvID: 3, Text: "Perfect. Ran sqlc generate and no signature drift.", Timestamp: t(140), Self: true},
				{ID: 3006, SenderID: 104, Sender: "charlie", ConvID: 3, Text: "Also tuned connection pooling: max_conns=25, min_conns=5.", Timestamp: t(130), Self: false},
				{ID: 3007, SenderID: 101, Sender: "alice", ConvID: 3, Text: "Good choice for memory footprint in container limits.", Timestamp: t(120), Self: true},
				{ID: 3008, SenderID: 104, Sender: "charlie", ConvID: 3, Text: "Benchmark results show 15,000 req/sec with under 4ms p99 latency.", Timestamp: t(100), Self: false},
				{ID: 3009, SenderID: 101, Sender: "alice", ConvID: 3, Text: "Outstanding numbers Charlie! Terminal chat is blazing fast.", Timestamp: t(90), Self: true},
				{ID: 3010, SenderID: 104, Sender: "charlie", ConvID: 3, Text: "Let's sync up after standup on the cursor pagination.", Timestamp: t(45), Self: false},
			},
			4: {
				{ID: 4001, SenderID: 105, Sender: "alex", ConvID: 4, Text: "Are we still on for the design sync at 4?", Timestamp: t(120), Self: false},
				{ID: 4002, SenderID: 101, Sender: "alice", ConvID: 4, Text: "Yes! Excited to see the new vibrant terminal palettes.", Timestamp: t(115), Self: true},
			},
			5: {
				{ID: 5001, SenderID: 106, Sender: "devops-bot", ConvID: 5, Text: "[CI] Build started for commit 7a82fc1 on branch main", Timestamp: t(144), Self: false},
				{ID: 5002, SenderID: 106, Sender: "devops-bot", ConvID: 5, Text: "[CI] Running tests across 12 packages...", Timestamp: t(143), Self: false},
				{ID: 5003, SenderID: 106, Sender: "devops-bot", ConvID: 5, Text: "CI Pipeline #842 passed in 1m 42s [main]", Timestamp: t(140), Self: false},
			},
			15: {
				{ID: 9001, SenderID: 999, Sender: "system", ConvID: 15, Text: "Welcome to TUI Chat Terminal Messenger! Direct messaging & real-time sockets enabled.", Timestamp: "3d ago", Self: false},
			},
		},
		profile: Profile{
			ID:        101,
			Name:      "Alice Walker",
			Username:  "alice",
			Password:  "SecretPass123",
			Token:     "",
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
