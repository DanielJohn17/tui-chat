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

type InboundMessage struct {
	ConvID      int64  `json:"conv_id,omitempty"`
	RecipientID int64  `json:"recipient_id,omitempty"`
	Content     string `json:"content"`
}
