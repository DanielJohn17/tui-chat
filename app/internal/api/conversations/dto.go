package conversations

type GetOrCreateDirectConvType struct {
	UserIDOne int64 `json:"user_id_one" validate:"required,gte=1"`
	UserIDTwo int64 `json:"user_id_two" validate:"required,gte=1"`
}

type GetConvParticipantType struct {
	ConvID          int64  `json:"conv_id"`
	UserID          int64  `json:"user_id"`
	Name            string `json:"name"`
	Username        string `json:"username"`
	LastMessage     string `json:"last_message,omitempty"`
	LastMessageTime string `json:"last_message_time,omitempty"`
	UnreadCount     int    `json:"unread_count"`
}
type CreateMessageType struct {
	SenderID int64
	ConvID   int64
	Content  string
}

type GetConvChatResponseType struct {
	ID        int64  `json:"id"`
	SenderID  int64  `json:"sender_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type CreateMessageResponseType struct {
	ID          int64  `json:"id"`
	ConvID      int64  `json:"conv_id"`
	SenderID    int64  `json:"sender_id"`
	RecipientID int64  `json:"recipient_id"`
	Content     string `json:"content"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}
