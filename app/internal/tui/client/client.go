package client

type Profile struct {
	Username    string
	AccountCode string
	Password    string
}

type Chat struct {
	ID          int
	Name        string
	LastMessage string
}

type Message struct {
	Text   string
	Sender string
	Self   bool
}

type Client interface {
	Chats() []Chat
	Messages(chatID int) []Message
	Profile() Profile
	UpdateProfile(Profile)
	Send(chatID int, text string)
}