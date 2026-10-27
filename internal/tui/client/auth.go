package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type authData struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Username  string `json:"username"`
	Token     string `json:"token"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
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

	bodyBytes, err := readLimitedBody(resp.Body, maxHTTPPayloadBytes)
	if err != nil {
		return nil, err
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
	h.sessionEpoch++
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

	bodyBytes, err := readLimitedBody(resp.Body, maxHTTPPayloadBytes)
	if err != nil {
		return nil, err
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
	h.sessionEpoch++
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
	h.sessionEpoch++
	sessPath := h.sessionPath
	h.profile = Profile{}
	h.chats = nil
	h.messages = make(map[int64][]Message)
	h.convIsSecret = make(map[int64]bool)
	h.convPeerID = make(map[int64]int64)
	h.mu.Unlock()

	h.peerKeyMu.Lock()
	h.identityKey = nil
	h.peerPubKeys = make(map[int64][32]byte)
	h.peerKeyMu.Unlock()

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
