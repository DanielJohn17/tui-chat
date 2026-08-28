package client

type mockClient struct {
	chats    []Chat
	messages map[int][]Message
	profile  Profile
	msgID    int
}

func NewMock() Client {
	return &mockClient{
		chats: []Chat{
			{ID: 0, Name: "general", LastMessage: "Hey everyone!"},
			{ID: 1, Name: "random", LastMessage: "check this meme"},
			{ID: 2, Name: "dev-team", LastMessage: "PR ready for review"},
			{ID: 3, Name: "off-topic", LastMessage: "anyone watching..."},
			{ID: 4, Name: "announcements", LastMessage: "Welcome aboard!"},
		},
		messages: map[int][]Message{
			0: {
				{Text: "Welcome to #general!", Sender: "system", Self: false},
				{Text: "Hey everyone! How's it going?", Sender: "alice", Self: true},
				{Text: "Doing great, working on the new feature", Sender: "bob", Self: false},
				{Text: "Sounds awesome, let me know if you need help", Sender: "alice", Self: true},
			},
			1: {
				{Text: "This is #random — anything goes", Sender: "system", Self: false},
				{Text: "check out this meme I found", Sender: "charlie", Self: false},
				{Text: "lol that's hilarious", Sender: "alice", Self: true},
			},
			2: {
				{Text: "Dev team channel", Sender: "system", Self: false},
				{Text: "PR #42 is ready for review", Sender: "bob", Self: false},
				{Text: "On it, will review after lunch", Sender: "alice", Self: true},
				{Text: "Looks good, just one small nit on line 87", Sender: "alice", Self: true},
				{Text: "Fixed it, thanks for the review!", Sender: "bob", Self: false},
			},
			3: {
				{Text: "Off-topic — talk about anything here", Sender: "system", Self: false},
				{Text: "anyone watching the new season?", Sender: "charlie", Self: false},
				{Text: "No spoilers! I'm only on episode 3", Sender: "alice", Self: true},
			},
			4: {
				{Text: "Welcome to the team, Alice!", Sender: "admin", Self: false},
				{Text: "Thanks everyone, excited to be here!", Sender: "alice", Self: true},
			},
		},
		profile: Profile{
			Username:    "alice",
			AccountCode: "AC-0001",
			Password:    "supersecret",
		},
		msgID: 100,
	}
}

func (m *mockClient) Chats() []Chat {
	return m.chats
}

func (m *mockClient) Messages(chatID int) []Message {
	return m.messages[chatID]
}

func (m *mockClient) Profile() Profile {
	return m.profile
}

func (m *mockClient) UpdateProfile(p Profile) {
	m.profile = p
}

func (m *mockClient) Send(chatID int, text string) {
	m.msgID++
	msg := Message{Text: text, Sender: "alice", Self: true}
	m.messages[chatID] = append(m.messages[chatID], msg)
}