package users_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DanielJohn17/tui-chat/internal/api/auth"
	"github.com/DanielJohn17/tui-chat/internal/api/config"
	"github.com/DanielJohn17/tui-chat/internal/api/conversations"
	"github.com/DanielJohn17/tui-chat/internal/api/helpers"
	"github.com/DanielJohn17/tui-chat/internal/api/router"
	"github.com/DanielJohn17/tui-chat/internal/api/users"
	"github.com/DanielJohn17/tui-chat/internal/api/ws"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type mockUserRepoForProfile struct {
	users.UserRepositoryInt
	user *users.GetUserResponseType
}

func (m *mockUserRepoForProfile) GetUserByID(ctx context.Context, id int64) (*users.GetUserResponseType, error) {
	if m.user != nil && m.user.ID == id {
		return m.user, nil
	}
	return nil, errors.New("user not found")
}

func (m *mockUserRepoForProfile) GetUserByInviteToken(ctx context.Context, token string) (*users.GetUserResponseType, error) {
	if m.user != nil && m.user.InviteToken == token {
		return m.user, nil
	}
	return nil, errors.New("invalid invite token")
}

func (m *mockUserRepoForProfile) UpdateUserProfile(ctx context.Context, id int64, name, bio string, isSearchable bool) (*users.GetUserResponseType, error) {
	if m.user != nil && m.user.ID == id {
		m.user.Name = name
		m.user.Bio = bio
		m.user.IsSearchable = isSearchable
		return m.user, nil
	}
	return nil, errors.New("user not found")
}

func (m *mockUserRepoForProfile) RegenerateInviteToken(ctx context.Context, id int64, token string) (*users.GetUserResponseType, error) {
	if m.user != nil && m.user.ID == id {
		m.user.InviteToken = token
		return m.user, nil
	}
	return nil, errors.New("user not found")
}

func TestUserProfileAndInviteRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockUser := &users.GetUserResponseType{
		ID:           42,
		Name:         "Alice",
		Username:     "alice",
		Bio:          "Hello world",
		IsSearchable: true,
		InviteToken:  "inv_original123",
		CreatedAt:    "2026-01-01T00:00:00Z",
		UpdatedAt:    "2026-01-01T00:00:00Z",
	}

	repo := &mockUserRepoForProfile{user: mockUser}
	r := router.NewRouter(router.Handlers{
		User: users.NewUserHandler(users.NewUserService(repo)),
		Auth: &auth.AuthHandler{},
		Conv: &conversations.ConvHandler{},
		WS:   &ws.WSHandler{},
	}, ctx)

	token, err := helpers.CreateToken(helpers.UserToken{ID: 42, Username: "alice"})
	require.NoError(t, err)

	sendReq := func(method, path string, body any, useAuth bool) *httptest.ResponseRecorder {
		var reqBody *bytes.Buffer
		if body != nil {
			b, _ := json.Marshal(body)
			reqBody = bytes.NewBuffer(b)
		} else {
			reqBody = bytes.NewBuffer(nil)
		}

		req := httptest.NewRequest(method, path, reqBody)
		req.Header.Set(config.ClientSecretHeader, config.ENV.AppSharedSecret)
		if useAuth {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}

	t.Run("GET /users/me unauthorized without token", func(t *testing.T) {
		w := sendReq(http.MethodGet, "/api/v1/users/me", nil, false)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("GET /users/me success", func(t *testing.T) {
		w := sendReq(http.MethodGet, "/api/v1/users/me", nil, true)
		require.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Success bool                      `json:"success"`
			Data    users.GetUserResponseType `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		require.True(t, resp.Success)
		require.Equal(t, "Hello world", resp.Data.Bio)
		require.Equal(t, "inv_original123", resp.Data.InviteToken)
		require.True(t, resp.Data.IsSearchable)
	})

	t.Run("PUT /users/me update validation errors", func(t *testing.T) {
		// Empty name
		w := sendReq(http.MethodPut, "/api/v1/users/me", map[string]any{
			"name":          "",
			"bio":           "test",
			"is_searchable": true,
		}, true)
		require.Equal(t, http.StatusBadRequest, w.Code)

		// Bio exceeds 255 chars
		w = sendReq(http.MethodPut, "/api/v1/users/me", map[string]any{
			"name":          "Valid Name",
			"bio":           strings.Repeat("a", 256),
			"is_searchable": true,
		}, true)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("PUT /users/me update success to invite-only", func(t *testing.T) {
		w := sendReq(http.MethodPut, "/api/v1/users/me", map[string]any{
			"name":          "Alice In Wonderland",
			"bio":           "Hacking in private",
			"is_searchable": false,
		}, true)
		require.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Success bool                      `json:"success"`
			Data    users.GetUserResponseType `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		require.True(t, resp.Success)
		require.Equal(t, "Alice In Wonderland", resp.Data.Name)
		require.Equal(t, "Hacking in private", resp.Data.Bio)
		require.False(t, resp.Data.IsSearchable)
	})

	t.Run("POST /users/me/invite-token regenerate token", func(t *testing.T) {
		w := sendReq(http.MethodPost, "/api/v1/users/me/invite-token", nil, true)
		require.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Success bool                      `json:"success"`
			Data    users.GetUserResponseType `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		require.True(t, resp.Success)
		require.NotEqual(t, "inv_original123", resp.Data.InviteToken)
		require.NotEmpty(t, resp.Data.InviteToken)
	})

	t.Run("GET /users/invite/:token resolve public user", func(t *testing.T) {
		// Valid token
		w := sendReq(http.MethodGet, "/api/v1/users/invite/"+mockUser.InviteToken, nil, true)
		require.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Success bool             `json:"success"`
			Data    users.PublicUser `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		require.True(t, resp.Success)
		require.Equal(t, "alice", resp.Data.Username)
		require.Equal(t, "Hacking in private", resp.Data.Bio)

		// Invalid token
		w = sendReq(http.MethodGet, "/api/v1/users/invite/non_existent_token", nil, true)
		require.Equal(t, http.StatusNotFound, w.Code)
	})
}
