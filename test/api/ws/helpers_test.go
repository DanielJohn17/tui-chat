package ws_test

import (
	"context"
	"sync"

	"github.com/DanielJohn17/tui-chat/internal/api/conversations"
	"github.com/stretchr/testify/mock"
)

type MockConvService struct {
	mock.Mock
}

func (m *MockConvService) GetContactUserIDs(ctx context.Context, userID int64) ([]int64, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]int64), args.Error(1)
}

type mockMarkReadCall struct {
	MessageID int64
	UserID    int64
	ConvID    int64
}

type mockPersister struct {
	mu            sync.Mutex
	markReadCalls []mockMarkReadCall
}

func (m *mockPersister) CreateMessage(
	ctx context.Context,
	convID, senderID int64,
	content string,
) (*conversations.CreateMessageResponseType, error) {
	return nil, nil
}

func (m *mockPersister) GetUnreadCount(
	ctx context.Context,
	userID, convID int64,
) (int, error) {
	return 0, nil
}

func (m *mockPersister) MarkAsRead(ctx context.Context, messageID, userID, convID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.markReadCalls = append(m.markReadCalls, mockMarkReadCall{
		MessageID: messageID,
		UserID:    userID,
		ConvID:    convID,
	})
}

func (m *mockPersister) getCalls() []mockMarkReadCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	cpy := make([]mockMarkReadCall, len(m.markReadCalls))
	copy(cpy, m.markReadCalls)
	return cpy
}
