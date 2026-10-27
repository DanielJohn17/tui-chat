package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type convParticipantData struct {
	ConvID          int64  `json:"conv_id"`
	IsSecret        bool   `json:"is_secret"`
	UserID          int64  `json:"user_id"`
	Name            string `json:"name"`
	Username        string `json:"username"`
	Bio             string `json:"bio"`
	LastMessage     string `json:"last_message"`
	LastMessageTime string `json:"last_message_time"`
	UnreadCount     int    `json:"unread_count"`
}

func formatFriendlyTime(tStr string) string {
	if tStr == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339, tStr)
	if err != nil {
		t, err = time.Parse(time.RFC3339Nano, tStr)
		if err != nil {
			return tStr
		}
	}

	now := time.Now()
	if t.Year() == now.Year() && t.YearDay() == now.YearDay() {
		return t.Format("15:04")
	}
	if t.Year() == now.Year() {
		return t.Format("Jan 02")
	}
	return t.Format("02/01/06")
}

func (h *HTTPClient) IsConvSecret(convID int64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.convIsSecret[convID]
}

func (h *HTTPClient) SetConvSecret(convID int64, isSecret bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.convIsSecret == nil {
		h.convIsSecret = make(map[int64]bool)
	}
	h.convIsSecret[convID] = isSecret
}

func (h *HTTPClient) FetchChats() ([]Chat, error) {
	h.mu.RLock()
	epoch := h.sessionEpoch
	h.mu.RUnlock()

	req, err := h.newRequest(http.MethodGet, "/api/v1/conversations", nil)
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
		return nil, fmt.Errorf("Failed to fetch conversations (HTTP %d)", resp.StatusCode)
	}

	var res apiResponse[[]convParticipantData]
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return nil, errors.New("Invalid server response format")
	}

	h.mu.RLock()
	if h.sessionEpoch != epoch || h.profile.Token == "" {
		h.mu.RUnlock()
		return nil, errors.New("session invalidated")
	}
	onlineMap := make(map[int64]bool, len(h.onlineUsers))
	for k, v := range h.onlineUsers {
		onlineMap[k] = v
	}
	h.mu.RUnlock()

	tempIsSecret := make(map[int64]bool)
	tempPeerID := make(map[int64]int64)
	contactIndex := make(map[int64]int)
	type chatMeta struct {
		rawTime string
	}
	var chatMetas []chatMeta
	var unifiedChats []Chat

	for _, c := range res.Data {
		tempIsSecret[c.ConvID] = c.IsSecret
		tempPeerID[c.ConvID] = c.UserID

		lastMsg := c.LastMessage
		if c.IsSecret && c.UserID != 0 && lastMsg != "" {
			lastMsg = h.DecryptMessage(c.UserID, lastMsg)
		}

		idx, exists := contactIndex[c.UserID]
		if !exists {
			ch := Chat{
				ID:          c.ConvID,
				RecipientID: c.UserID,
				Name:        c.Name,
				Username:    c.Username,
				Bio:         c.Bio,
				LastMessage: lastMsg,
				Time:        formatFriendlyTime(c.LastMessageTime),
				Unread:      c.UnreadCount,
				Online:      onlineMap[c.UserID],
				IsSecret:    c.IsSecret,
			}
			if c.IsSecret {
				ch.SecretID = c.ConvID
			} else {
				ch.CloudID = c.ConvID
			}
			contactIndex[c.UserID] = len(unifiedChats)
			unifiedChats = append(unifiedChats, ch)
			chatMetas = append(chatMetas, chatMeta{rawTime: c.LastMessageTime})
		} else {
			existing := &unifiedChats[idx]
			if c.IsSecret {
				existing.SecretID = c.ConvID
			} else {
				existing.CloudID = c.ConvID
			}
			existing.Unread += c.UnreadCount
			// If this conversation has the newer message, use it
			if c.LastMessageTime >= chatMetas[idx].rawTime || existing.LastMessage == "" {
				existing.LastMessage = lastMsg
				existing.IsSecret = c.IsSecret
				existing.Time = formatFriendlyTime(c.LastMessageTime)
				chatMetas[idx].rawTime = c.LastMessageTime
			}
			if existing.CloudID != 0 {
				existing.ID = existing.CloudID
			}
		}
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	if h.sessionEpoch != epoch || h.profile.Token == "" {
		return nil, errors.New("session invalidated")
	}

	if h.convIsSecret == nil {
		h.convIsSecret = make(map[int64]bool)
	}
	if h.convPeerID == nil {
		h.convPeerID = make(map[int64]int64)
	}
	for k, v := range tempIsSecret {
		h.convIsSecret[k] = v
	}
	for k, v := range tempPeerID {
		h.convPeerID[k] = v
	}
	h.chats = unifiedChats

	return unifiedChats, nil
}

func (h *HTTPClient) Chats() []Chat {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]Chat, len(h.chats))
	copy(out, h.chats)
	return out
}

func (h *HTTPClient) SetChats(chats []Chat) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for i := range chats {
		if isOnline, ok := h.onlineUsers[chats[i].RecipientID]; ok {
			chats[i].Online = isOnline
		}
	}
	h.chats = chats
}

func (h *HTTPClient) FindChatByConvID(convID int64) (Chat, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, ch := range h.chats {
		if ch.ID == convID || ch.CloudID == convID || ch.SecretID == convID {
			return ch, true
		}
	}
	return Chat{}, false
}

func (h *HTTPClient) SetUserOnline(userID int64, online bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.onlineUsers == nil {
		h.onlineUsers = make(map[int64]bool)
	}
	h.onlineUsers[userID] = online
	for i := range h.chats {
		if h.chats[i].RecipientID == userID {
			h.chats[i].Online = online
		}
	}
}

func (h *HTTPClient) SetOnlineUsers(userIDs []int64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.onlineUsers = make(map[int64]bool, len(userIDs))
	for _, id := range userIDs {
		h.onlineUsers[id] = true
	}
	for i := range h.chats {
		h.chats[i].Online = h.onlineUsers[h.chats[i].RecipientID]
	}
}

func (h *HTTPClient) UpdateChatSnippet(convID int64, lastMsg, timeStr string, unreadDelta int, setExactUnread *int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	isSecret := h.convIsSecret[convID]
	for i := range h.chats {
		if h.chats[i].ID == convID || h.chats[i].CloudID == convID || h.chats[i].SecretID == convID {
			if lastMsg != "" {
				h.chats[i].LastMessage = lastMsg
				h.chats[i].IsSecret = isSecret
			}
			if timeStr != "" {
				h.chats[i].Time = timeStr
			}
			if setExactUnread != nil {
				h.chats[i].Unread = *setExactUnread
			} else if unreadDelta != 0 {
				h.chats[i].Unread += unreadDelta
				if h.chats[i].Unread < 0 {
					h.chats[i].Unread = 0
				}
			}
			if lastMsg != "" && i > 0 {
				target := h.chats[i]
				copy(h.chats[1:i+1], h.chats[0:i])
				h.chats[0] = target
			}
			break
		}
	}
}
