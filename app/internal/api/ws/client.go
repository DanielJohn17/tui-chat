package ws

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/DanielJohn17/tui-chat/app/internal/api/conversations"
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
}

type Client struct {
	UserID      int64
	Username    string
	ConvID      int64
	Send        chan []byte
	Conn        *websocket.Conn
	Hub         *Hub
	convService MessagePersister
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
			inboundMessage.Content = string(message)
		}

		convID := c.ConvID
		if inboundMessage.ConvID > 0 {
			convID = inboundMessage.ConvID
		}

		if inboundMessage.Content == "" {
			continue
		}

		saved, err := c.convService.CreateMessage(
			context.Background(),
			convID,
			c.UserID,
			inboundMessage.Content,
		)
		if err != nil {
			log.Printf("failed to persist message from user %d: %v", c.UserID, err)
			c.sendError(
				convID,
				inboundMessage.Content,
				"Failed to save message. Please retry.",
				true,
			)
			continue
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

		encoding, err := json.Marshal(outbound)
		if err != nil {
			log.Printf("ws marshal error: %v", err)
			continue
		}

		c.Hub.SendDirect <- &DirectMessage{
			SenderID:    saved.SenderID,
			RecipientID: saved.RecipientID,
			ConvID:      saved.ConvID,
			Payload:     encoding,
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		c.Hub.UnRegister <- c
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

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			if _, err := w.Write(message); err != nil {
				_ = w.Close()
				return
			}

			// Add queued chat messages to the current websocket message
			n := len(c.Send)
			for range n {
				if _, err := w.Write([]byte{'\n'}); err != nil {
					_ = w.Close()
					return
				}

				if _, err := w.Write(<-c.Send); err != nil {
					_ = w.Close()
					return
				}
			}
			if err := w.Close(); err != nil {
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
