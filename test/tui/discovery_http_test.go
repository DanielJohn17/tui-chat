package test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DanielJohn17/tui-chat/internal/tui/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscoveryHTTPSearchUsers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/v1/users/search", r.URL.Path)
		assert.Equal(t, "q=alice+%2B+bob%26role%3Dadmin%3F%23%2F%25", r.URL.RawQuery)
		assert.Equal(t, "alice + bob&role=admin?#/%", r.URL.Query().Get("q"))
		assert.Len(t, r.URL.Query(), 1)
		assert.Equal(t, "Bearer discovery-token", r.Header.Get("Authorization"))
		assert.Equal(t, testAppSecret, r.Header.Get("X-App-Secret"))
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.Empty(t, body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"success":true,"data":{"users":[{"id":2,"name":"Alice","username":"alice"},{"id":3,"name":"Bob","username":"bob"}]}}`)
	}))
	defer server.Close()

	c := client.NewHTTPClient(server.URL, testAppSecret)
	c.SetProfile(client.Profile{ID: 1, Token: "discovery-token"})
	users, err := c.SearchUsers(context.Background(), " \t alice + bob&role=admin?#/% \n")
	require.NoError(t, err)
	assert.Equal(t, []client.UserSummary{
		{ID: 2, Name: "Alice", Username: "alice"},
		{ID: 3, Name: "Bob", Username: "bob"},
	}, users)
	assert.Empty(t, c.Chats())
}

func TestDiscoveryHTTPSearchUsersEmpty(t *testing.T) {
	for _, data := range []string{`{"users":[]}`, `{"users":null}`} {
		t.Run(data, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "q=", r.URL.RawQuery)
				_, _ = io.WriteString(w, `{"success":true,"data":`+data+`}`)
			}))
			defer server.Close()
			c := client.NewHTTPClient(server.URL, testAppSecret)
			c.SetProfile(client.Profile{ID: 1, Token: "discovery-token"})
			users, err := c.SearchUsers(context.Background(), " \t\n")
			require.NoError(t, err)
			assert.Empty(t, users)
		})
	}
}

func TestDiscoveryHTTPOpenDirectConversation(t *testing.T) {
	for _, status := range []int{http.StatusOK, http.StatusCreated} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/api/v1/conversations", r.URL.Path)
				assert.Empty(t, r.URL.RawQuery)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "Bearer discovery-token", r.Header.Get("Authorization"))
				assert.Equal(t, testAppSecret, r.Header.Get("X-App-Secret"))
				var body map[string]int64
				assert.NoError(t, json.NewDecoder(r.Body).Decode(&body))
				assert.Equal(t, map[string]int64{"user_id_two": 2}, body)
				w.WriteHeader(status)
				_, _ = io.WriteString(w, `{"success":true,"data":[{"conv_id":42,"user_id":1,"name":"Self","username":"self"},{"conv_id":42,"user_id":2,"name":"Alice","username":"alice","last_message":"Hello","last_message_time":"2000-01-02T12:34:56Z","unread_count":3}]}`)
			}))
			defer server.Close()
			c := client.NewHTTPClient(server.URL, testAppSecret)
			c.SetProfile(client.Profile{ID: 1, Token: "discovery-token"})
			c.SetUserOnline(2, true)
			existing := []client.Chat{{ID: 10, RecipientID: 3, Name: "Existing"}}
			c.SetChats(existing)
			chat, err := c.OpenDirectConversation(context.Background(), 2)
			require.NoError(t, err)
			assert.Equal(t, client.Chat{
				ID: 42, RecipientID: 2, Name: "Alice", Username: "alice",
				LastMessage: "Hello", Time: "02/01/00", Unread: 3, Online: true,
			}, chat)
			assert.Equal(t, existing, c.Chats(), "discovery must not apply results to the chat cache")
		})
	}
}

func TestDiscoveryHTTPResponseFailures(t *testing.T) {
	for _, method := range []string{"search", "open"} {
		t.Run(method, func(t *testing.T) {
			for _, tt := range []struct {
				name   string
				status int
				body   string
				err    string
			}{
				{"API error", 400, `{"success":false,"error":"Discovery denied"}`, "Discovery denied"},
				{"failure on OK", 200, `{"success":false,"error":"Not allowed"}`, "Not allowed"},
				{"failure without message", 200, `{"success":false}`, "Request failed (HTTP 200)"},
				{"HTTP failure despite success", 503, `{"success":true,"data":null}`, "Request failed (HTTP 503)"},
				{"unexpected success status", 202, `{"success":true,"data":null}`, "Request failed (HTTP 202)"},
				{"malformed envelope", 200, `{`, "Invalid server response format"},
				{"empty body", 200, ``, "Invalid server response format"},
				{"missing wrapper", 200, `{"users":[]}`, "Request failed (HTTP 200)"},
				{"missing data", 200, `{"success":true}`, "unexpected end of JSON input"},
				{"wrong data type", 200, `{"success":true,"data":"invalid"}`, "cannot unmarshal"},
			} {
				t.Run(tt.name, func(t *testing.T) {
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(tt.status)
						_, _ = io.WriteString(w, tt.body)
					}))
					defer server.Close()
					c := client.NewHTTPClient(server.URL, testAppSecret)
					c.SetProfile(client.Profile{ID: 1, Token: "discovery-token"})
					if method == "search" {
						users, err := c.SearchUsers(context.Background(), "alice")
						require.ErrorContains(t, err, tt.err)
						assert.Empty(t, users)
					} else {
						chat, err := c.OpenDirectConversation(context.Background(), 2)
						require.ErrorContains(t, err, tt.err)
						assert.Equal(t, client.Chat{}, chat)
					}
				})
			}
		})
	}
}

func TestDiscoveryHTTPOpenInvalidConversation(t *testing.T) {
	for _, data := range []string{
		`[]`, `null`,
		`[{"conv_id":42,"user_id":1}]`,
		`[{"conv_id":42,"user_id":3}]`,
		`[{"conv_id":0,"user_id":2}]`,
		`[{"conv_id":-1,"user_id":2}]`,
	} {
		t.Run(data, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.WriteString(w, `{"success":true,"data":`+data+`}`)
			}))
			defer server.Close()
			c := client.NewHTTPClient(server.URL, testAppSecret)
			c.SetProfile(client.Profile{ID: 1, Token: "discovery-token"})
			chat, err := c.OpenDirectConversation(context.Background(), 2)
			require.EqualError(t, err, "Server returned an invalid conversation")
			assert.Equal(t, client.Chat{}, chat)
		})
	}
}

func TestDiscoveryHTTPValidationBeforeRequest(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	c := client.NewHTTPClient(server.URL, testAppSecret)
	users, err := c.SearchUsers(context.Background(), "alice")
	require.EqualError(t, err, "unauthenticated")
	assert.Empty(t, users)
	chat, err := c.OpenDirectConversation(context.Background(), 2)
	require.EqualError(t, err, "unauthenticated")
	assert.Equal(t, client.Chat{}, chat)
	c.SetProfile(client.Profile{ID: 1, Token: "discovery-token"})
	for _, id := range []int64{-1, 0, 1} {
		chat, err := c.OpenDirectConversation(context.Background(), id)
		require.EqualError(t, err, "Select another user to start a conversation")
		assert.Equal(t, client.Chat{}, chat)
	}
	assert.Zero(t, requests.Load())
}

func TestDiscoveryHTTPCancellationAndStaleSession(t *testing.T) {
	for _, method := range []string{"search", "open"} {
		t.Run(method, func(t *testing.T) {
			for _, change := range []string{"cancel", "logout", "logout and restore", "replace token", "replace user"} {
				t.Run(change, func(t *testing.T) {
					started := make(chan struct{})
					release := make(chan struct{})
					var once sync.Once
					unblock := func() { once.Do(func() { close(release) }) }
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						close(started)
						select {
						case <-release:
						case <-r.Context().Done():
							return
						}
						if method == "search" {
							_, _ = io.WriteString(w, `{"success":true,"data":{"users":[{"id":2,"name":"Stale"}]}}`)
						} else {
							_, _ = io.WriteString(w, `{"success":true,"data":[{"conv_id":42,"user_id":2,"name":"Stale"}]}`)
						}
					}))
					defer server.Close()
					defer unblock()
					ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
					defer cancel()
					c := client.NewHTTPClient(server.URL, testAppSecret)
					profile := client.Profile{ID: 1, Token: "discovery-token"}
					c.SetProfile(profile)
					done := make(chan error, 1)
					go func() {
						if method == "search" {
							users, err := c.SearchUsers(ctx, "alice")
							assert.Empty(t, users)
							done <- err
						} else {
							chat, err := c.OpenDirectConversation(ctx, 2)
							assert.Equal(t, client.Chat{}, chat)
							done <- err
						}
					}()
					select {
					case <-started:
					case <-ctx.Done():
						t.Fatal("discovery request did not reach server")
					}
					switch change {
					case "cancel":
						cancel()
					case "logout", "logout and restore":
						require.NoError(t, c.Logout())
						if change == "logout and restore" {
							c.SetProfile(profile)
						}
					case "replace token":
						profile.Token = "new-token"
						c.SetProfile(profile)
					case "replace user":
						profile.ID = 3
						c.SetProfile(profile)
					}
					unblock()
					select {
					case err := <-done:
						if change == "cancel" {
							require.ErrorIs(t, err, context.Canceled)
						} else {
							require.EqualError(t, err, "session invalidated")
						}
					case <-time.After(3 * time.Second):
						t.Fatal("discovery request did not finish")
					}
					assert.Empty(t, c.Chats())
					assert.Empty(t, c.Messages(42))
				})
			}
		})
	}
}
