package client

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/DanielJohn17/tui-chat/internal/tui/crypto"
)

func (h *HTTPClient) SetIdentityKey(key *crypto.IdentityKey) {
	h.peerKeyMu.Lock()
	defer h.peerKeyMu.Unlock()
	h.identityKey = key
}

func (h *HTTPClient) IdentityKey() *crypto.IdentityKey {
	h.peerKeyMu.RLock()
	defer h.peerKeyMu.RUnlock()
	return h.identityKey
}

func (h *HTTPClient) UploadPublicKey(pubKey []byte) error {
	hexKey := fmt.Sprintf("%x", pubKey)
	reqBody, _ := json.Marshal(map[string]string{
		"public_key": hexKey,
	})

	req, err := h.newRequest(http.MethodPut, "/api/v1/keys/me", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return h.formatNetworkError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to upload public key: HTTP %d", resp.StatusCode)
	}
	return nil
}

func (h *HTTPClient) fetchPeerPublicKeyFromAPI(userID int64) ([]byte, error) {
	req, err := h.newRequest(http.MethodGet, fmt.Sprintf("/api/v1/keys/%d", userID), nil)
	if err != nil {
		return nil, err
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, h.formatNetworkError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch peer public key: HTTP %d", resp.StatusCode)
	}

	bodyBytes, err := readLimitedBody(resp.Body, maxHTTPPayloadBytes)
	if err != nil {
		return nil, err
	}

	// PublicKeyResponse has { user_id, public_key string }
	var res apiResponse[struct {
		UserID    int64  `json:"user_id"`
		PublicKey string `json:"public_key"`
	}]
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return nil, errors.New("invalid public key response")
	}

	decoded, err := hex.DecodeString(res.Data.PublicKey)
	if err != nil || len(decoded) != 32 {
		return nil, errors.New("invalid hex in public key response: expected 32 bytes")
	}

	var pubKeyBytes [32]byte
	copy(pubKeyBytes[:], decoded)

	h.peerKeyMu.Lock()
	h.peerPubKeys[userID] = pubKeyBytes
	h.peerKeyMu.Unlock()

	return pubKeyBytes[:], nil
}

func (h *HTTPClient) GetPeerPublicKey(userID int64) ([]byte, error) {
	h.peerKeyMu.RLock()
	cached, ok := h.peerPubKeys[userID]
	h.peerKeyMu.RUnlock()
	if ok {
		return cached[:], nil
	}

	return h.fetchPeerPublicKeyFromAPI(userID)
}

func (h *HTTPClient) DecryptMessage(peerID int64, ciphertext string) string {
	if ciphertext == "" || peerID == 0 {
		return ciphertext
	}
	peerPub, err := h.GetPeerPublicKey(peerID)
	if err != nil {
		return "[cannot decrypt — key from another device]"
	}
	var peerArr [32]byte
	copy(peerArr[:], peerPub)

	idKey := h.IdentityKey()
	if idKey == nil {
		return "[cannot decrypt — key from another device]"
	}

	symKey, err := crypto.DeriveConvKey(idKey.PrivateKey, peerArr)
	if err != nil {
		return "[cannot decrypt — key from another device]"
	}

	decrypted, err := crypto.Decrypt(symKey, ciphertext)
	if err == nil {
		return decrypted
	}
	if errors.Is(err, crypto.ErrLegacyMessage) {
		return ciphertext
	}

	// If decryption failed, peer may have uploaded a new identity key.
	// Invalidate cache and fetch the latest key from the server once.
	if freshPub, err := h.fetchPeerPublicKeyFromAPI(peerID); err == nil && !bytes.Equal(freshPub, peerPub) {
		var freshArr [32]byte
		copy(freshArr[:], freshPub)
		if freshSymKey, err := crypto.DeriveConvKey(idKey.PrivateKey, freshArr); err == nil {
			if freshDec, err := crypto.Decrypt(freshSymKey, ciphertext); err == nil {
				return freshDec
			}
		}
	}

	return "[cannot decrypt — key from another device]"
}
