package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/DanielJohn17/tui-chat/internal/tui/crypto"
)

type convChatData struct {
	ID        int64  `json:"id"`
	SenderID  int64  `json:"sender_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type bulkChatData struct {
	ID        int64  `json:"id"`
	SenderID  int64  `json:"sender_id"`
	ConvID    int64  `json:"conv_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (h *HTTPClient) FetchMessages(chatID int64) ([]Message, error) {
	h.mu.RLock()
	epoch := h.sessionEpoch
	currentUserID := h.profile.ID
	h.mu.RUnlock()

	path := fmt.Sprintf("/api/v1/conversations/%d/chats?limit=50", chatID)
	req, err := h.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, h.formatNetworkError(err)
	}
	defer resp.Body.Close()

	bodyBytes, err := readLimitedBody(resp.Body, maxHTTPPayloadBytes)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		var errResp apiResponse[any]
		if json.Unmarshal(bodyBytes, &errResp) == nil && errResp.Error != "" {
			return nil, errors.New(errResp.Error)
		}
		return nil, fmt.Errorf("Failed to fetch messages (HTTP %d)", resp.StatusCode)
	}

	var res apiResponse[[]convChatData]
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return nil, errors.New("Invalid server response format")
	}

	var recipientName string
	var recipientID int64
	var isSecret bool

	h.mu.RLock()
	isSecret = h.convIsSecret[chatID]
	h.mu.RUnlock()

	if ch, ok := h.FindChatByConvID(chatID); ok {
		recipientName = ch.Name
		recipientID = ch.RecipientID
		if recipientName == "" {
			recipientName = ch.Username
		}
		if !isSecret && ch.SecretID == chatID {
			isSecret = true
		}
	}

	if recipientID == 0 {
		h.mu.RLock()
		recipientID = h.convPeerID[chatID]
		h.mu.RUnlock()
	}

	n := len(res.Data)
	messages := make([]Message, n)
	for i, c := range res.Data {
		isSelf := c.SenderID == currentUserID
		senderName := recipientName
		if isSelf {
			senderName = "You"
		}

		text := c.Content
		if isSecret {
			text = h.DecryptMessage(recipientID, text)
		}

		if len(text) > maxRenderTextLength {
			text = text[:maxRenderTextLength] + "… [truncated]"
		}
		// API returns chats in DESC order (newest first).
		// Store them chronologically (oldest at index 0, newest at bottom index n-1)
		messages[n-1-i] = Message{
			ID:        c.ID,
			SenderID:  c.SenderID,
			Sender:    senderName,
			ConvID:    chatID,
			Text:      text,
			Timestamp: formatFriendlyTime(c.CreatedAt),
			Self:      isSelf,
			IsSecret:  isSecret,
		}
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	if h.sessionEpoch != epoch || h.profile.Token == "" {
		return nil, errors.New("session invalidated")
	}

	existing := h.messages[chatID]
	if len(existing) == 0 {
		h.messages[chatID] = messages
	} else {
		seen := make(map[int64]bool, len(messages))
		for _, m := range messages {
			if m.ID > 0 {
				seen[m.ID] = true
			}
		}
		merged := append([]Message(nil), messages...)
		for _, m := range existing {
			if m.ID == 0 || !seen[m.ID] {
				merged = append(merged, m)
			}
		}
		h.messages[chatID] = merged
	}

	return h.messages[chatID], nil
}

func (h *HTTPClient) FetchBulkMessages() (map[int64][]Message, error) {
	h.mu.RLock()
	epoch := h.sessionEpoch
	currentUserID := h.profile.ID
	chatsMap := make(map[int64]string)
	recipientMap := make(map[int64]int64)
	for _, ch := range h.chats {
		name := ch.Name
		if name == "" {
			name = ch.Username
		}
		chatsMap[ch.ID] = name
		recipientMap[ch.ID] = ch.RecipientID
		if ch.CloudID != 0 {
			chatsMap[ch.CloudID] = name
			recipientMap[ch.CloudID] = ch.RecipientID
		}
		if ch.SecretID != 0 {
			chatsMap[ch.SecretID] = name
			recipientMap[ch.SecretID] = ch.RecipientID
		}
	}
	h.mu.RUnlock()

	// Pre-derive symmetric keys for all Secret conversations
	symKeys := make(map[int64][32]byte) // map convID -> key
	if idKey := h.IdentityKey(); idKey != nil {
		for convID, peerID := range recipientMap {
			if h.convIsSecret[convID] {
				if peerPub, err := h.GetPeerPublicKey(peerID); err == nil {
					var peerArr [32]byte
					copy(peerArr[:], peerPub)
					if key, err := crypto.DeriveConvKey(idKey.PrivateKey, peerArr); err == nil {
						symKeys[convID] = key
					}
				}
			}
		}
	}

	path := "/api/v1/conversations/bulk"
	req, err := h.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, h.formatNetworkError(err)
	}
	defer resp.Body.Close()

	bodyBytes, err := readLimitedBody(resp.Body, maxHTTPBulkPayloadBytes)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		var errResp apiResponse[any]
		if json.Unmarshal(bodyBytes, &errResp) == nil && errResp.Error != "" {
			return nil, errors.New(errResp.Error)
		}
		return nil, fmt.Errorf("Failed to fetch bulk messages (HTTP %d)", resp.StatusCode)
	}

	var res apiResponse[[]bulkChatData]
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return nil, errors.New("Invalid server response format")
	}

	grouped := make(map[int64][]Message)
	for _, c := range res.Data {
		isSelf := c.SenderID == currentUserID
		senderName := chatsMap[c.ConvID]
		if senderName == "" {
			senderName = "Recipient"
		}
		if isSelf {
			senderName = "You"
		}

		text := c.Content
		isSecret := h.convIsSecret[c.ConvID]
		if isSecret {
			peerID := recipientMap[c.ConvID]
			if peerID == 0 {
				peerID = h.convPeerID[c.ConvID]
			}
			text = h.DecryptMessage(peerID, text)
		}

		if len(text) > maxRenderTextLength {
			text = text[:maxRenderTextLength] + "… [truncated]"
		}

		msg := Message{
			ID:        c.ID,
			SenderID:  c.SenderID,
			Sender:    senderName,
			ConvID:    c.ConvID,
			Text:      text,
			Timestamp: formatFriendlyTime(c.CreatedAt),
			Self:      isSelf,
			IsSecret:  isSecret,
		}
		// prepend to reverse chronological API order to oldest-first
		grouped[c.ConvID] = append([]Message{msg}, grouped[c.ConvID]...)
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	if h.sessionEpoch != epoch || h.profile.Token == "" {
		return nil, errors.New("session invalidated")
	}

	for convID, msgs := range grouped {
		existing := h.messages[convID]
		if len(existing) == 0 {
			h.messages[convID] = msgs
			continue
		}
		seen := make(map[int64]bool, len(msgs))
		for _, m := range msgs {
			if m.ID > 0 {
				seen[m.ID] = true
			}
		}
		merged := append([]Message(nil), msgs...)
		for _, m := range existing {
			if m.ID == 0 || !seen[m.ID] {
				merged = append(merged, m)
			}
		}
		h.messages[convID] = merged
	}

	return h.messages, nil
}

func (h *HTTPClient) Messages(chatID int64) []Message {
	h.mu.RLock()
	defer h.mu.RUnlock()
	msgs := h.messages[chatID]
	out := make([]Message, len(msgs))
	copy(out, msgs)
	return out
}

func (h *HTTPClient) AppendMessage(msg Message) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.messages[msg.ConvID] = append(h.messages[msg.ConvID], msg)
}
