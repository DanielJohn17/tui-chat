package test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DanielJohn17/tui-chat/internal/tui/client"
	"github.com/DanielJohn17/tui-chat/internal/tui/theme"
	"github.com/DanielJohn17/tui-chat/internal/tui/views/chat"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 1. Sanitization tests for OSC 52 clipboard injection, CSI escape sequences, and control chars
func TestSanitizeANSIAndOSC52(t *testing.T) {
	// Test OSC 52 clipboard injection payload with BEL terminator
	osc52PayloadBEL := "malicious text \x1b]52;c;c2VjcmV0\x07 payload"
	sanitizedBEL := theme.SanitizeText(osc52PayloadBEL)
	assert.NotContains(t, sanitizedBEL, "\x1b]52;")
	assert.NotContains(t, sanitizedBEL, "\x07")
	assert.Equal(t, "malicious text  payload", sanitizedBEL)

	// Test OSC 52 with String Terminator (ESC \)
	osc52PayloadST := "attack \x1b]52;c;dGVzdA==\x1b\\ done"
	sanitizedST := theme.SanitizeText(osc52PayloadST)
	assert.NotContains(t, sanitizedST, "\x1b]52;")
	assert.NotContains(t, sanitizedST, "\x1b\\")
	assert.Equal(t, "attack  done", sanitizedST)

	// Test CSI color codes and cursor movement
	csiPayload := "\x1b[31;1mRed Alert!\x1b[0m \x1b[2JScreen Cleared\x1b[H"
	sanitizedCSI := theme.SanitizeText(csiPayload)
	assert.Equal(t, "Red Alert! Screen Cleared", sanitizedCSI)

	// Test raw control characters
	controlChars := "Hello\x00\x01\x02\x03\x04\x05\x06\a\bWorld\r\n"
	sanitizedControls := theme.SanitizeText(controlChars)
	assert.Equal(t, "HelloWorld\n", sanitizedControls)

	// Test SanitizeLine collapses newlines and trims whitespace
	multiLine := "Line 1\r\n\x1b[32mLine 2\x1b[0m\nLine 3"
	sanitizedLine := theme.SanitizeLine(multiLine)
	assert.Equal(t, "Line 1 Line 2 Line 3", sanitizedLine)

	// Test preservation of UTF-8, multi-byte emojis, and formatting
	safeText := "Hello 🌍! 🚀 Testing UTF-8 runes: 日本語, привет\n\tIndented"
	assert.Equal(t, safeText, theme.SanitizeText(safeText))
}

// FailingSendClient simulates SendWS error
type FailingSendClient struct {
	*TestClient
	sendErr error
}

func (f *FailingSendClient) SendWS(convID int64, text string) error {
	if f.sendErr != nil {
		return f.sendErr
	}
	return f.TestClient.SendWS(convID, text)
}

// 2. Test that SendWS failure does NOT append message and retains typed input
func TestSendWSFailureDoesNotAppendMessage(t *testing.T) {
	baseClient := newTestClient()
	failingClient := &FailingSendClient{
		TestClient: baseClient,
		sendErr:    errors.New("connection broken"),
	}

	chatView := chat.New(failingClient)
	chatView.SetSize(80, 24)
	chatView.SetActiveChat(1)
	chatView.FocusInput()

	initialMsgCount := len(failingClient.Messages(1))

	// Type a message into input
	testDraft := "Important message that shouldn't be lost"
	for _, r := range testDraft {
		chatView, _ = chatView.Update(tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{r},
		})
	}

	// Press enter
	chatView, _ = chatView.Update(tea.KeyMsg{
		Type: tea.KeyEnter,
	})

	// Assert message was NOT appended to client memory
	afterMsgCount := len(failingClient.Messages(1))
	assert.Equal(t, initialMsgCount, afterMsgCount, "Message must not be appended when SendWS fails")

	// Assert input was retained so user draft is not lost
	assert.Equal(t, testDraft, chatView.InputValue(), "User input must be preserved on send failure")
}

// 3. Test that in-flight fetches do not overwrite post-logout state
func TestFetchSessionEpochDiscardAfterLogout(t *testing.T) {
	requestStarted := make(chan struct{})
	allowResponse := make(chan struct{})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/conversations") {
			close(requestStarted)
			<-allowResponse
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"success": true,
				"data": []map[string]any{
					{
						"conv_id":           1,
						"user_id":           102,
						"name":              "Stale User",
						"username":          "stale",
						"last_message":      "Stale message",
						"last_message_time": "2026-09-28T12:00:00Z",
						"unread_count":      0,
					},
				},
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	httpClient := client.NewHTTPClient(server.URL)
	httpClient.SetProfile(client.Profile{
		ID:    1,
		Token: "test-token",
	})

	// Start FetchChats asynchronously
	fetchDone := make(chan error, 1)
	go func() {
		_, err := httpClient.FetchChats()
		fetchDone <- err
	}()

	// Wait for request to be in flight
	<-requestStarted

	// User logs out while fetch is in-flight
	_ = httpClient.Logout()

	// Release HTTP server response
	close(allowResponse)

	// Fetch should return session invalidated error
	err := <-fetchDone
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session invalidated")

	// Ensure chats remain empty
	assert.Empty(t, httpClient.Chats(), "Chats should remain empty after logout despite completing in-flight fetch")
}

// 4. Test that FetchBulkMessages merges with existing WebSocket messages without wiping them out
func TestFetchBulkMessagesMergesWithoutOverwritingNewerWSMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"data": []map[string]any{
				{
					"id":         1,
					"sender_id":  102,
					"conv_id":    1,
					"content":    "Old message from bulk",
					"created_at": "2026-09-28T10:00:00Z",
					"updated_at": "2026-09-28T10:00:00Z",
				},
			},
		})
	}))
	defer server.Close()

	httpClient := client.NewHTTPClient(server.URL)
	httpClient.SetProfile(client.Profile{
		ID:    1,
		Token: "test-token",
	})

	// Simulate a real-time message received via WebSocket before or during bulk fetch
	wsMessage := client.Message{
		ID:        999,
		ConvID:    1,
		SenderID:  102,
		Sender:    "Bob",
		Text:      "Real-time message arrived first!",
		Timestamp: "now",
	}
	httpClient.AppendMessage(wsMessage)

	// Fetch bulk messages
	_, err := httpClient.FetchBulkMessages()
	require.NoError(t, err)

	// Verify both the fetched message and the real-time message are present
	messages := httpClient.Messages(1)
	require.Len(t, messages, 2, "Bulk fetch must merge rather than overwrite existing messages")
	assert.Equal(t, int64(1), messages[0].ID)
	assert.Equal(t, int64(999), messages[1].ID)
}

// 5. Test that CloseWS does NOT emit a disconnected error event
func TestCloseWSDistinguishesIntentionalClose(t *testing.T) {
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		// Keep open until client closes
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}))
	defer server.Close()

	httpClient := client.NewHTTPClient(server.URL)
	httpClient.SetProfile(client.Profile{
		ID:    1,
		Token: "test-token",
	})

	eventsChan := make(chan any, 10)
	err := httpClient.ConnectWS(eventsChan)
	require.NoError(t, err)

	// Close WS intentionally
	err = httpClient.CloseWS()
	require.NoError(t, err)

	// Give reader defer a brief moment to execute
	time.Sleep(50 * time.Millisecond)

	// Verify eventsChan did not receive a "disconnected" payload
	select {
	case evt := <-eventsChan:
		t.Fatalf("Unexpected event received on intentional close: %+v", evt)
	default:
		// Clean: no unwanted reconnect event fired
	}
}
