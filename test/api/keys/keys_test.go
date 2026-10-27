package keys_test

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DanielJohn17/tui-chat/internal/api/apierrs"
	"github.com/DanielJohn17/tui-chat/internal/api/keys"
	"github.com/DanielJohn17/tui-chat/internal/api/types"
	"github.com/DanielJohn17/tui-chat/internal/api/users"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockKeyService struct {
	mock.Mock
}

func (m *MockKeyService) UploadPublicKey(ctx context.Context, userID int64, hexKey string) error {
	return m.Called(ctx, userID, hexKey).Error(0)
}

func (m *MockKeyService) GetPublicKey(ctx context.Context, userID int64) (*keys.PublicKeyResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*keys.PublicKeyResponse), args.Error(1)
}

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) SearchUsers(ctx context.Context, userID int64, query string) (*users.SearchUsersResponse, error) {
	return nil, nil
}
func (m *MockUserService) CreateUser(ctx context.Context, input users.CreateUserType) (*users.CreateUserResponseType, error) {
	return nil, nil
}
func (m *MockUserService) GetUserByUsername(ctx context.Context, username string) (*users.GetUserResponseType, error) {
	return nil, nil
}
func (m *MockUserService) GetUserByID(ctx context.Context, id int64) (*users.GetUserResponseType, error) {
	return nil, nil
}
func (m *MockUserService) DeleteUser(ctx context.Context, id int64) (*users.DeleteUserResponseType, error) {
	return nil, nil
}
func (m *MockUserService) GetUserByInviteToken(ctx context.Context, inviteToken string) (*users.PublicUser, error) {
	return nil, nil
}
func (m *MockUserService) UpdateProfile(ctx context.Context, userID int64, input users.UpdateUserProfileRequest) (*users.GetUserResponseType, error) {
	return nil, nil
}
func (m *MockUserService) RegenerateInviteToken(ctx context.Context, userID int64) (*users.GetUserResponseType, error) {
	return nil, nil
}
func (m *MockUserService) SetPublicKey(ctx context.Context, userID int64, pubKey []byte) error {
	return m.Called(ctx, userID, pubKey).Error(0)
}
func (m *MockUserService) GetPublicKey(ctx context.Context, userID int64) ([]byte, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func setupKeyTestRouter(h keys.KeyHandlerInt) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())
	return r
}

