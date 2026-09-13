// Package ws
package ws

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

type InboundMessage struct {
	ConvID  int64  `json:"conv_id,omitempty"`
	Content string `json:"content"`
}
