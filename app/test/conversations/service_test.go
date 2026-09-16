package conversations_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DanielJohn17/tui-chat/app/internal/api/conversations"
	apierrors "github.com/DanielJohn17/tui-chat/app/internal/api/errors"
	"github.com/DanielJohn17/tui-chat/app/internal/api/types"
	"github.com/DanielJohn17/tui-chat/app/internal/api/users"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockConvRepo mocks conversations.ConvRepositoryInt using testify/mock.
type MockConvRepo struct {
	mock.Mock
}

func (m *MockConvRepo) IsUserInConversation(ctx context.Context, convID, userID int64) bool {
	return m.Called(ctx, convID, userID).Bool(0)
}

func (m *MockConvRepo) MarkConvForDeleting(ctx context.Context, convID int64) error {
	return m.Called(ctx, convID).Error(0)
}

func (m *MockConvRepo) DeleteMessagesAsBatch(ctx context.Context, convID, batchSize int64) (int64, error) {
	args := m.Called(ctx, convID, batchSize)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockConvRepo) DeleteConvByID(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}

func (m *MockConvRepo) GetOrCreateDirectConversation(ctx context.Context, input conversations.GetOrCreateDirectConvType) ([]conversations.GetConvParticipantType, error) {
	args := m.Called(ctx, input)
	if res := args.Get(0); res != nil {
		return res.([]conversations.GetConvParticipantType), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockConvRepo) GetConvsByUserID(ctx context.Context, input int64) []conversations.GetConvParticipantType {
	args := m.Called(ctx, input)
	if res := args.Get(0); res != nil {
		return res.([]conversations.GetConvParticipantType)
	}
	return nil
}

func (m *MockConvRepo) GetConvChats(ctx context.Context, convID int64, input types.URLQueryParams) []conversations.GetConvChatResponseType {
	args := m.Called(ctx, convID, input)
	if res := args.Get(0); res != nil {
		return res.([]conversations.GetConvChatResponseType)
	}
	return nil
}

func (m *MockConvRepo) GetConvChatsPaginated(ctx context.Context, convID int64, input types.URLQueryParams) []conversations.GetConvChatResponseType {
	args := m.Called(ctx, convID, input)
	if res := args.Get(0); res != nil {
		return res.([]conversations.GetConvChatResponseType)
	}
	return nil
}

func (m *MockConvRepo) CreateMessage(ctx context.Context, convID, senderID int64, content string) (*conversations.CreateMessageResponseType, error) {
	args := m.Called(ctx, convID, senderID, content)
	if res := args.Get(0); res != nil {
		return res.(*conversations.CreateMessageResponseType), args.Error(1)
	}
	return nil, args.Error(1)
}

// MockUserService mocks users.UserServiceInt using testify/mock.
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(ctx context.Context, input users.CreateUserType) (*users.CreateUserResponseType, error) {
	return nil, m.Called(ctx, input).Error(1)
}

func (m *MockUserService) GetUserByUsername(ctx context.Context, input string) (*users.GetUserResponseType, error) {
	args := m.Called(ctx, input)
	if res := args.Get(0); res != nil {
		return res.(*users.GetUserResponseType), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserService) GetUserByID(ctx context.Context, input int64) (*users.GetUserResponseType, error) {
	args := m.Called(ctx, input)
	if res := args.Get(0); res != nil {
		return res.(*users.GetUserResponseType), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserService) DeleteUser(ctx context.Context, input int64) (*users.DeleteUserResponseType, error) {
	return nil, m.Called(ctx, input).Error(1)
}

// -----------------------------------------------------------------------------
// WipeConversation Tests
// -----------------------------------------------------------------------------

func TestWipeConversation_UserNotParticipant(t *testing.T) {
	repo := new(MockConvRepo)
	repo.On("IsUserInConversation", mock.Anything, int64(10), int64(99)).Return(false)

	s := conversations.NewConvService(repo, new(MockUserService))
	err := s.WipeConversation(context.Background(), 10, 99)

	var apiErr *apierrors.APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 403, apiErr.Code)
	repo.AssertNotCalled(t, "DeleteMessagesAsBatch", mock.Anything, mock.Anything, mock.Anything)
	repo.AssertNotCalled(t, "DeleteConvByID", mock.Anything, mock.Anything)
}

func TestWipeConversation_MarkDeletingFailure(t *testing.T) {
	repo := new(MockConvRepo)
	repo.On("IsUserInConversation", mock.Anything, int64(10), int64(99)).Return(true)
	repo.On("MarkConvForDeleting", mock.Anything, int64(10)).Return(errors.New("db error"))

	s := conversations.NewConvService(repo, new(MockUserService))
	err := s.WipeConversation(context.Background(), 10, 99)

	var apiErr *apierrors.APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 500, apiErr.Code)
	repo.AssertNotCalled(t, "DeleteMessagesAsBatch", mock.Anything, mock.Anything, mock.Anything)
	repo.AssertNotCalled(t, "DeleteConvByID", mock.Anything, mock.Anything)
}

func TestWipeConversation_AsyncBatchChunkingSuccess(t *testing.T) {
	repo := new(MockConvRepo)
	repo.On("IsUserInConversation", mock.Anything, int64(42), int64(1)).Return(true)
	repo.On("MarkConvForDeleting", mock.Anything, int64(42)).Return(nil)

	deletedDone := make(chan struct{})
	repo.On("DeleteMessagesAsBatch", mock.Anything, int64(42), int64(500)).Return(int64(500), nil).Once()
	repo.On("DeleteMessagesAsBatch", mock.Anything, int64(42), int64(500)).Return(int64(500), nil).Once()
	repo.On("DeleteMessagesAsBatch", mock.Anything, int64(42), int64(500)).Return(int64(120), nil).Once()
	repo.On("DeleteConvByID", mock.Anything, int64(42)).Return(nil).Run(func(args mock.Arguments) {
		close(deletedDone)
	}).Once()

	s := conversations.NewConvServiceWithOptions(repo, new(MockUserService), 5, 500, 1*time.Millisecond)
	require.NoError(t, s.WipeConversation(context.Background(), 42, 1))

	select {
	case <-deletedDone:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for conversation deletion")
	}

	repo.AssertExpectations(t)
}

func TestWipeConversation_BatchDeleteErrorAbortsBeforeDeleteConv(t *testing.T) {
	batchAttempted := make(chan struct{})
	repo := new(MockConvRepo)
	repo.On("IsUserInConversation", mock.Anything, int64(55), int64(1)).Return(true)
	repo.On("MarkConvForDeleting", mock.Anything, int64(55)).Return(nil)
	repo.On("DeleteMessagesAsBatch", mock.Anything, int64(55), int64(500)).Return(int64(0), errors.New("db timeout")).Run(func(args mock.Arguments) {
		close(batchAttempted)
	})

	s := conversations.NewConvService(repo, new(MockUserService))
	require.NoError(t, s.WipeConversation(context.Background(), 55, 1))

	select {
	case <-batchAttempted:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for batch attempt")
	}

	time.Sleep(50 * time.Millisecond)
	repo.AssertNotCalled(t, "DeleteConvByID", mock.Anything, mock.Anything)
}

func TestWipeConversation_PanicRecovery(t *testing.T) {
	panicAttempted := make(chan struct{})
	secondWipeDone := make(chan struct{})

	repo := new(MockConvRepo)
	repo.On("IsUserInConversation", mock.Anything, mock.Anything, mock.Anything).Return(true)
	repo.On("MarkConvForDeleting", mock.Anything, mock.Anything).Return(nil)

	repo.On("DeleteMessagesAsBatch", mock.Anything, int64(77), int64(500)).Run(func(args mock.Arguments) {
		close(panicAttempted)
		panic("simulated worker panic")
	})
	repo.On("DeleteMessagesAsBatch", mock.Anything, int64(78), int64(500)).Return(int64(0), nil)
	repo.On("DeleteConvByID", mock.Anything, int64(78)).Return(nil).Run(func(args mock.Arguments) {
		close(secondWipeDone)
	})

	s := conversations.NewConvServiceWithOptions(repo, new(MockUserService), 1, 500, 1*time.Millisecond)
	require.NoError(t, s.WipeConversation(context.Background(), 77, 1))

	select {
	case <-panicAttempted:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for panic")
	}

	time.Sleep(50 * time.Millisecond)

	// Since workerSlots = 1, if slot leaked, second wipe would block
	require.NoError(t, s.WipeConversation(context.Background(), 78, 1))

	select {
	case <-secondWipeDone:
	case <-time.After(2 * time.Second):
		t.Fatal("semaphore slot was leaked after panic")
	}
}

func TestWipeConversation_ParentContextCancellationDoesNotAbortWorker(t *testing.T) {
	deletedDone := make(chan struct{})
	repo := new(MockConvRepo)
	repo.On("IsUserInConversation", mock.Anything, int64(88), int64(1)).Return(true)
	repo.On("MarkConvForDeleting", mock.Anything, int64(88)).Return(nil)
	repo.On("DeleteMessagesAsBatch", mock.Anything, int64(88), int64(500)).Return(int64(0), nil)
	repo.On("DeleteConvByID", mock.Anything, int64(88)).Return(nil).Run(func(args mock.Arguments) {
		close(deletedDone)
	})

	s := conversations.NewConvServiceWithOptions(repo, new(MockUserService), 5, 500, 1*time.Millisecond)

	parentCtx, parentCancel := context.WithCancel(context.Background())
	require.NoError(t, s.WipeConversation(parentCtx, 88, 1))
	parentCancel()

	select {
	case <-deletedDone:
	case <-time.After(2 * time.Second):
		t.Fatal("worker aborted when parent context was cancelled")
	}
}

// -----------------------------------------------------------------------------
// Other ConvService Methods Tests
// -----------------------------------------------------------------------------

func TestGetOrCreateDirectConversation_Success(t *testing.T) {
	expected := []conversations.GetConvParticipantType{
		{ConvID: 1, UserID: 10, Name: "User 10", Username: "user10"},
		{ConvID: 1, UserID: 20, Name: "User 20", Username: "user20"},
	}
	input := conversations.GetOrCreateDirectConvType{UserIDOne: 10, UserIDTwo: 20}

	uService := new(MockUserService)
	uService.On("GetUserByID", mock.Anything, int64(10)).Return(&users.GetUserResponseType{ID: 10}, nil)
	uService.On("GetUserByID", mock.Anything, int64(20)).Return(&users.GetUserResponseType{ID: 20}, nil)

	repo := new(MockConvRepo)
	repo.On("GetOrCreateDirectConversation", mock.Anything, input).Return(expected, nil)

	s := conversations.NewConvService(repo, uService)
	participants, err := s.GetOrCreateDirectConversation(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, expected, participants)
}

func TestGetOrCreateDirectConversation_UserNotFound(t *testing.T) {
	uService := new(MockUserService)
	uService.On("GetUserByID", mock.Anything, int64(10)).Return(&users.GetUserResponseType{ID: 10}, nil)
	uService.On("GetUserByID", mock.Anything, int64(999)).Return(nil, errors.New("user not found"))

	s := conversations.NewConvService(new(MockConvRepo), uService)
	_, err := s.GetOrCreateDirectConversation(context.Background(), conversations.GetOrCreateDirectConvType{
		UserIDOne: 10,
		UserIDTwo: 999,
	})

	var apiErr *apierrors.APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 404, apiErr.Code)
}

