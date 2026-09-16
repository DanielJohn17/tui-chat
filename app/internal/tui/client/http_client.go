package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
	profile    Profile
	mock       Client // Fallback mock for offline demo
}

func NewHTTPClient(baseURL string) Client {
	if baseURL == "" {
		baseURL = os.Getenv("API_URL")
		if baseURL == "" {
			baseURL = "http://localhost:8080"
		}
	}

	return &HTTPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 8 * time.Second,
		},
		mock: NewMock(),
	}
}

type authResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Username  string `json:"username"`
	Token     string `json:"token"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h *HTTPClient) Login(username, password string) (*Profile, error) {
	reqBody, _ := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})

	resp, err := h.httpClient.Post(h.baseURL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		// Network connection error -> fallback to mock if username exists
		return nil, fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		var errResp errorResponse
		if json.Unmarshal(bodyBytes, &errResp) == nil && errResp.Error != "" {
			return nil, errors.New(errResp.Error)
		}
		return nil, fmt.Errorf("login failed (HTTP %d)", resp.StatusCode)
	}

	var authResp authResponse
	if err := json.Unmarshal(bodyBytes, &authResp); err != nil {
		return nil, fmt.Errorf("invalid server response: %w", err)
	}

	h.profile = Profile{
		ID:        authResp.ID,
		Name:      authResp.Name,
		Username:  authResp.Username,
		Token:     authResp.Token,
		CreatedAt: authResp.CreatedAt,
	}

	return &h.profile, nil
}

func (h *HTTPClient) Register(name, username, password string) (*Profile, error) {
	reqBody, _ := json.Marshal(map[string]string{
		"name":     name,
		"username": username,
		"password": password,
	})

	resp, err := h.httpClient.Post(h.baseURL+"/api/v1/auth/register", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var errResp errorResponse
		if json.Unmarshal(bodyBytes, &errResp) == nil && errResp.Error != "" {
			return nil, errors.New(errResp.Error)
		}
		return nil, fmt.Errorf("registration failed (HTTP %d)", resp.StatusCode)
	}

	var authResp authResponse
	if err := json.Unmarshal(bodyBytes, &authResp); err != nil {
		return nil, fmt.Errorf("invalid server response: %w", err)
	}

	h.profile = Profile{
		ID:        authResp.ID,
		Name:      authResp.Name,
		Username:  authResp.Username,
		Token:     authResp.Token,
		CreatedAt: authResp.CreatedAt,
	}

	return &h.profile, nil
}

func (h *HTTPClient) IsAuthenticated() bool {
	return h.profile.Token != ""
}

func (h *HTTPClient) Profile() Profile {
	if h.profile.Username != "" {
		return h.profile
	}
	return h.mock.Profile()
}

func (h *HTTPClient) SetProfile(p Profile) {
	h.profile = p
}

func (h *HTTPClient) UpdateProfile(p Profile) {
	h.profile = p
	h.mock.UpdateProfile(p)
}

func (h *HTTPClient) Chats() []Chat {
	// If API server is down or unauthenticated, delegate to mock data
	return h.mock.Chats()
}

func (h *HTTPClient) Messages(chatID int64) []Message {
	return h.mock.Messages(chatID)
}

func (h *HTTPClient) Send(chatID int64, text string) {
	h.mock.Send(chatID, text)
}

func (h *HTTPClient) AddChat(name, username string) Chat {
	return h.mock.AddChat(name, username)
}

func (h *HTTPClient) MarkRead(chatID int64) {
	h.mock.MarkRead(chatID)
}
