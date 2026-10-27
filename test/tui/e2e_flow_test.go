package test

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DanielJohn17/tui-chat/internal/tui/client"
	"github.com/DanielJohn17/tui-chat/internal/tui/crypto"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPClient_PeerPublicKeyAndE2EE(t *testing.T) {
	// Generate Alice & Bob keys
	aliceKey, err := crypto.GenerateIdentityKey()
	require.NoError(t, err)
	bobKey, err := crypto.GenerateIdentityKey()
	require.NoError(t, err)

	bobPubHex := hex.EncodeToString(bobKey.PublicKey[:])
	var uploadedAliceKey string

	// Create a mock API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/api/v1/keys/me":
			if r.Method == http.MethodPut {
				var req map[string]string
				_ = json.NewDecoder(r.Body).Decode(&req)
				uploadedAliceKey = req["public_key"]
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"success":true}`))
				return
			}
		case "/api/v1/keys/2": // Bob's user ID
			if r.Method == http.MethodGet {
				w.WriteHeader(http.StatusOK)
				resp := fmt.Sprintf(`{"success":true,"data":{"user_id":2,"public_key":%q}}`, bobPubHex)
				_, _ = w.Write([]byte(resp))
				return
			}
		case "/api/v1/keys/99": // User without key
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"success":false,"error":"public key not found"}`))
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c := client.NewHTTPClient(server.URL, "secret")
	c.SetIdentityKey(aliceKey)

	t.Run("UploadPublicKey sends valid hex", func(t *testing.T) {
		err := c.UploadPublicKey(aliceKey.PublicKey[:])
		require.NoError(t, err)
		assert.Equal(t, hex.EncodeToString(aliceKey.PublicKey[:]), uploadedAliceKey)
	})

	t.Run("GetPeerPublicKey correctly decodes hex and caches result", func(t *testing.T) {
		// First fetch: hits HTTP server
		pubKey, err := c.GetPeerPublicKey(2)
		require.NoError(t, err, "GetPeerPublicKey must succeed and decode 32 bytes hex")
		assert.Equal(t, bobKey.PublicKey[:], pubKey)

		// Second fetch: served from cache
		pubKeyCached, err := c.GetPeerPublicKey(2)
		require.NoError(t, err)
		assert.Equal(t, pubKey, pubKeyCached)
	})

	t.Run("GetPeerPublicKey returns error on 404", func(t *testing.T) {
		_, err := c.GetPeerPublicKey(99)
		assert.Error(t, err)
	})

	t.Run("Logout clears cached identity key and peer public keys", func(t *testing.T) {
		assert.NotNil(t, c.IdentityKey())
		err := c.Logout()
		require.NoError(t, err)

		// Identity key should be cleared
		assert.Nil(t, c.IdentityKey())
	})
}