func TestGetConvsByUserID(t *testing.T) {
	expected := []conversations.GetConvParticipantType{
		{ConvID: 1, UserID: 2, Name: "Alice", Username: "alice"},
	}
	uService := new(MockUserService)
	uService.On("GetUserByID", mock.Anything, int64(1)).Return(&users.GetUserResponseType{ID: 1}, nil)

	repo := new(MockConvRepo)
	repo.On("GetConvsByUserID", mock.Anything, int64(1)).Return(expected)

	s := conversations.NewConvService(repo, uService)
	convs, err := s.GetConvsByUserID(context.Background(), 1)

	require.NoError(t, err)
	assert.Equal(t, expected, convs)
}

func TestGetConvChats_PaginationBranches(t *testing.T) {
	repo := new(MockConvRepo)
	unpaginatedResp := []conversations.GetConvChatResponseType{{ID: 1, Content: "First"}}
	paginatedResp := []conversations.GetConvChatResponseType{{ID: 2, Content: "Paginated"}}

	emptyQuery := types.URLQueryParams{}
	cursorQuery := types.URLQueryParams{CursorID: 1, CursorTime: time.Now()}

	repo.On("GetConvChats", mock.Anything, int64(1), emptyQuery).Return(unpaginatedResp)
	repo.On("GetConvChatsPaginated", mock.Anything, int64(1), cursorQuery).Return(paginatedResp)

	s := conversations.NewConvService(repo, new(MockUserService))

	assert.Equal(t, unpaginatedResp, s.GetConvChats(context.Background(), 1, emptyQuery))
	assert.Equal(t, paginatedResp, s.GetConvChats(context.Background(), 1, cursorQuery))
}

func TestCreateMessage(t *testing.T) {
	expected := &conversations.CreateMessageResponseType{
		ID: 101, ConvID: 1, SenderID: 2, Content: "hello world",
	}
	repo := new(MockConvRepo)
	repo.On("CreateMessage", mock.Anything, int64(1), int64(2), "hello world").Return(expected, nil)

	s := conversations.NewConvService(repo, new(MockUserService))
	msg, err := s.CreateMessage(context.Background(), 1, 2, "hello world")

	require.NoError(t, err)
	assert.Equal(t, expected, msg)
}
