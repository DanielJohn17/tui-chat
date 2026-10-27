package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/DanielJohn17/tui-chat/internal/tui/crypto"
	"github.com/gorilla/websocket"
)

type wsInboundEnvelope struct {
	Action  string `json:"action"`
	Payload any    `json:"payload"`
}

type wsOutboundEnvelope struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

func (h *HTTPClient) ConnectWS(eventsChan chan<- any) error {
	h.wsMu.Lock()
	defer h.wsMu.Unlock()

	if h.wsConnected.Load() && h.wsConn != nil {
		return nil
	}
	if h.wsConn != nil {
		_ = h.wsConn.Close()
		h.wsConn = nil
	}

	h.mu.RLock()
	token := h.profile.Token
	h.mu.RUnlock()

	if token == "" {
		return errors.New("Cannot connect websocket: unauthenticated")
	}

	u, err := url.Parse(h.baseURL)
	if err != nil {
		return errors.New("Invalid chat server URL")
	}

	wsScheme := "ws"
	if u.Scheme == "https" {
		wsScheme = "wss"
	}
	wsURL := fmt.Sprintf("%s://%s/api/v1/ws", wsScheme, u.Host)

	header := http.Header{}
	header.Set("Authorization", "Bearer "+token)
	header.Set(ClientHeaderApp, ClientAppName)
	header.Set("User-Agent", "tui-chat/1.0")
	if h.appSecret != "" {
		header.Set(ClientHeaderSecret, h.appSecret)
	}

	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}

	conn, resp, err := dialer.Dial(wsURL, header)
	if err != nil {
		if resp != nil {
			if resp.StatusCode == http.StatusForbidden {
				return errors.New("Client verification failed: unauthorized app secret")
			}
			if resp.StatusCode == http.StatusUnauthorized {
				return errors.New("Websocket authentication failed: unauthorized")
			}
		}
		return errors.New("Chat server is currently offline or unreachable")
	}

	conn.SetReadLimit(maxWSPayloadBytes)

	connDone := make(chan struct{})
	h.wsDoneChan = connDone
	h.wsConn = conn
	h.wsConnected.Store(true)
	h.wsIntentionalClose.Store(false)

	// Launch reader pump
	go func() {
		defer func() {
			h.wsMu.Lock()
			wasCurrent := (h.wsConn == conn)
			if wasCurrent {
				h.wsConnected.Store(false)
				h.wsConn = nil
				if h.wsDoneChan == connDone {
					h.wsDoneChan = nil
				}
			}
			isIntentional := h.wsIntentionalClose.Swap(false)
			h.wsMu.Unlock()

			select {
			case <-connDone:
			default:
				close(connDone)
			}
			_ = conn.Close()

			if wasCurrent && !isIntentional {
				select {
				case eventsChan <- WSErrorPayload{
					Type:  "disconnected",
					Error: "websocket connection lost",
				}:
				default:
				}
			}
		}()

		sendEvent := func(evt any) {
			select {
			case eventsChan <- evt:
			case <-connDone:
			default:
			}
		}

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var env wsOutboundEnvelope
			if err := json.Unmarshal(message, &env); err != nil {
				continue
			}

			switch env.Type {
			case "chat_message":
				var p WSChatMessagePayload
				if err := json.Unmarshal(env.Payload, &p); err == nil {
					if len(p.Content) > maxRenderTextLength {
						p.Content = p.Content[:maxRenderTextLength] + "… [truncated]"
					}
					sendEvent(p)
				}
			case "chat_notification":
				var p WSChatNotificationPayload
				if err := json.Unmarshal(env.Payload, &p); err == nil {
					if len(p.Content) > maxRenderTextLength {
						p.Content = p.Content[:maxRenderTextLength] + "… [truncated]"
					}
					sendEvent(p)
				}
			case "conversation_read":
				var p WSConversationReadPayload
				if err := json.Unmarshal(env.Payload, &p); err == nil {
					sendEvent(p)
				}
			case "user_presence":
				var p WSUserPresencePayload
				if err := json.Unmarshal(env.Payload, &p); err == nil {
					sendEvent(p)
				}
			case "presence_snapshot":
				var p WSPresenceSnapshotPayload
				if err := json.Unmarshal(env.Payload, &p); err == nil {
					sendEvent(p)
				}
			case "error":
				var p WSErrorPayload
				if err := json.Unmarshal(env.Payload, &p); err == nil {
					sendEvent(p)
				}
			}
		}
	}()

	return nil
}

