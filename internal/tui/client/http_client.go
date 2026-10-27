package client

import (
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/DanielJohn17/tui-chat/internal/tui/crypto"
	"github.com/gorilla/websocket"
)

const (
	ClientHeaderApp    = "X-Client-App"
	ClientAppName      = "tui-chat"
	ClientHeaderSecret = "X-App-Secret"
)

const (
	maxHTTPPayloadBytes     int64 = 2 * 1024 * 1024  // 2 MB for standard responses
	maxHTTPBulkPayloadBytes int64 = 10 * 1024 * 1024 // 10 MB for bulk chat history
	maxWSPayloadBytes       int64 = 512 * 1024       // 512 KB per WebSocket message
	maxRenderTextLength           = 8000             // Clamp message length before lipgloss rendering
)

type HTTPClient struct {
	baseURL            string
	appSecret          string
	httpClient         *http.Client
	profile            Profile
	sessionPath        string
	mu                 sync.RWMutex
	sessionEpoch       uint64
	chats              []Chat
	onlineUsers        map[int64]bool
	messages           map[int64][]Message
	wsConn             *websocket.Conn
	wsConnected        atomic.Bool
	wsIntentionalClose atomic.Bool
	wsDoneChan         chan struct{}
	wsMu               sync.Mutex

	// E2EE fields
	identityKey  *crypto.IdentityKey
	peerPubKeys  map[int64][32]byte // userID → cached public key
	peerKeyMu    sync.RWMutex
	convIsSecret map[int64]bool
	convPeerID   map[int64]int64
}

func NewHTTPClient(baseURL, appSecret string) Client {
	if baseURL == "" {
		baseURL = os.Getenv("API_URL")
		if baseURL == "" {
			baseURL = "http://localhost:8080"
		}
	}
	baseURL = strings.TrimRight(baseURL, "/")

	return &HTTPClient{
		baseURL:   baseURL,
		appSecret: appSecret,
		httpClient: &http.Client{
			Timeout: 8 * time.Second,
		},
		onlineUsers:  make(map[int64]bool),
		messages:     make(map[int64][]Message),
		peerPubKeys:  make(map[int64][32]byte),
		convIsSecret: make(map[int64]bool),
		convPeerID:   make(map[int64]int64),
	}
}

// API Response Wrappers
type apiResponse[T any] struct {
	Success bool   `json:"success"`
	Data    T      `json:"data"`
	Error   string `json:"error,omitempty"`
}

func readLimitedBody(r io.Reader, limit int64) ([]byte, error) {
	lr := io.LimitReader(r, limit+1)
	bodyBytes, err := io.ReadAll(lr)
	if err != nil {
		return nil, errors.New("Failed to read server response")
	}
	if int64(len(bodyBytes)) > limit {
		return nil, errors.New("Server response exceeded maximum allowed payload size")
	}
	return bodyBytes, nil
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

	if h.appSecret != "" {
		req.Header.Set(ClientHeaderSecret, h.appSecret)
	}

	h.mu.RLock()
	token := h.profile.Token
	h.mu.RUnlock()
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return req, nil
}
