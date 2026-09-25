package conversations_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DanielJohn17/tui-chat/internal/api/conversations"
	"github.com/DanielJohn17/tui-chat/internal/api/types"
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

func (m *MockConvService) GetContactUserIDs(ctx context.Context, userID int64) ([]int64, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]int64), args.Error(1)
}

func TestConvHandler_GetConvsByUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockConvService)

	convsData := []conversations.GetConvParticipantType{
		{ConvID: 1, UserID: 102, Name: "Bob Martin", Username: "bob"},
		{ConvID: 2, UserID: 103, Name: "Sarah Connor", Username: "sarah"},
	}

	mockService.On("GetConvsByUserID", mock.Anything, int64(101)).Return(convsData, nil)

	handler := conversations.NewConvHandler(mockService)

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
	assert.Equal(t, int64(102), res.Data[0].UserID)
	assert.Equal(t, int64(103), res.Data[1].UserID)
}
