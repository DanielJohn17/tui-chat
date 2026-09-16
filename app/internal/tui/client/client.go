package client

type Profile struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Username  string `json:"username"`
	Password  string `json:"password,omitempty"`
	Token     string `json:"token,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

type Chat struct {
	ID          int64  `json:"id"`
	RecipientID int64  `json:"recipient_id"`
	Name        string `json:"name"`
	Username    string `json:"username"`
	LastMessage string `json:"last_message"`
	Time        string `json:"time"`
	Unread      int    `json:"unread"`
	Online      bool   `json:"online"`
}

type Message struct {
	ID        int64  `json:"id"`
	SenderID  int64  `json:"sender_id"`
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	ConvID    int64  `json:"conv_id"`
	Text      string `json:"text"`
	Timestamp string `json:"timestamp"`
	Self      bool   `json:"self"`
}

type Client interface {
	Chats() []Chat
	Messages(chatID int64) []Message
	Profile() Profile
	UpdateProfile(Profile)
	Send(chatID int64, text string)
	AddChat(name, username string) Chat
	MarkRead(chatID int64)

	// Auth operations
	Login(username, password string) (*Profile, error)
	Register(name, username, password string) (*Profile, error)
	IsAuthenticated() bool
	SetProfile(Profile)
}
