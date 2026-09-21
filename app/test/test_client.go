package test

import (
	"time"

	"github.com/DanielJohn17/tui-chat/app/internal/tui/client"
)

type TestClient struct {
	profile  client.Profile
	chats    []client.Chat
	messages map[int64][]client.Message
	msgID    int64
}

func newTestClient() *TestClient {
	now := time.Now()
	t := func(minutesAgo int) string {
		return now.Add(-time.Duration(minutesAgo) * time.Minute).Format("15:04")
	}

	chats := []client.Chat{
		{
			ID:          1,
			RecipientID: 102,
			Name:        "Bob Martin",
			Username:    "bob",
			LastMessage: "Fixed it, thanks for the review!",
			Time:        t(2),
			Unread:      0,
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
			Unread:      4,
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
	}

	messages := map[int64][]client.Message{
		1: {
			{ID: 1001, SenderID: 102, Sender: "bob", ConvID: 1, Text: "Hey Alice! Did you see the new status bar layout?", Timestamp: t(180), Self: false},
			{ID: 1002, SenderID: 101, Sender: "alice", ConvID: 1, Text: "Yes! Looks incredible. The muted grey separator is super clean.", Timestamp: t(175), Self: true},
			{ID: 1003, SenderID: 102, Sender: "bob", ConvID: 1, Text: "What do you think of the pill badges for NAV / INPUT mode?", Timestamp: t(160), Self: false},
			{ID: 1004, SenderID: 101, Sender: "alice", ConvID: 1, Text: "Huge usability improvement. Never get confused about active focus now.", Timestamp: t(155), Self: true},
			{ID: 1005, SenderID: 102, Sender: "bob", ConvID: 1, Text: "Also added keyboard hints like ? for Help Desk modal.", Timestamp: t(140), Self: false},
			{ID: 1006, SenderID: 101, Sender: "alice", ConvID: 1, Text: "Tested the F1 and Esc keys - modals open and close smoothly.", Timestamp: t(130), Self: true},
			{ID: 1007, SenderID: 102, Sender: "bob", ConvID: 1, Text: "Did you check the new message timestamp and unread badge colors?", Timestamp: t(120), Self: false},
			{ID: 1008, SenderID: 101, Sender: "alice", ConvID: 1, Text: "Yes, the neon cyan accent really pops in the dark theme.", Timestamp: t(115), Self: true},
			{ID: 1009, SenderID: 102, Sender: "bob", ConvID: 1, Text: "Awesome. I'm finishing up the WebSocket reconnect handler.", Timestamp: t(105), Self: false},
			{ID: 1010, SenderID: 101, Sender: "alice", ConvID: 1, Text: "Make sure exponential backoff is capped at 30 seconds.", Timestamp: t(100), Self: true},
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
		},
	}

	return &TestClient{
		profile: client.Profile{
			ID:        101,
			Name:      "Alice Walker",
			Username:  "alice",
			Token:     "",
			CreatedAt: "2026-09-01",
		},
		chats:    chats,
		messages: messages,
		msgID:    6000,
	}
}

func (m *TestClient) Login(username, password string) (*client.Profile, error) {
	m.profile.Username = username
	m.profile.Token = "logged-in-token"
	return &m.profile, nil
}

func (m *TestClient) Register(name, username, password string) (*client.Profile, error) {
	m.profile.Name = name
	m.profile.Username = username
	m.profile.Token = "registered-token"
	return &m.profile, nil
}

func (m *TestClient) IsAuthenticated() bool {
	return m.profile.Token != ""
}

func (m *TestClient) Profile() client.Profile {
	return m.profile
}

func (m *TestClient) SetProfile(p client.Profile) {
	m.profile = p
}

func (m *TestClient) UpdateProfile(p client.Profile) {
	m.profile = p
}

func (m *TestClient) FetchChats() ([]client.Chat, error) {
	return m.Chats(), nil
}

func (m *TestClient) Chats() []client.Chat {
	out := make([]client.Chat, len(m.chats))
	copy(out, m.chats)
	return out
}

func (m *TestClient) SetChats(chats []client.Chat) {
	m.chats = chats
}

func (m *TestClient) FetchMessages(chatID int64) ([]client.Message, error) {
	return m.Messages(chatID), nil
}

func (m *TestClient) Messages(chatID int64) []client.Message {
	msgs := m.messages[chatID]
	out := make([]client.Message, len(msgs))
	copy(out, msgs)
	return out
}

func (m *TestClient) AppendMessage(msg client.Message) {
	m.messages[msg.ConvID] = append(m.messages[msg.ConvID], msg)
}

func (m *TestClient) UpdateChatSnippet(convID int64, lastMsg, timeStr string, unreadDelta int, setExactUnread *int) {
	for i := range m.chats {
		if m.chats[i].ID == convID {
			if lastMsg != "" {
				m.chats[i].LastMessage = lastMsg
			}
			if timeStr != "" {
				m.chats[i].Time = timeStr
			}
			if setExactUnread != nil {
				m.chats[i].Unread = *setExactUnread
			} else if unreadDelta != 0 {
				m.chats[i].Unread += unreadDelta
			}
			if lastMsg != "" && i > 0 {
				target := m.chats[i]
				copy(m.chats[1:i+1], m.chats[0:i])
				m.chats[0] = target
			}
			break
		}
	}
}

func (m *TestClient) ConnectWS(eventsChan chan<- any) error {
	return nil
}

func (m *TestClient) CloseWS() error {
	return nil
}

func (m *TestClient) SendWS(convID int64, text string) error {
	m.msgID++
	m.AppendMessage(client.Message{
		ID:        m.msgID,
		SenderID:  m.profile.ID,
		Sender:    "You",
		ConvID:    convID,
		Text:      text,
		Timestamp: "now",
		Self:      true,
	})
	m.UpdateChatSnippet(convID, text, "now", 0, nil)
	return nil
}

func (m *TestClient) FocusConv(convID int64) error {
	return nil
}

func (m *TestClient) MarkRead(convID int64, messageID int64) error {
	zero := 0
	m.UpdateChatSnippet(convID, "", "", 0, &zero)
	return nil
}

func (m *TestClient) IsWSConnected() bool {
	return true
}
