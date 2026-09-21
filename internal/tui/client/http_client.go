package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

const (
	ClientHeaderApp = "X-Client-App"
	ClientAppName   = "tui-chat"
)

type HTTPClient struct {
	baseURL     string
	httpClient  *http.Client
	profile     Profile
	sessionPath string
	mu          sync.RWMutex
	chats       []Chat
	messages    map[int64][]Message
	wsConn      *websocket.Conn
	wsConnected atomic.Bool
	wsMu        sync.Mutex
}

func NewHTTPClient(baseURL string) Client {
	if baseURL == "" {
		baseURL = os.Getenv("API_URL")
		if baseURL == "" {
			baseURL = "http://localhost:8080"
		}
	}
	baseURL = strings.TrimRight(baseURL, "/")

	return &HTTPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 8 * time.Second,
		},
		messages: make(map[int64][]Message),
	}
}

// API Response Wrappers
type apiResponse[T any] struct {
	Success bool   `json:"success"`
	Data    T      `json:"data"`
	Error   string `json:"error,omitempty"`
}

type authData struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Username  string `json:"username"`
	Token     string `json:"token"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type convParticipantData struct {
	ConvID          int64  `json:"conv_id"`
	UserID          int64  `json:"user_id"`
	Name            string `json:"name"`
	Username        string `json:"username"`
	LastMessage     string `json:"last_message"`
	LastMessageTime string `json:"last_message_time"`
	UnreadCount     int    `json:"unread_count"`
}

type convChatData struct {
	ID        int64  `json:"id"`
	SenderID  int64  `json:"sender_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// WebSocket Envelope
type wsInboundEnvelope struct {
	Action  string `json:"action"`
	Payload any    `json:"payload"`
}

type wsOutboundEnvelope struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

func (h *HTTPClient) formatNetworkError(err error) error {
	if errors.Is(err, os.ErrDeadlineExceeded) || strings.Contains(err.Error(), "Client.Timeout") {
		return errors.New("Request timed out: chat server is taking too long to respond")
	}
	if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "no such host") || strings.Contains(err.Error(), "network is unreachable") {
		return errors.New("Chat server is currently offline or unreachable")
	}
	return errors.New("Unable to connect to chat server. Please check your internet connection.")
}

func (h *HTTPClient) newRequest(method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, h.baseURL+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(ClientHeaderApp, ClientAppName)
	req.Header.Set("User-Agent", "tui-chat/1.0")

	h.mu.RLock()
	token := h.profile.Token
	h.mu.RUnlock()
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return req, nil
}

func (h *HTTPClient) Login(username, password string) (*Profile, error) {
	reqBody, _ := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})

	req, err := h.newRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, h.formatNetworkError(err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("Failed to read server response")
	}

	var res apiResponse[authData]
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return nil, errors.New("Server returned an invalid response format")
	}

	if !res.Success || resp.StatusCode != http.StatusOK {
		if res.Error != "" {
			return nil, errors.New(res.Error)
		}
		return nil, fmt.Errorf("Login failed (HTTP %d)", resp.StatusCode)
	}

	h.mu.Lock()
	h.profile = Profile{
		ID:        res.Data.ID,
		Name:      res.Data.Name,
		Username:  res.Data.Username,
		Token:     res.Data.Token,
		CreatedAt: res.Data.CreatedAt,
	}
	sessPath := h.sessionPath
	profCopy := h.profile
	h.mu.Unlock()

	if sessPath != "" {
		_ = SaveSession(sessPath, profCopy)
	}

	return &h.profile, nil
}

func (h *HTTPClient) Register(name, username, password string) (*Profile, error) {
	reqBody, _ := json.Marshal(map[string]string{
		"name":     name,
		"username": username,
		"password": password,
	})

	req, err := h.newRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, h.formatNetworkError(err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("Failed to read server response")
	}

	var res apiResponse[authData]
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return nil, errors.New("Server returned an invalid response format")
	}

	if !res.Success || (resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated) {
		if res.Error != "" {
			return nil, errors.New(res.Error)
		}
		return nil, fmt.Errorf("Registration failed (HTTP %d)", resp.StatusCode)
	}

	h.mu.Lock()
	h.profile = Profile{
		ID:        res.Data.ID,
		Name:      res.Data.Name,
		Username:  res.Data.Username,
		Token:     res.Data.Token,
		CreatedAt: res.Data.CreatedAt,
	}
	sessPath := h.sessionPath
	profCopy := h.profile
	h.mu.Unlock()

	if sessPath != "" {
		_ = SaveSession(sessPath, profCopy)
	}

	return &h.profile, nil
}

func (h *HTTPClient) Logout() error {
	_ = h.CloseWS()
	h.mu.Lock()
	sessPath := h.sessionPath
	h.profile = Profile{}
	h.chats = nil
	h.messages = make(map[int64][]Message)
	h.mu.Unlock()

	if sessPath != "" {
		_ = ClearSession(sessPath)
	}
	return nil
}

func (h *HTTPClient) SessionPath() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.sessionPath
}