func TestHTTPClient_CloudPlaintext_vs_SecretEncrypted(t *testing.T) {
	aliceKey, err := crypto.GenerateIdentityKey()
	require.NoError(t, err)
	bobKey, err := crypto.GenerateIdentityKey()
	require.NoError(t, err)

	bobPubHex := hex.EncodeToString(bobKey.PublicKey[:])
	alicePubHex := hex.EncodeToString(aliceKey.PublicKey[:])

	// Derived shared key between Alice & Bob
	sharedKey, err := crypto.DeriveConvKey(aliceKey.PrivateKey, bobKey.PublicKey)
	require.NoError(t, err)

	// Create test HTTP/WS server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/keys/2":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(fmt.Sprintf(`{"success":true,"data":{"user_id":2,"public_key":%q}}`, bobPubHex)))
			return
		case "/api/v1/keys/1":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(fmt.Sprintf(`{"success":true,"data":{"user_id":1,"public_key":%q}}`, alicePubHex)))
			return
		case "/api/v1/conversations":
			w.WriteHeader(http.StatusOK)
			// Return two conversations for Bob (user 2): cloud (conv 10) and secret (conv 20)
			_, _ = w.Write([]byte(`{
				"success": true,
				"data": [
					{"conv_id": 10, "user_id": 2, "name": "Bob", "username": "bob", "last_message": "cloud msg", "is_secret": false},
					{"conv_id": 20, "user_id": 2, "name": "Bob", "username": "bob", "last_message": "secret msg", "is_secret": true}
				]
			}`))
			return
		case "/api/v1/conversations/10/chats":
			// Cloud messages are stored as plaintext
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"success": true,
				"data": [
					{"id": 1, "conv_id": 10, "sender_id": 2, "content": "Plaintext cloud message", "created_at": "2026-10-12T12:00:00Z"}
				]
			}`))
			return
		case "/api/v1/conversations/20/chats":
			// API returns chats in DESC order (newest first).
			// id: 3 is newer, id: 2 is older.
			encryptedMsg, encErr := crypto.Encrypt(sharedKey, "Confidential secret payload")
			require.NoError(t, encErr)

			// Generate another device's encrypted message (different key) -> fails decryption with ErrDecryptFailed
			otherKey, otherErr := crypto.DeriveConvKey(bobKey.PrivateKey, bobKey.PublicKey)
			require.NoError(t, otherErr)
			foreignMsg, encForeignErr := crypto.Encrypt(otherKey, "Secret from Bob other device")
			require.NoError(t, encForeignErr)

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(fmt.Sprintf(`{
				"success": true,
				"data": [
					{"id": 3, "conv_id": 20, "sender_id": 2, "content": %q, "created_at": "2026-10-12T12:06:00Z"},
					{"id": 2, "conv_id": 20, "sender_id": 2, "content": %q, "created_at": "2026-10-12T12:05:00Z"}
				]
			}`, foreignMsg, encryptedMsg)))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c := client.NewHTTPClient(server.URL, "secret")
	c.SetIdentityKey(aliceKey)
	c.SetProfile(client.Profile{ID: 1, Name: "Alice", Token: "alice-token"})

	// Fetch chats so client tracks convIsSecret and contacts
	chats, err := c.FetchChats()
	require.NoError(t, err)
	require.Len(t, chats, 1)
	assert.Equal(t, int64(10), chats[0].CloudID)
	assert.Equal(t, int64(20), chats[0].SecretID)

	// Verify FetchMessages: Cloud returns plaintext
	cloudMsgs, err := c.FetchMessages(10)
	require.NoError(t, err)
	require.Len(t, cloudMsgs, 1)
	assert.Equal(t, "Plaintext cloud message", cloudMsgs[0].Text)
	assert.False(t, cloudMsgs[0].IsSecret)

	// Verify FetchMessages: Secret decrypts valid message and flags foreign key as cannot decrypt
	secretMsgs, err := c.FetchMessages(20)
	require.NoError(t, err)
	require.Len(t, secretMsgs, 2)
	assert.Equal(t, "Confidential secret payload", secretMsgs[0].Text)
	assert.True(t, secretMsgs[0].IsSecret)
	assert.Equal(t, "[cannot decrypt — key from another device]", secretMsgs[1].Text)
	assert.True(t, secretMsgs[1].IsSecret)

	// -------------------------------------------------------------
	// Verify SendWS: Cloud sends plaintext, Secret sends ciphertext
	// -------------------------------------------------------------
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	receivedWS := make(chan []byte, 10)

	serverCombined := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Upgrade") == "websocket" {
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}
			defer conn.Close()
			for {
				_, msg, err := conn.ReadMessage()
				if err != nil {
					break
				}
				receivedWS <- msg
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/v1/keys/2" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(fmt.Sprintf(`{"success":true,"data":{"user_id":2,"public_key":%q}}`, bobPubHex)))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer serverCombined.Close()

	wsClient := client.NewHTTPClient(serverCombined.URL, "secret")
	wsClient.SetIdentityKey(aliceKey)
	wsClient.SetProfile(client.Profile{ID: 1, Name: "Alice", Token: "alice-token"})
	wsClient.SetChats([]client.Chat{
		{
			ID:          10,
			RecipientID: 2,
			CloudID:     10,
			SecretID:    20,
		},
	})
	wsClient.SetConvSecret(10, false)
	wsClient.SetConvSecret(20, true)

	// Connect WS
	eventsChan := make(chan any, 10)
	err = wsClient.ConnectWS(eventsChan)
	require.NoError(t, err)
	defer wsClient.CloseWS()

	// 1. Send Cloud message (conv 10, isSecret = false) -> plaintext
	err = wsClient.SendWS(10, "Hello over Cloud")
	require.NoError(t, err)

	select {
	case raw := <-receivedWS:
		var env struct {
			Action  string `json:"action"`
			Payload struct {
				ConvID  int64  `json:"conv_id"`
				Content string `json:"content"`
			} `json:"payload"`
		}
		require.NoError(t, json.Unmarshal(raw, &env))
		assert.Equal(t, "send_message", env.Action)
		assert.Equal(t, int64(10), env.Payload.ConvID)
		assert.Equal(t, "Hello over Cloud", env.Payload.Content, "Cloud message must be sent in plaintext")
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for cloud WS message")
	}

	// 2. Send Secret message (conv 20, isSecret = true) -> ciphertext
	err = wsClient.SendWS(20, "Secret message for Bob")
	require.NoError(t, err)

	select {
	case raw := <-receivedWS:
		var env struct {
			Action  string `json:"action"`
			Payload struct {
				ConvID  int64  `json:"conv_id"`
				Content string `json:"content"`
			} `json:"payload"`
		}
		require.NoError(t, json.Unmarshal(raw, &env))
		assert.Equal(t, "send_message", env.Action)
		assert.Equal(t, int64(20), env.Payload.ConvID)
		assert.NotEqual(t, "Secret message for Bob", env.Payload.Content, "Secret message must NOT be plaintext")

		// Verify ciphertext decrypts with sharedKey
		decrypted, decErr := crypto.Decrypt(sharedKey, env.Payload.Content)
		require.NoError(t, decErr)
		assert.Equal(t, "Secret message for Bob", decrypted)
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for secret WS message")
	}
}
