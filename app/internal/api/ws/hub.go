package ws

import (
	"context"
	"encoding/json"
	"log/slog"
)

type DirectMessage struct {
	SenderID       int64
	SenderUsername string
	RecipientID    int64
	ConvID         int64
	Message        WSMessage
	UnreadCount    int
}

type Hub struct {
	// support multiple client per user
	users map[int64]map[*Client]bool

	// set active clients in conversation
	convs map[int64]map[*Client]bool

	SendDirect    chan *DirectMessage
	BroadcastRead chan *ConversationReadPayload

	Register   chan *Client
	UnRegister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		users:         make(map[int64]map[*Client]bool),
		convs:         make(map[int64]map[*Client]bool),
		SendDirect:    make(chan *DirectMessage, 256),
		BroadcastRead: make(chan *ConversationReadPayload, 256),
		Register:      make(chan *Client, 32),
		UnRegister:    make(chan *Client, 32),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			// Register client to users map
			if _, ok := h.users[client.UserID]; !ok {
				h.users[client.UserID] = make(map[*Client]bool)
			}
			h.users[client.UserID][client] = true

			// Register client to convs map if convID is set
			if client.ConvID > 0 {
				if _, ok := h.convs[client.ConvID]; !ok {
					h.convs[client.ConvID] = make(map[*Client]bool)
				}
				h.convs[client.ConvID][client] = true
			}

		case client := <-h.UnRegister:
			// Unregister clent from users map
			if clients, ok := h.users[client.UserID]; ok {
				if _, exists := clients[client]; exists {
					delete(clients, client)
					close(client.Send)
					if len(clients) == 0 {
						delete(h.users, client.UserID)
					}
				}
			}

			// Unregister client from convs map
			if client.ConvID > 0 {
				if clients, ok := h.convs[client.ConvID]; ok {
					if _, exists := clients[client]; exists {
						delete(clients, client)
						if len(clients) == 0 {
							delete(h.convs, client.ConvID)
						}
					}
				}
			}

		case msg := <-h.SendDirect:
			// Deliver direct message to recipient
			if msg.RecipientID <= 0 {
				continue
			}

			chatMsgFrame, err := json.Marshal(WSNotification{
				Type:    TypeChatMessage,
				Payload: msg.Message,
			})
			if err != nil {
				continue
			}

			notificationFrame, err := json.Marshal(WSNotification{
				Type: TypeChatNotification,
				Payload: ChatNotificationPayload{
					ConvID:      msg.ConvID,
					SenderID:    msg.SenderID,
					SenderName:  msg.SenderUsername,
					Content:     msg.Message.Content,
					UnreadCount: msg.UnreadCount,
					CreatedAt:   msg.Message.CreatedAt,
				},
			})
			if err != nil {
				continue
			}

			if recipientClients, isOnline := h.users[msg.RecipientID]; isOnline {
				for c := range recipientClients {
					activeConv := c.ActiveConvID.Load()

					if activeConv < 0 {
						h.dropClient(c)
						continue
					}

					if activeConv == msg.ConvID {
						// User is in the current looking conversation
						select {
						case c.Send <- chatMsgFrame:
							// Auto-mark as read in background since recipient is actively looking
							go c.convService.MarkAsRead(
								context.Background(),
								msg.Message.ID,
								c.UserID,
								msg.ConvID,
							)
						default:
							h.dropClient(c)
						}
					} else {
						// Recipient is in another conversation or in the sidebar
						select {
						case c.Send <- notificationFrame:
						default:
							h.dropClient(c)
						}
					}

				}
			}

			// Echo back to the sender's client(s) for delivery confirmation
			if msg.SenderID != msg.RecipientID {
				if senderClients, ok := h.users[msg.SenderID]; ok {
					for c := range senderClients {
						select {
						case c.Send <- chatMsgFrame:
						default:
							h.dropClient(c)
						}
					}
				}
			}
		case readEvt := <-h.BroadcastRead:
			frame, err := json.Marshal(WSNotification{
				Type:    TypeConversationRead,
				Payload: readEvt,
			})
			if err != nil {
				slog.Error("failed to create read event frame to be sent", "error", err)
				continue
			}

			if readClients, ok := h.users[readEvt.UserID]; ok {
				for c := range readClients {
					select {
					case c.Send <- frame:
					default:
						h.dropClient(c)
					}
				}
			}
		}

	}
}

func (h *Hub) dropClient(c *Client) {
	if clients, ok := h.users[c.UserID]; ok {
		if _, exists := clients[c]; exists {
			delete(clients, c)
			if len(clients) == 0 {
				delete(h.users, c.UserID)
			}
		}
	}

	if c.ConvID > 0 {
		if clients, ok := h.convs[c.ConvID]; ok {
			delete(clients, c)
			if len(clients) == 0 {
				delete(h.convs, c.ConvID)
			}
		}
	}

	close(c.Send)
}
