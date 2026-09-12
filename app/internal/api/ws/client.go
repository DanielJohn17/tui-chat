package ws

import (
	"encoding/json"
	"log"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024
)

// temporary counter
var msgIDCounter int64 = 1000

type Client struct {
	UserID   int64
	Username string
	ConvID   int64
	Send     chan []byte
	Conn     *websocket.Conn
	Hub      *Hub
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

		now := time.Now().Format(time.RFC3339)
		msgID := atomic.AddInt64(&msgIDCounter, 1)

		outbound := WSMessage{
			ID:          msgID,
			SenderID:    c.UserID,
			RecipientID: inboundMessage.RecipientID,
			ConvID:      convID,
			Content:     inboundMessage.Content,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		encoding, err := json.Marshal(outbound)
		if err != nil {
			log.Printf("ws marshal error: %v", err)
			continue
		}

		c.Hub.SendDirect <- &DirectMessage{
			SenderID:    c.UserID,
			RecipientID: inboundMessage.RecipientID,
			ConvID:      convID,
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
