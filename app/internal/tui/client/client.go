package client

type Profile struct {
	ID        int64
	Name      string
	Username  string
	Password  string
	Token     string
	CreatedAt string
}

type Chat struct {
	ID          int64
	RecipientID int64
	Name        string
	Username    string
	LastMessage string
	Time        string
	Unread      int
	Online      bool
}

type Message struct {
	ID        int64
	SenderID  int64
	Sender    string
	Recipient string
	ConvID    int64
	Text      string
	Timestamp string
	Self      bool
}

type Client interface {
	Chats() []Chat
	Messages(chatID int64) []Message
	Profile() Profile
	UpdateProfile(Profile)
	Send(chatID int64, text string)
	AddChat(name, username string) Chat
}