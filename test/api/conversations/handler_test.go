package conversations_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DanielJohn17/tui-chat/internal/api/conversations"
	"github.com/DanielJohn17/tui-chat/internal/api/types"
	"github.com/DanielJohn17/tui-chat/internal/api/ws"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockConvService struct {
	mock.Mock
}

func (m *MockConvService) GetOrCreateDirectConversation(ctx context.Context, input conversations.GetOrCreateDirectConvType) ([]conversations.GetConvParticipantType, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]conversations.GetConvParticipantType), args.Error(1)
}

func (m *MockConvService) GetConvsByUserID(ctx context.Context, input int64) ([]conversations.GetConvParticipantType, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]conversations.GetConvParticipantType), args.Error(1)
}

func (m *MockConvService) GetBulkChatsByUserID(ctx context.Context, userID, limit int64) ([]conversations.BulkChatsResponseType, error) {
	args := m.Called(ctx, userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]conversations.BulkChatsResponseType), args.Error(1)
}

func (m *MockConvService) GetConvChats(ctx context.Context, convID int64, query types.URLQueryParams) []conversations.GetConvChatResponseType {
	args := m.Called(ctx, convID, query)
	return args.Get(0).([]conversations.GetConvChatResponseType)
}

func (m *MockConvService) CreateMessage(ctx context.Context, convID, senderID int64, content string) (*conversations.CreateMessageResponseType, error) {
	args := m.Called(ctx, convID, senderID, content)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*conversations.CreateMessageResponseType), args.Error(1)
}

func (m *MockConvService) WipeConversation(ctx context.Context, convID, userID int64) error {
	return m.Called(ctx, convID, userID).Error(0)
}

func (m *MockConvService) GetUnreadCount(ctx context.Context, userID, convID int64) (int, error) {
	args := m.Called(ctx, userID, convID)
	return args.Int(0), args.Error(1)
}

func (m *MockConvService) MarkAsRead(ctx context.Context, messageID, userID, convID int64) {
	m.Called(ctx, messageID, userID, convID)
}

type MockPresenceChecker struct {
	mock.Mock
}

func (m *MockPresenceChecker) GetOnlineStatus(ctx context.Context, userIDs []int64) (map[int64]bool, error) {
	args := m.Called(ctx, userIDs)
	return args.Get(0).(map[int64]bool), args.Error(1)
}

func TestGetConvsByUserID_WithPresence(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockConvService)
	mockPresence := new(MockPresenceChecker)

	convsData := []conversations.GetConvParticipantType{
		{ConvID: 1, UserID: 102, Name: "Bob Martin", Username: "bob", Online: false},
		{ConvID: 2, UserID: 103, Name: "Sarah Connor", Username: "sarah", Online: false},
	}

	mockService.On("GetConvsByUserID", mock.Anything, int64(101)).Return(convsData, nil)
	mockPresence.On("GetOnlineStatus", mock.Anything, []int64{102, 103}).Return(map[int64]bool{
		102: true,
		103: false,
	}, nil)

	handler := conversations.NewConvHandler(mockService, mockPresence)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userId", int64(101))
		c.Next()
	})
	r.GET("/conversations", handler.GetConvsByUserID)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/conversations", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var res struct {
		Success bool                                   `json:"success"`
		Data    []conversations.GetConvParticipantType `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)

	require.True(t, res.Success)
	require.Len(t, res.Data, 2)
	assert.True(t, res.Data[0].Online, "User 102 (Bob) should be online")
	assert.False(t, res.Data[1].Online, "User 103 (Sarah) should be offline")
}

func TestHub_GetOnlineStatus(t *testing.T) {
	hub := ws.NewHub()
	go hub.Run()

	// 1. Initially both users are offline
	statusMap, err := hub.GetOnlineStatus(context.Background(), []int64{102, 103})
	require.NoError(t, err)
	assert.False(t, statusMap[102])
	assert.False(t, statusMap[103])

	// 2. Register user 102
	client102 := &ws.Client{
		UserID: 102,
		Send:   make(chan []byte, 10),
		Hub:    hub,
	}
	hub.Register <- client102

	// Allow goroutine to process registration
	require.Eventually(t, func() bool {
		m, err := hub.GetOnlineStatus(context.Background(), []int64{102})
		return err == nil && m[102]
	}, 1*time.Second, 10*time.Millisecond)

	// 3. User 103 is still offline
	statusMap2, err := hub.GetOnlineStatus(context.Background(), []int64{102, 103})
	require.NoError(t, err)
	assert.True(t, statusMap2[102])
	assert.False(t, statusMap2[103])

	// 4. Unregister user 102
	hub.UnRegister <- client102
	require.Eventually(t, func() bool {
		m, err := hub.GetOnlineStatus(context.Background(), []int64{102})
		return err == nil && !m[102]
	}, 1*time.Second, 10*time.Millisecond)
}
