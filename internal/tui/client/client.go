package client

import "github.com/DanielJohn17/tui-chat/internal/tui/crypto"
import "context"

type UserSummary struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
}

type Profile struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Username  string `json:"username"`
	Bio       string `json:"bio,omitempty"`
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
	IsSecret    bool   `json:"is_secret"`
	CloudID     int64  `json:"cloud_id,omitempty"`
	SecretID    int64  `json:"secret_id,omitempty"`
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
	IsSecret  bool   `json:"is_secret"`
}

// Inbound WebSocket Event Payloads
type WSChatMessagePayload struct {
	ID          int64  `json:"id"`
	SenderID    int64  `json:"sender_id"`
	RecipientID int64  `json:"recipient_id,omitempty"`
	ConvID      int64  `json:"conv_id"`
	Content     string `json:"content"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type WSChatNotificationPayload struct {
	ConvID      int64  `json:"conv_id"`
	SenderID    int64  `json:"sender_id"`
	SenderName  string `json:"sender_name"`
	Content     string `json:"content"`
	UnreadCount int    `json:"unread_count"`
	CreatedAt   string `json:"created_at"`
}

type WSConversationReadPayload struct {
	ConvID    int64 `json:"conv_id"`
	UserID    int64 `json:"user_id"`
	MessageID int64 `json:"message_id"`
}

type WSUserPresencePayload struct {
	UserID int64 `json:"user_id"`
	Online bool  `json:"online"`
}

type WSPresenceSnapshotPayload struct {
	OnlineUserIDs []int64 `json:"online_user_ids"`
}

type WSErrorPayload struct {
	Type      string `json:"type"`
	ConvID    int64  `json:"conv_id"`
	Content   string `json:"content"`
	Error     string `json:"error"`
	Retryable bool   `json:"retryable"`
}

type Client interface {
	// Auth operations
	Login(username, password string) (*Profile, error)
	Register(name, username, password string) (*Profile, error)
	Logout() error
	IsAuthenticated() bool
	Profile() Profile
	SetProfile(Profile)
	UpdateProfile(Profile)
	SessionPath() string
	SetSessionPath(string)

	// REST Data operations
	SearchUsers(context.Context, string) ([]UserSummary, error)
	OpenDirectConversation(context.Context, int64, ...bool) (Chat, error)
	FetchChats() ([]Chat, error)
	Chats() []Chat
	SetChats([]Chat)
	SetUserOnline(userID int64, online bool)
	SetOnlineUsers(userIDs []int64)
	FetchMessages(chatID int64) ([]Message, error)
	FetchBulkMessages() (map[int64][]Message, error)
	Messages(chatID int64) []Message
	AppendMessage(msg Message)
	UpdateChatSnippet(convID int64, lastMsg, timeStr string, unreadDelta int, setExactUnread *int)
	IsConvSecret(convID int64) bool
	SetConvSecret(convID int64, isSecret bool)
	FindChatByConvID(convID int64) (Chat, bool)

	// WebSocket operations
	ConnectWS(eventsChan chan<- any) error
	CloseWS() error
	SendWS(convID int64, text string) error
	FocusConv(convID int64) error
	MarkRead(convID int64, messageID int64) error
	IsWSConnected() bool

	// E2EE operations
	// SetIdentityKey stores the local X25519 identity key for use in all crypto operations.
	SetIdentityKey(key *crypto.IdentityKey)
	// IdentityKey returns the local identity key, or nil if not yet initialised.
	IdentityKey() *crypto.IdentityKey
	// UploadPublicKey publishes the local public key to the server so peers can derive shared keys.
	UploadPublicKey(pubKey []byte) error
	// GetPeerPublicKey fetches and caches a peer's public key from the server.
	GetPeerPublicKey(userID int64) ([]byte, error)
	// DecryptMessage derives the shared secret with peerID and decrypts ciphertext, handling legacy and fallback formats.
	DecryptMessage(peerID int64, ciphertext string) string
}
