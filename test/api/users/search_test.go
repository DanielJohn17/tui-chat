package users_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/DanielJohn17/tui-chat/internal/api/apierrs"
	"github.com/DanielJohn17/tui-chat/internal/api/auth"
	"github.com/DanielJohn17/tui-chat/internal/api/config"
	"github.com/DanielJohn17/tui-chat/internal/api/conversations"
	"github.com/DanielJohn17/tui-chat/internal/api/database"
	"github.com/DanielJohn17/tui-chat/internal/api/helpers"
	"github.com/DanielJohn17/tui-chat/internal/api/router"
	"github.com/DanielJohn17/tui-chat/internal/api/users"
	"github.com/DanielJohn17/tui-chat/internal/api/ws"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type searchRepository struct {
	users.UserRepositoryInt
	search func(context.Context, int64, string) ([]users.PublicUser, error)
}

func (r searchRepository) SearchUsers(ctx context.Context, id int64, q string) ([]users.PublicUser, error) {
	return r.search(ctx, id, q)
}

type searchQuerier struct {
	users.UserQuerier
	search func(context.Context, database.SearchUsersParams) ([]database.SearchUsersRow, error)
}

func (q searchQuerier) SearchUsers(ctx context.Context, p database.SearchUsersParams) ([]database.SearchUsersRow, error) {
	return q.search(ctx, p)
}

func TestSearchValidation(t *testing.T) {
	for _, q := range []string{"", "  ", "a", strings.Repeat("a", 21), string([]byte{0xff, 0xff})} {
		t.Run(url.QueryEscape(q), func(t *testing.T) {
			s := users.NewUserService(searchRepository{search: func(context.Context, int64, string) ([]users.PublicUser, error) {
				t.Fatal("invalid query reached repository")
				return nil, nil
			}})
			result, err := s.SearchUsers(context.Background(), 7, q)
			require.Nil(t, result)
			var apiErr *apierrs.APIError
			require.ErrorAs(t, err, &apiErr)
			require.Equal(t, http.StatusBadRequest, apiErr.Code)
		})
	}
	for _, q := range []string{"ab", strings.Repeat("a", 20), "\u00e9\u00e9", strings.Repeat("\u00e9", 20)} {
		s := users.NewUserService(searchRepository{search: func(ctx context.Context, id int64, query string) ([]users.PublicUser, error) {
			require.Equal(t, int64(7), id)
			require.Equal(t, q, query)
			return nil, nil
		}})
		result, err := s.SearchUsers(context.Background(), 7, " \t"+q+"\n")
		require.NoError(t, err)
		require.NotNil(t, result.Users)
		require.Empty(t, result.Users)
	}
}

func TestSearchRepository(t *testing.T) {
	for query, pattern := range map[string]string{"Ab": "Ab%", "a%b_c\\d": `a\%b\_c\\d%`} {
		r := users.NewUserRepository(searchQuerier{search: func(ctx context.Context, p database.SearchUsersParams) ([]database.SearchUsersRow, error) {
			require.Equal(t, database.SearchUsersParams{UserID: 7, Pattern: pattern, Query: query}, p)
			return []database.SearchUsersRow{{ID: 8, Name: "Alice", Username: "alice"}}, nil
		}})
		result, err := r.SearchUsers(context.Background(), 7, query)
		require.NoError(t, err)
		require.Equal(t, []users.PublicUser{{ID: 8, Name: "Alice", Username: "alice"}}, result)
	}
	r := users.NewUserRepository(searchQuerier{search: func(context.Context, database.SearchUsersParams) ([]database.SearchUsersRow, error) { return nil, nil }})
	result, err := r.SearchUsers(context.Background(), 7, "ab")
	require.NoError(t, err)
	require.NotNil(t, result)
	dbErr := errors.New("database unavailable")
	r = users.NewUserRepository(searchQuerier{search: func(context.Context, database.SearchUsersParams) ([]database.SearchUsersRow, error) {
		return nil, dbErr
	}})
	_, err = r.SearchUsers(context.Background(), 7, "ab")
	require.ErrorIs(t, err, dbErr)
	_, err = users.NewUserService(r).SearchUsers(context.Background(), 7, "ab")
	var apiErr *apierrs.APIError
	require.ErrorAs(t, err, &apiErr)
	require.Equal(t, http.StatusInternalServerError, apiErr.Code)
}

func TestSearchRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	calls := 0
	repo := searchRepository{search: func(ctx context.Context, id int64, query string) ([]users.PublicUser, error) {
		calls++
		require.Equal(t, int64(7), id)
		if query == "fail" {
			return nil, errors.New("private database details")
		}
		if query == "none" {
			return nil, nil
		}
		require.Equal(t, "Al", query)
		return []users.PublicUser{{ID: 8, Name: "Alice", Username: "alice"}}, nil
	}}
	r := router.NewRouter(router.Handlers{
		User: users.NewUserHandler(users.NewUserService(repo)),
		Auth: &auth.AuthHandler{}, Conv: &conversations.ConvHandler{}, WS: &ws.WSHandler{},
	}, ctx)
	token, err := helpers.CreateToken(helpers.UserToken{ID: 7, Username: "self"})
	require.NoError(t, err)
	request := func(q, authorization, secret string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users/search?q="+url.QueryEscape(q), nil)
		req.Header.Set(config.ClientSecretHeader, secret)
		req.Header.Set("Authorization", authorization)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}
	for i := 0; i < 25; i++ {
		require.Equal(t, http.StatusUnauthorized, request("Al", "", config.ENV.AppSharedSecret).Code)
	}
	require.Equal(t, http.StatusUnauthorized, request("Al", "Bearer invalid", config.ENV.AppSharedSecret).Code)
	require.Equal(t, http.StatusForbidden, request("Al", "Bearer "+token, "").Code)
	require.Zero(t, calls)
	w := request(" Al ", "Bearer "+token, config.ENV.AppSharedSecret)
	require.Equal(t, http.StatusOK, w.Code)
	require.JSONEq(t, `{"success":true,"data":{"users":[{"id":8,"name":"Alice","username":"alice"}]}}`, w.Body.String())
	w = request("none", "Bearer "+token, config.ENV.AppSharedSecret)
	require.JSONEq(t, `{"success":true,"data":{"users":[]}}`, w.Body.String())
	w = request("a", "Bearer "+token, config.ENV.AppSharedSecret)
	require.Equal(t, http.StatusBadRequest, w.Code)
	w = request("fail", "Bearer "+token, config.ENV.AppSharedSecret)
	require.Equal(t, http.StatusInternalServerError, w.Code)
	require.NotContains(t, w.Body.String(), "private database details")
	for i := 0; i < 16; i++ {
		require.Equal(t, http.StatusOK, request("Al", "Bearer "+token, config.ENV.AppSharedSecret).Code)
	}
	w = request("Al", "Bearer "+token, config.ENV.AppSharedSecret)
	require.Equal(t, http.StatusTooManyRequests, w.Code)
	require.NotEmpty(t, w.Header().Get("Retry-After"))
}

func TestSearchHandlerRejectsInvalidIdentity(t *testing.T) {
	for _, id := range []any{nil, "7", int64(0), int64(-1)} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/?q=ab", nil)
		if id != nil {
			c.Set("userId", id)
		}
		users.NewUserHandler(nil).SearchUsers(c)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	}
}