func TestKeyHandler_UploadPublicKey(t *testing.T) {
	validKey := hex.EncodeToString(make([]byte, 32))

	t.Run("success", func(t *testing.T) {
		mockSvc := new(MockKeyService)
		mockSvc.On("UploadPublicKey", mock.Anything, int64(42), validKey).Return(nil)

		handler := keys.NewKeyHandler(mockSvc)
		r := setupKeyTestRouter(handler)
		r.PUT("/keys/me", func(c *gin.Context) {
			c.Set("userId", int64(42))
			handler.UploadPublicKey(c)
		})

		body, _ := json.Marshal(keys.UploadPublicKeyRequest{PublicKey: validKey})
		req, _ := http.NewRequest(http.MethodPut, "/keys/me", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("unauthorized when userId is missing", func(t *testing.T) {
		mockSvc := new(MockKeyService)
		handler := keys.NewKeyHandler(mockSvc)
		r := setupKeyTestRouter(handler)
		r.PUT("/keys/me", handler.UploadPublicKey)

		body, _ := json.Marshal(keys.UploadPublicKeyRequest{PublicKey: validKey})
		req, _ := http.NewRequest(http.MethodPut, "/keys/me", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("bad request on invalid body", func(t *testing.T) {
		mockSvc := new(MockKeyService)
		handler := keys.NewKeyHandler(mockSvc)
		r := setupKeyTestRouter(handler)
		r.PUT("/keys/me", func(c *gin.Context) {
			c.Set("userId", int64(42))
			handler.UploadPublicKey(c)
		})

		req, _ := http.NewRequest(http.MethodPut, "/keys/me", bytes.NewBufferString("not-json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestKeyHandler_GetPublicKey(t *testing.T) {
	validKey := hex.EncodeToString(make([]byte, 32))

	t.Run("success", func(t *testing.T) {
		mockSvc := new(MockKeyService)
		mockSvc.On("GetPublicKey", mock.Anything, int64(42)).Return(&keys.PublicKeyResponse{
			UserID:    42,
			PublicKey: validKey,
		}, nil)

		handler := keys.NewKeyHandler(mockSvc)
		r := setupKeyTestRouter(handler)
		r.GET("/keys/:user_id", handler.GetPublicKey)

		req, _ := http.NewRequest(http.MethodGet, "/keys/42", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp types.APIResponse[keys.PublicKeyResponse]
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Equal(t, validKey, resp.Data.PublicKey)
	})

	t.Run("invalid user_id param", func(t *testing.T) {
		mockSvc := new(MockKeyService)
		handler := keys.NewKeyHandler(mockSvc)
		r := setupKeyTestRouter(handler)
		r.GET("/keys/:user_id", handler.GetPublicKey)

		req, _ := http.NewRequest(http.MethodGet, "/keys/invalid", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("user not found or has no key", func(t *testing.T) {
		mockSvc := new(MockKeyService)
		mockSvc.On("GetPublicKey", mock.Anything, int64(99)).Return(nil, apierrs.NewNotFoundError("public key not found"))

		handler := keys.NewKeyHandler(mockSvc)
		r := setupKeyTestRouter(handler)
		r.GET("/keys/:user_id", handler.GetPublicKey)

		req, _ := http.NewRequest(http.MethodGet, "/keys/99", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestKeyService_UploadPublicKey(t *testing.T) {
	t.Run("rejects non-hex string", func(t *testing.T) {
		mockUserSvc := new(MockUserService)
		svc := keys.NewKeyService(mockUserSvc)

		err := svc.UploadPublicKey(context.Background(), 1, "not-a-hex-string!!!")
		assert.Error(t, err)
	})

	t.Run("accepts valid hex and delegates", func(t *testing.T) {
		mockUserSvc := new(MockUserService)
		rawKey := make([]byte, 32)
		rawKey[0] = 0xAA
		hexKey := hex.EncodeToString(rawKey)

		mockUserSvc.On("SetPublicKey", mock.Anything, int64(1), rawKey).Return(nil)

		svc := keys.NewKeyService(mockUserSvc)
		err := svc.UploadPublicKey(context.Background(), 1, hexKey)
		assert.NoError(t, err)
		mockUserSvc.AssertExpectations(t)
	})
}

func TestKeyService_GetPublicKey(t *testing.T) {
	t.Run("returns hex-encoded public key", func(t *testing.T) {
		mockUserSvc := new(MockUserService)
		rawKey := make([]byte, 32)
		rawKey[0] = 0xBB
		hexKey := hex.EncodeToString(rawKey)

		mockUserSvc.On("GetPublicKey", mock.Anything, int64(1)).Return(rawKey, nil)

		svc := keys.NewKeyService(mockUserSvc)
		resp, err := svc.GetPublicKey(context.Background(), 1)
		require.NoError(t, err)
		assert.Equal(t, hexKey, resp.PublicKey)
		assert.Equal(t, int64(1), resp.UserID)
	})

	t.Run("propagates error when key not found", func(t *testing.T) {
		mockUserSvc := new(MockUserService)
		mockUserSvc.On("GetPublicKey", mock.Anything, int64(2)).Return(nil, apierrs.NewNotFoundError("public key not found"))

		svc := keys.NewKeyService(mockUserSvc)
		resp, err := svc.GetPublicKey(context.Background(), 2)
		assert.Nil(t, resp)
		assert.Error(t, err)
	})
}

type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) SearchUsers(ctx context.Context, userID int64, query string) ([]users.PublicUser, error) {
	return nil, nil
}
func (m *MockUserRepo) CreateUser(ctx context.Context, user users.CreateUserType) (*users.CreateUserResponseType, error) {
	return nil, nil
}
func (m *MockUserRepo) GetUserByUsername(ctx context.Context, username string) (*users.GetUserResponseType, error) {
	return nil, nil
}
func (m *MockUserRepo) GetUserByID(ctx context.Context, id int64) (*users.GetUserResponseType, error) {
	return nil, nil
}
func (m *MockUserRepo) DeleteUser(ctx context.Context, id int64) (*users.DeleteUserResponseType, error) {
	return nil, nil
}
func (m *MockUserRepo) GetUserByInviteToken(ctx context.Context, inviteToken string) (*users.GetUserResponseType, error) {
	return nil, nil
}
func (m *MockUserRepo) UpdateUserProfile(ctx context.Context, id int64, name, bio string, isSearchable bool) (*users.GetUserResponseType, error) {
	return nil, nil
}
func (m *MockUserRepo) RegenerateInviteToken(ctx context.Context, id int64, token string) (*users.GetUserResponseType, error) {
	return nil, nil
}
func (m *MockUserRepo) SetPublicKey(ctx context.Context, userID int64, pubKey []byte) error {
	return m.Called(ctx, userID, pubKey).Error(0)
}
func (m *MockUserRepo) GetPublicKey(ctx context.Context, userID int64) ([]byte, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func TestUserService_PublicKeyLogic(t *testing.T) {
	t.Run("SetPublicKey validates 32 bytes length", func(t *testing.T) {
		mockRepo := new(MockUserRepo)
		svc := users.NewUserService(mockRepo)

		errShort := svc.SetPublicKey(context.Background(), 1, []byte("too-short"))
		assert.Error(t, errShort)

		validKey := make([]byte, 32)
		mockRepo.On("SetPublicKey", mock.Anything, int64(1), validKey).Return(nil)
		errValid := svc.SetPublicKey(context.Background(), 1, validKey)
		assert.NoError(t, errValid)
		mockRepo.AssertExpectations(t)
	})

	t.Run("GetPublicKey returns NotFoundError when key is nil or empty", func(t *testing.T) {
		mockRepo := new(MockUserRepo)
		svc := users.NewUserService(mockRepo)

		// Simulating NULL in database scanned as nil []byte without error
		mockRepo.On("GetPublicKey", mock.Anything, int64(10)).Return(nil, nil)

		pk, err := svc.GetPublicKey(context.Background(), 10)
		assert.Nil(t, pk)
		require.Error(t, err)
		var apiErr *apierrs.APIError
		assert.True(t, errors.As(err, &apiErr))
		assert.Equal(t, http.StatusNotFound, apiErr.Code)
	})

	t.Run("GetPublicKey returns NotFoundError when repo returns error", func(t *testing.T) {
		mockRepo := new(MockUserRepo)
		svc := users.NewUserService(mockRepo)

		mockRepo.On("GetPublicKey", mock.Anything, int64(11)).Return(nil, errors.New("db error"))

		pk, err := svc.GetPublicKey(context.Background(), 11)
		assert.Nil(t, pk)
		require.Error(t, err)
	})

	t.Run("GetPublicKey returns key when 32 bytes present", func(t *testing.T) {
		mockRepo := new(MockUserRepo)
		svc := users.NewUserService(mockRepo)
		expected := make([]byte, 32)
		expected[0] = 0x42

		mockRepo.On("GetPublicKey", mock.Anything, int64(12)).Return(expected, nil)

		pk, err := svc.GetPublicKey(context.Background(), 12)
		require.NoError(t, err)
		assert.Equal(t, expected, pk)
	})
}
