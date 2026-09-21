// Package ws
package ws

import "encoding/json"

type ActionType string
type WSNotificationType string

const (
	ActionSendMessage ActionType = "send_message"
	ActionFocusConv   ActionType = "focus_conv"
	ActionMarkRead    ActionType = "mark_read"
)

const (
	TypeChatMessage      WSNotificationType = "chat_message"      // Recipient is actively viewing this conv
	TypeChatNotification WSNotificationType = "chat_notification" // Recipient is in another conv or idle
	TypeConversationRead WSNotificationType = "conversation_read" // Synced read status
)

type WSMessage struct {
	ID          int64  `json:"id"`
	SenderID    int64  `json:"sender_id"`
	RecipientID int64  `json:"recipient_id,omitempty"`
	ConvID      int64  `json:"conv_id"`
	Content     string `json:"content"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type WSMessageError struct {
	Type      string `json:"type"` // error
	ConvID    int64  `json:"conv_id"`
	Content   string `json:"content"` // text content to retry
	Error     string `json:"error"`   // human readable error message
	Retryable bool   `json:"retryable"`
}

// Individual Message Payload Structs

type SendMessagePayload struct {
	ConvID  int64  `json:"conv_id" validate:"gte=1"`
	Content string `json:"content" validate:"gte=1"`
}

type FocusConvPayload struct {
	ConvID int64 `json:"conv_id" validate:"gte=1"`
}

type MarkReadPayload struct {
	ConvID    int64 `json:"conv_id"    validate:"gte=1"`
	MessageID int64 `json:"message_id" validate:"gte=1"`
}

type InboundMessage struct {
	Action  ActionType      `json:"action"`
	Payload json.RawMessage `json:"payload"`
}

type WSNotification struct {
	Type    WSNotificationType `json:"type"`
	Payload any                `json:"payload"`
}
type ChatNotificationPayload struct {
	ConvID      int64  `json:"conv_id"`
	SenderID    int64  `json:"sender_id"`
	SenderName  string `json:"sender_name"`
	Content     string `json:"content"`
	UnreadCount int    `json:"unread_count"`
	CreatedAt   string `json:"created_at"`
}

type ConversationReadPayload struct {
	ConvID    int64 `json:"conv_id"`
	UserID    int64 `json:"user_id"`    // The user who read the chat
	MessageID int64 `json:"message_id"` // Read up to this message ID
}