func (h *HTTPClient) SetSessionPath(path string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.sessionPath = path
}

func (h *HTTPClient) IsAuthenticated() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.profile.Token != ""
}

func (h *HTTPClient) Profile() Profile {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.profile
}

func (h *HTTPClient) SetProfile(p Profile) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.profile = p
}

func (h *HTTPClient) UpdateProfile(p Profile) {
	h.mu.Lock()
	h.profile = p
	sessPath := h.sessionPath
	profCopy := h.profile
	h.mu.Unlock()

	if sessPath != "" && profCopy.Token != "" {
		_ = SaveSession(sessPath, profCopy)
	}
}

func (h *HTTPClient) FetchChats() ([]Chat, error) {
	req, err := h.newRequest(http.MethodGet, "/api/v1/conversations", nil)
	if err != nil {
		return nil, err
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, h.formatNetworkError(err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("Failed to read server response")
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

	chats := make([]Chat, len(res.Data))
	for i, c := range res.Data {
		timeStr := formatFriendlyTime(c.LastMessageTime)
		chats[i] = Chat{
			ID:          c.ConvID,
			RecipientID: c.UserID,
			Name:        c.Name,
			Username:    c.Username,
			LastMessage: c.LastMessage,
			Time:        timeStr,
			Unread:      c.UnreadCount,
			Online:      false,
		}
	}

	h.mu.Lock()
	h.chats = chats
	h.mu.Unlock()

	return chats, nil
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
	h.chats = chats
}

func (h *HTTPClient) FetchMessages(chatID int64) ([]Message, error) {
	h.mu.RLock()
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

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("Failed to read server response")
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
	h.mu.RLock()
	for _, ch := range h.chats {
		if ch.ID == chatID {
			recipientName = ch.Name
			if recipientName == "" {
				recipientName = ch.Username
			}
			break
		}
	}
	h.mu.RUnlock()

	n := len(res.Data)
	messages := make([]Message, n)
	for i, c := range res.Data {
		isSelf := c.SenderID == currentUserID
		senderName := recipientName
		if isSelf {
			senderName = "You"
		}
		// API returns chats in DESC order (newest first).
		// Store them chronologically (oldest at index 0, newest at bottom index n-1)
		messages[n-1-i] = Message{
			ID:        c.ID,
			SenderID:  c.SenderID,
			Sender:    senderName,
			ConvID:    chatID,
			Text:      c.Content,
			Timestamp: formatFriendlyTime(c.CreatedAt),
			Self:      isSelf,
		}
	}

	h.mu.Lock()
	h.messages[chatID] = messages
	h.mu.Unlock()

	return messages, nil
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

func (h *HTTPClient) UpdateChatSnippet(convID int64, lastMsg, timeStr string, unreadDelta int, setExactUnread *int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for i := range h.chats {
		if h.chats[i].ID == convID {
			if lastMsg != "" {
				h.chats[i].LastMessage = lastMsg
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

// WebSocket Connection & Event Loop
func (h *HTTPClient) ConnectWS(eventsChan chan<- any) error {
	h.wsMu.Lock()
	defer h.wsMu.Unlock()

	if h.wsConnected.Load() && h.wsConn != nil {
		return nil
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

	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}

	conn, resp, err := dialer.Dial(wsURL, header)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusUnauthorized {
			return errors.New("Websocket authentication failed: unauthorized")
		}
		return errors.New("Chat server is currently offline or unreachable")
	}

	h.wsConn = conn
	h.wsConnected.Store(true)

	// Launch reader pump
	go func() {
		defer func() {
			h.wsConnected.Store(false)
			_ = conn.Close()
			eventsChan <- WSErrorPayload{
				Type:  "disconnected",
				Error: "websocket connection lost",
			}
		}()

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
					eventsChan <- p
				}
			case "chat_notification":
				var p WSChatNotificationPayload
				if err := json.Unmarshal(env.Payload, &p); err == nil {
					eventsChan <- p
				}
			case "conversation_read":
				var p WSConversationReadPayload
				if err := json.Unmarshal(env.Payload, &p); err == nil {
					eventsChan <- p
				}
			case "error":
				var p WSErrorPayload
				if err := json.Unmarshal(env.Payload, &p); err == nil {
					eventsChan <- p
				}
			}
		}
	}()

	return nil
}

func (h *HTTPClient) CloseWS() error {
	h.wsMu.Lock()
	defer h.wsMu.Unlock()

	h.wsConnected.Store(false)
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
	if !h.wsConnected.Load() || h.wsConn == nil {
		return errors.New("websocket is not connected")
	}

	h.wsMu.Lock()
	defer h.wsMu.Unlock()

	payload, _ := json.Marshal(map[string]any{
		"conv_id": convID,
		"content": text,
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
	if !h.wsConnected.Load() || h.wsConn == nil {
		return nil
	}

	h.wsMu.Lock()
	defer h.wsMu.Unlock()

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
	if !h.wsConnected.Load() || h.wsConn == nil {
		return nil
	}

	h.wsMu.Lock()
	defer h.wsMu.Unlock()

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
