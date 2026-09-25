package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"
)

type ConvServiceInt interface {
	GetContactUserIDs(ctx context.Context, userID int64) ([]int64, error)
}

type DirectMessage struct {
	SenderID       int64
	SenderUsername string
	RecipientID    int64
	ConvID         int64
	Message        WSMessage
	UnreadCount    int
}

type Hub struct {
	mu sync.RWMutex
	// support multiple client per user
	users map[int64]map[*Client]bool

	SendDirect    chan *DirectMessage
	BroadcastRead chan *ConversationReadPayload

	Register   chan *Client
	UnRegister chan *Client

	convService ConvServiceInt
}

func NewHub(convService ConvServiceInt) *Hub {
	return &Hub{
		users:         make(map[int64]map[*Client]bool),
		SendDirect:    make(chan *DirectMessage, 256),
		BroadcastRead: make(chan *ConversationReadPayload, 256),
		Register:      make(chan *Client, 32),
		UnRegister:    make(chan *Client, 32),
		convService:   convService,
	}
}

func (h *Hub) GetOnlineStatus(ctx context.Context, userIDs []int64) (map[int64]bool, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make(map[int64]bool, len(userIDs))
	for _, uid := range userIDs {
		result[uid] = len(h.users[uid]) > 0
	}
	return result, nil
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			// Register client to users map
			h.mu.Lock()
			wasOnline := len(h.users[client.UserID]) > 0
			if _, ok := h.users[client.UserID]; !ok {
				h.users[client.UserID] = make(map[*Client]bool)
			}
			h.users[client.UserID][client] = true
			h.mu.Unlock()

			// Fetch conversation peer IDs in background
			go func(c *Client, isNewOnline bool) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				peerIDs, err := h.convService.GetContactUserIDs(ctx, c.UserID)
				if err != nil {
					slog.WarnContext(ctx, "failed to get contact user IDs on register", "UserID", c.UserID, "error", err)
					return
				}

				h.sendPresenceSnapshot(c, peerIDs)

				if isNewOnline {
					h.notifyPeers(peerIDs, c.UserID, true)
				}
			}(client, !wasOnline)

		case client := <-h.UnRegister:
			// Unregister clent from users map
			h.mu.Lock()
			becameOffline := false
			if clients, ok := h.users[client.UserID]; ok {
				if _, exists := clients[client]; exists {
					delete(clients, client)
					if len(clients) == 0 {
						delete(h.users, client.UserID)
						becameOffline = true
					}
				}
			}
			h.mu.Unlock()

			if becameOffline {
				go func(userID int64) {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()

					peerIDs, err := h.convService.GetContactUserIDs(ctx, userID)
					if err != nil {
						slog.WarnContext(ctx, err.Error(), "UserID", userID, "error", err)
						return
					}

					h.notifyPeers(peerIDs, userID, false)
				}(client.UserID)
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
							go func(msgID, userID, convID int64, service MessagePersister) {
								ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
								defer cancel()

								service.MarkAsRead(ctx, msgID, userID, convID)
							}(msg.Message.ID, c.UserID, msg.ConvID, c.convService)
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

func (h *Hub) sendPresenceSnapshot(c *Client, peerIDs []int64) {
	h.mu.RLock()
	var onlinePeerIDs []int64
	for _, peerID := range peerIDs {
		if len(h.users[peerID]) > 0 {
			onlinePeerIDs = append(onlinePeerIDs, peerID)
		}
	}
	h.mu.RUnlock()

	frame, err := json.Marshal(WSNotification{
		Type: TypePresenceSnapshot,
		Payload: PresenceSnapshotPayload{
			OnlineUserIDs: onlinePeerIDs,
		},
	})
	if err != nil {
		slog.Error("failed to create presence snapshot frame to be sent", "error", err)
		return
	}

	select {
	case c.Send <- frame:
	default:
		h.dropClient(c)
	}
}

func (h *Hub) notifyPeers(peerIDs []int64, userID int64, online bool) {
	payload := UserPresencePayload{
		UserID: userID,
		Online: online,
	}
	frame, err := json.Marshal(WSNotification{
		Type:    TypeUserPresence,
		Payload: payload,
	})
	if err != nil {
		slog.Error("failed to create user presence frame to be sent", "error", err)
		return
	}

	h.mu.RLock()
	var targetClients []*Client
	for _, peerID := range peerIDs {
		// Fast map lookup O(1) per conversation partner
		if clients, isOnline := h.users[peerID]; isOnline {
			for c := range clients {
				targetClients = append(targetClients, c)
			}
		}
	}
	h.mu.RUnlock()

	// Non-blocking dispatch
	for _, c := range targetClients {
		select {
		case c.Send <- frame:
		default:
			h.dropClient(c)
		}
	}
}

func (h *Hub) dropClient(c *Client) {
	h.mu.Lock()
	becameOffline := false
	if clients, ok := h.users[c.UserID]; ok {
		if _, exists := clients[c]; exists {
			delete(clients, c)
			if len(clients) == 0 {
				delete(h.users, c.UserID)
				becameOffline = true
			}
		}
	}
	h.mu.Unlock()

	c.closeSendChannel()

	if becameOffline {
		go func(userID int64) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			peerIDs, err := h.convService.GetContactUserIDs(ctx, userID)
			if err != nil {
				slog.WarnContext(ctx, err.Error(), "UserID", userID, "error", err)
				return
			}

			h.notifyPeers(peerIDs, userID, false)
		}(c.UserID)
	}
}
