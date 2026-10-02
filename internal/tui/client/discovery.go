package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func (h *HTTPClient) SearchUsers(ctx context.Context, query string) ([]UserSummary, error) {
	var data struct {
		Users []UserSummary `json:"users"`
	}
	if err := h.discoveryRequest(ctx, http.MethodGet, "/api/v1/users/search?q="+url.QueryEscape(strings.TrimSpace(query)), nil, &data); err != nil {
		return nil, err
	}
	return data.Users, nil
}

func (h *HTTPClient) OpenDirectConversation(ctx context.Context, userID int64) (Chat, error) {
	self := h.Profile().ID
	if userID <= 0 || userID == self {
		return Chat{}, errors.New("Select another user to start a conversation")
	}
	body, err := json.Marshal(map[string]int64{"user_id_two": userID})
	if err != nil {
		return Chat{}, err
	}
	var participants []convParticipantData
	if err := h.discoveryRequest(ctx, http.MethodPost, "/api/v1/conversations", body, &participants); err != nil {
		return Chat{}, err
	}
	for _, p := range participants {
		if p.UserID == userID && p.ConvID > 0 {
			h.mu.RLock()
			online := h.onlineUsers[userID]
			h.mu.RUnlock()
			return Chat{ID: p.ConvID, RecipientID: p.UserID, Name: p.Name, Username: p.Username, LastMessage: p.LastMessage, Time: formatFriendlyTime(p.LastMessageTime), Unread: p.UnreadCount, Online: online}, nil
		}
	}
	return Chat{}, errors.New("Server returned an invalid conversation")
}

// Discovery requests return data only; the app applies results after checking request identity.
func (h *HTTPClient) discoveryRequest(ctx context.Context, method, path string, body []byte, out any) error {
	h.mu.RLock()
	epoch := h.sessionEpoch
	token := h.profile.Token
	userID := h.profile.ID
	h.mu.RUnlock()
	if token == "" {
		return errors.New("unauthenticated")
	}
	req, err := h.newRequest(method, path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := h.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return h.formatNetworkError(err)
	}
	defer resp.Body.Close()
	data, err := readLimitedBody(resp.Body, maxHTTPPayloadBytes)
	if err != nil {
		return err
	}
	var envelope apiResponse[json.RawMessage]
	if err := json.Unmarshal(data, &envelope); err != nil {
		return errors.New("Invalid server response format")
	}
	if !envelope.Success || (resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated) {
		if envelope.Error != "" {
			return errors.New(envelope.Error)
		}
		return fmt.Errorf("Request failed (HTTP %d)", resp.StatusCode)
	}
	h.mu.RLock()
	valid := h.sessionEpoch == epoch && h.profile.Token == token && h.profile.ID == userID
	h.mu.RUnlock()
	if !valid {
		return errors.New("session invalidated")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return json.Unmarshal(envelope.Data, out)
}
