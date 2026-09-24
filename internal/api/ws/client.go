package ws

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/DanielJohn17/tui-chat/internal/api/conversations"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024
)

type MessagePersister interface {
	CreateMessage(
		ctx context.Context,
		convID, senderID int64,
		content string,
	) (*conversations.CreateMessageResponseType, error)

	GetUnreadCount(
		ctx context.Context,
		userID, convID int64,
	) (int, error)

	MarkAsRead(ctx context.Context, messageID, userID, convID int64)
}

type Client struct {
	UserID       int64
	Username     string
	ActiveConvID atomic.Int64
	Send         chan []byte
	Conn         *websocket.Conn
	Hub          *Hub
	convService  MessagePersister
	closeSend    sync.Once
}

func (c *Client) closeSendChannel() {
	c.closeSend.Do(func() {
		if c.Send != nil {
			close(c.Send)
		}
	})
}

func (c *Client) readPump() {
	defer func() {
		c.Hub.UnRegister <- c
		_ = c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	if err := c.Conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		return
	}
	c.Conn.SetPongHandler(func(string) error {
		return c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(
				err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
			) {
				log.Printf("ws read error: %v", err)
			}
			break
		}

		var inboundMessage InboundMessage
		if err := json.Unmarshal(message, &inboundMessage); err != nil {
			slog.Error("invalid inboundmessage", "error", err)
			continue
		}

		switch inboundMessage.Action {
		case ActionFocusConv:
			var payload FocusConvPayload
			if err := json.Unmarshal(inboundMessage.Payload, &payload); err != nil {
				slog.Error("invalid focus_conv", "error", err)
				continue
			}

			c.ActiveConvID.Store(payload.ConvID)
			if payload.ConvID > 0 {
				c.triggerReadReceipt(MarkReadPayload{
					ConvID:    payload.ConvID,
					MessageID: 0,
				})
			}
			continue

		case ActionSendMessage:
			c.handleSendMessage(inboundMessage.Payload)
			continue

		case ActionMarkRead:
			var payload MarkReadPayload
			if err := json.Unmarshal(inboundMessage.Payload, &payload); err != nil {
				slog.Error("invalid mark_read", "error", err)
				continue
			}

			if payload.ConvID > 0 {
				c.triggerReadReceipt(payload)
			}
			continue
		}

	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if err := c.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				return
			}

			if !ok {
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			if err := c.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				return
			}
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) sendError(convID int64, content, errMsg string, retryable bool) {
	errPayload, err := json.Marshal(WSMessageError{
		Type:      "error",
		ConvID:    convID,
		Content:   content,
		Error:     errMsg,
		Retryable: retryable,
	})
	if err != nil {
		return
	}

	select {
	case c.Send <- errPayload:
	default:
		log.Printf("ws send buffer full for user %d, dropped error frame", c.UserID)
	}
}

func (c *Client) triggerReadReceipt(payload MarkReadPayload) {
	// Persist to DB asynchronously so not to block the websocket read loop
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		c.convService.MarkAsRead(ctx, payload.MessageID, c.UserID, payload.ConvID)
	}()

	// hub broadcast
	c.Hub.BroadcastRead <- &ConversationReadPayload{
		ConvID:    payload.ConvID,
		UserID:    c.UserID,
		MessageID: payload.MessageID,
	}
}

func (c *Client) handleSendMessage(rawPayload json.RawMessage) {
	var payload SendMessagePayload
	if err := json.Unmarshal(rawPayload, &payload); err != nil {
		slog.Error("invalid send_message", "error", err)
		return
	}

	if payload.ConvID <= 0 || strings.TrimSpace(payload.Content) == "" {
		c.sendError(payload.ConvID, payload.Content, "invalid conversation id or empty message", false)
		return
	}

	convID := payload.ConvID

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	saved, err := c.convService.CreateMessage(
		ctx,
		convID,
		c.UserID,
		payload.Content,
	)
	if err != nil {
		slog.Error("failed to persist message from user", "ID", c.UserID, "error", err)
		c.sendError(
			convID,
			payload.Content,
			"Failed to save message. Please retry.",
			true,
		)
		return
	}

	unreadCount, err := c.convService.GetUnreadCount(
		ctx,
		saved.RecipientID,
		saved.ConvID,
	)
	if err != nil {
		slog.Error("failed fetching unread count", "error", err)
		unreadCount = 1
	}

	outbound := WSMessage{
		ID:          saved.ID,
		SenderID:    saved.SenderID,
		RecipientID: saved.RecipientID,
		ConvID:      saved.ConvID,
		Content:     saved.Content,
		CreatedAt:   saved.CreatedAt,
		UpdatedAt:   saved.UpdatedAt,
	}

	c.Hub.SendDirect <- &DirectMessage{
		SenderUsername: c.Username,
		SenderID:       saved.SenderID,
		RecipientID:    saved.RecipientID,
		ConvID:         saved.ConvID,
		Message:        outbound,
		UnreadCount:    unreadCount,
	}
}