func (h *HTTPClient) CloseWS() error {
	h.wsMu.Lock()
	defer h.wsMu.Unlock()

	h.wsIntentionalClose.Store(true)
	h.wsConnected.Store(false)
	if h.wsDoneChan != nil {
		select {
		case <-h.wsDoneChan:
		default:
			close(h.wsDoneChan)
		}
		h.wsDoneChan = nil
	}
	if h.wsConn != nil {
		err := h.wsConn.Close()
		h.wsConn = nil
		return err
	}
	return nil
}

func (h *HTTPClient) IsWSConnected() bool {
	return h.wsConnected.Load()
}

func (h *HTTPClient) SendWS(convID int64, text string) error {
	h.mu.RLock()
	isSecret := h.convIsSecret[convID]
	h.mu.RUnlock()

	ch, ok := h.FindChatByConvID(convID)
	if !ok || ch.RecipientID == 0 {
		return errors.New("chat not found")
	}
	recipientID := ch.RecipientID

	content := text

	if isSecret {
		// Fetch peer public key
		peerPub, err := h.GetPeerPublicKey(recipientID)
		if err != nil {
			return fmt.Errorf("failed to get peer public key: %w", err)
		}

		// Derive symmetric key
		var peerPubArr [32]byte
		copy(peerPubArr[:], peerPub)

		idKey := h.IdentityKey()
		if idKey == nil {
			return errors.New("identity key not loaded")
		}

		symKey, err := crypto.DeriveConvKey(idKey.PrivateKey, peerPubArr)
		if err != nil {
			return fmt.Errorf("failed to derive key: %w", err)
		}

		// Encrypt
		encrypted, err := crypto.Encrypt(symKey, text)
		if err != nil {
			return fmt.Errorf("encryption failed: %w", err)
		}
		content = encrypted
	}

	h.wsMu.Lock()
	defer h.wsMu.Unlock()

	if !h.wsConnected.Load() || h.wsConn == nil {
		return errors.New("websocket is not connected")
	}

	payload, _ := json.Marshal(map[string]any{
		"conv_id": convID,
		"content": content,
	})

	inbound := wsInboundEnvelope{
		Action:  "send_message",
		Payload: json.RawMessage(payload),
	}

	data, err := json.Marshal(inbound)
	if err != nil {
		return err
	}

	_ = h.wsConn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	return h.wsConn.WriteMessage(websocket.TextMessage, data)
}

func (h *HTTPClient) FocusConv(convID int64) error {
	h.wsMu.Lock()
	defer h.wsMu.Unlock()

	if !h.wsConnected.Load() || h.wsConn == nil {
		return nil
	}

	payload, _ := json.Marshal(map[string]any{
		"conv_id": convID,
	})

	inbound := wsInboundEnvelope{
		Action:  "focus_conv",
		Payload: json.RawMessage(payload),
	}

	data, err := json.Marshal(inbound)
	if err != nil {
		return err
	}

	_ = h.wsConn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	return h.wsConn.WriteMessage(websocket.TextMessage, data)
}

func (h *HTTPClient) MarkRead(convID int64, messageID int64) error {
	h.wsMu.Lock()
	defer h.wsMu.Unlock()

	if !h.wsConnected.Load() || h.wsConn == nil {
		return nil
	}

	payload, _ := json.Marshal(map[string]any{
		"conv_id":    convID,
		"message_id": messageID,
	})

	inbound := wsInboundEnvelope{
		Action:  "mark_read",
		Payload: json.RawMessage(payload),
	}

	data, err := json.Marshal(inbound)
	if err != nil {
		return err
	}

	_ = h.wsConn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	return h.wsConn.WriteMessage(websocket.TextMessage, data)
}
