package conversations_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DanielJohn17/tui-chat/app/internal/api/conversations"
	apierrors "github.com/DanielJohn17/tui-chat/app/internal/api/errors"
	"github.com/DanielJohn17/tui-chat/app/internal/api/types"
	"github.com/DanielJohn17/tui-chat/app/internal/api/users"
)

// mockConvRepo implements conversations.ConvRepositoryInt for unit testing.
type mockConvRepo struct {
	mu sync.Mutex

	isUserInConversationFn  func(ctx context.Context, convID, userID int64) bool
	markConvForDeletingFn   func(ctx context.Context, convID int64) error
	deleteMessagesAsBatchFn func(ctx context.Context, convID, batchSize int64) (int64, error)
	deleteConvByIDFn        func(ctx context.Context, id int64) error
	getOrCreateDirectConvFn func(ctx context.Context, input conversations.GetOrCreateDirectConvType) ([]conversations.GetConvParticipantType, error)
	getConvsByUserIDFn      func(ctx context.Context, input int64) []conversations.GetConvParticipantType
	getConvChatsFn          func(ctx context.Context, convID int64, input types.URLQueryParams) []conversations.GetConvChatResponseType
	getConvChatsPaginatedFn func(ctx context.Context, convID int64, input types.URLQueryParams) []conversations.GetConvChatResponseType
	createMessageFn         func(ctx context.Context, convID, senderID int64, content string) (*conversations.CreateMessageResponseType, error)

	batchCalls  int64
	deleteCalls int64
}

func (m *mockConvRepo) IsUserInConversation(ctx context.Context, convID, userID int64) bool {
	if m.isUserInConversationFn != nil {
		return m.isUserInConversationFn(ctx, convID, userID)
	}
	return true
}

func (m *mockConvRepo) MarkConvForDeleting(ctx context.Context, convID int64) error {
	if m.markConvForDeletingFn != nil {
		return m.markConvForDeletingFn(ctx, convID)
	}
	return nil
}

func (m *mockConvRepo) DeleteMessagesAsBatch(ctx context.Context, convID, batchSize int64) (int64, error) {
	atomic.AddInt64(&m.batchCalls, 1)
	if m.deleteMessagesAsBatchFn != nil {
		return m.deleteMessagesAsBatchFn(ctx, convID, batchSize)
	}
	return 0, nil
}

func (m *mockConvRepo) DeleteConvByID(ctx context.Context, id int64) error {
	atomic.AddInt64(&m.deleteCalls, 1)
	if m.deleteConvByIDFn != nil {
		return m.deleteConvByIDFn(ctx, id)
	}
	return nil
}

func (m *mockConvRepo) GetOrCreateDirectConversation(ctx context.Context, input conversations.GetOrCreateDirectConvType) ([]conversations.GetConvParticipantType, error) {
	if m.getOrCreateDirectConvFn != nil {
		return m.getOrCreateDirectConvFn(ctx, input)
	}
	return nil, nil
}

func (m *mockConvRepo) GetConvsByUserID(ctx context.Context, input int64) []conversations.GetConvParticipantType {
	if m.getConvsByUserIDFn != nil {
		return m.getConvsByUserIDFn(ctx, input)
	}
	return nil
}

func (m *mockConvRepo) GetConvChats(ctx context.Context, convID int64, input types.URLQueryParams) []conversations.GetConvChatResponseType {
	if m.getConvChatsFn != nil {
		return m.getConvChatsFn(ctx, convID, input)
	}
	return nil
}

func (m *mockConvRepo) GetConvChatsPaginated(ctx context.Context, convID int64, input types.URLQueryParams) []conversations.GetConvChatResponseType {
	if m.getConvChatsPaginatedFn != nil {
		return m.getConvChatsPaginatedFn(ctx, convID, input)
	}
	return nil
}

func (m *mockConvRepo) CreateMessage(ctx context.Context, convID, senderID int64, content string) (*conversations.CreateMessageResponseType, error) {
	if m.createMessageFn != nil {
		return m.createMessageFn(ctx, convID, senderID, content)
	}
	return nil, nil
}

// mockUserService implements users.UserServiceInt for testing.
type mockUserService struct {
	getUserByIDFn func(ctx context.Context, id int64) (*users.GetUserResponseType, error)
}

func (m *mockUserService) CreateUser(ctx context.Context, input users.CreateUserType) (*users.CreateUserResponseType, error) {
	return nil, nil
}

func (m *mockUserService) GetUserByUsername(ctx context.Context, input string) (*users.GetUserResponseType, error) {
	return nil, nil
}

func (m *mockUserService) GetUserByID(ctx context.Context, input int64) (*users.GetUserResponseType, error) {
	if m.getUserByIDFn != nil {
		return m.getUserByIDFn(ctx, input)
	}
	return &users.GetUserResponseType{ID: input, Name: "Test User", Username: "testuser"}, nil
}

func (m *mockUserService) DeleteUser(ctx context.Context, input int64) (*users.DeleteUserResponseType, error) {
	return nil, nil
}

// -----------------------------------------------------------------------------
// WipeConversation Tests
// -----------------------------------------------------------------------------

func TestWipeConversation_UserNotParticipant(t *testing.T) {
	repo := &mockConvRepo{
		isUserInConversationFn: func(ctx context.Context, convID, userID int64) bool {
			return false
		},
	}
	uService := &mockUserService{}
	s := conversations.NewConvService(repo, uService)

	err := s.WipeConversation(context.Background(), 10, 99)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var apiErr *apierrors.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *apierrors.APIError, got %T", err)
	}

	if apiErr.Code != 403 {
		t.Fatalf("expected HTTP 403 Forbidden, got %d", apiErr.Code)
	}

	if atomic.LoadInt64(&repo.batchCalls) != 0 {
		t.Errorf("expected 0 batch calls, got %d", repo.batchCalls)
	}
	if atomic.LoadInt64(&repo.deleteCalls) != 0 {
		t.Errorf("expected 0 delete calls, got %d", repo.deleteCalls)
	}
}

func TestWipeConversation_MarkDeletingFailure(t *testing.T) {
	repo := &mockConvRepo{
		isUserInConversationFn: func(ctx context.Context, convID, userID int64) bool {
			return true
		},
		markConvForDeletingFn: func(ctx context.Context, convID int64) error {
			return errors.New("db error marking conversation")
		},
	}
	uService := &mockUserService{}
	s := conversations.NewConvService(repo, uService)

	err := s.WipeConversation(context.Background(), 10, 99)
	if err == nil {
		t.Fatal("expected error from MarkConvForDeleting, got nil")
	}

	var apiErr *apierrors.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *apierrors.APIError, got %T", err)
	}
	if apiErr.Code != 500 {
		t.Fatalf("expected HTTP 500 InternalServerError, got %d", apiErr.Code)
	}

	if atomic.LoadInt64(&repo.batchCalls) != 0 {
		t.Errorf("expected 0 batch calls, got %d", repo.batchCalls)
	}
	if atomic.LoadInt64(&repo.deleteCalls) != 0 {
		t.Errorf("expected 0 delete calls, got %d", repo.deleteCalls)
	}
}

func TestWipeConversation_AsyncBatchChunkingSuccess(t *testing.T) {
	deletedDone := make(chan struct{})
	var batchStep int64

	repo := &mockConvRepo{
		isUserInConversationFn: func(ctx context.Context, convID, userID int64) bool {
			return true
		},
		markConvForDeletingFn: func(ctx context.Context, convID int64) error {
			return nil
		},
		deleteMessagesAsBatchFn: func(ctx context.Context, convID, batchSize int64) (int64, error) {
			step := atomic.AddInt64(&batchStep, 1)
			if step == 1 {
				return 500, nil // Full batch 1
			}
			if step == 2 {
				return 500, nil // Full batch 2
			}
			if step == 3 {
				return 120, nil // Last partial batch (< 500)
			}
			return 0, nil
		},
		deleteConvByIDFn: func(ctx context.Context, id int64) error {
			if id != 42 {
				t.Errorf("expected convID 42, got %d", id)
			}
			close(deletedDone)
			return nil
		},
	}

	uService := &mockUserService{}
	s := conversations.NewConvServiceWithOptions(repo, uService, 5, 500, 1*time.Millisecond)

	// Call WipeConversation (should return nil immediately)
	err := s.WipeConversation(context.Background(), 42, 1)
	if err != nil {
		t.Fatalf("WipeConversation failed: %v", err)
	}

	// Wait for background goroutine to finish deleting
	select {
	case <-deletedDone:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for conversation deletion to finish")
	}

	if calls := atomic.LoadInt64(&repo.batchCalls); calls != 3 {
		t.Errorf("expected 3 batch delete calls, got %d", calls)
	}
	if calls := atomic.LoadInt64(&repo.deleteCalls); calls != 1 {
		t.Errorf("expected 1 delete conversation call, got %d", calls)
	}
}

func TestWipeConversation_BatchDeleteErrorAbortsBeforeDeleteConv(t *testing.T) {
	batchAttempted := make(chan struct{})

	repo := &mockConvRepo{
		isUserInConversationFn: func(ctx context.Context, convID, userID int64) bool {
			return true
		},
		markConvForDeletingFn: func(ctx context.Context, convID int64) error {
			return nil
		},
		deleteMessagesAsBatchFn: func(ctx context.Context, convID, batchSize int64) (int64, error) {
			close(batchAttempted)
			return 0, errors.New("simulated database timeout")
		},
		deleteConvByIDFn: func(ctx context.Context, id int64) error {
			t.Fatal("DeleteConvByID must NEVER be called when batch deletion fails!")
			return nil
		},
	}

	uService := &mockUserService{}
	s := conversations.NewConvService(repo, uService)

	err := s.WipeConversation(context.Background(), 55, 1)
	if err != nil {
		t.Fatalf("WipeConversation failed: %v", err)
	}

	select {
	case <-batchAttempted:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for batch attempt")
	}

	// Allow goroutine time to exit cleanly
	time.Sleep(50 * time.Millisecond)

	if calls := atomic.LoadInt64(&repo.deleteCalls); calls != 0 {
		t.Errorf("expected 0 deleteCalls on batch error, got %d", calls)
	}
}

func TestWipeConversation_PanicRecovery(t *testing.T) {
	panicAttempted := make(chan struct{})
	secondWipeDone := make(chan struct{})

	repo := &mockConvRepo{
		isUserInConversationFn: func(ctx context.Context, convID, userID int64) bool {
			return true
		},
		markConvForDeletingFn: func(ctx context.Context, convID int64) error {
			return nil
		},
		deleteMessagesAsBatchFn: func(ctx context.Context, convID, batchSize int64) (int64, error) {
			if convID == 77 {
				close(panicAttempted)
				panic("simulated critical runtime panic in worker")
			}
			return 0, nil
		},
		deleteConvByIDFn: func(ctx context.Context, id int64) error {
			if id == 78 {
				close(secondWipeDone)
			}
			return nil
		},
	}

	uService := &mockUserService{}
	// Use 1 worker slot to strictly test semaphore release on panic
	s := conversations.NewConvServiceWithOptions(repo, uService, 1, 500, 1*time.Millisecond)

	err := s.WipeConversation(context.Background(), 77, 1)
	if err != nil {
		t.Fatalf("WipeConversation returned error on initiation: %v", err)
	}

	select {
	case <-panicAttempted:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for panic to occur")
	}

	// Wait for recover() to run
	time.Sleep(50 * time.Millisecond)

	// Since workerSlots = 1, if the slot leaked, the next wipe cannot execute
	err = s.WipeConversation(context.Background(), 78, 1)
	if err != nil {
		t.Fatalf("second WipeConversation failed: %v", err)
	}

	select {
	case <-secondWipeDone:
		// Successfully executed next wipe, proving semaphore token was released on panic
	case <-time.After(2 * time.Second):
		t.Fatal("semaphore slot was leaked after panic; second wipe blocked")
	}
}

func TestWipeConversation_ParentContextCancellationDoesNotAbortWorker(t *testing.T) {
	deletedDone := make(chan struct{})

	repo := &mockConvRepo{
		isUserInConversationFn: func(ctx context.Context, convID, userID int64) bool {
			return true
		},
		markConvForDeletingFn: func(ctx context.Context, convID int64) error {
			return nil
		},
		deleteMessagesAsBatchFn: func(ctx context.Context, convID, batchSize int64) (int64, error) {
			return 0, nil // 0 messages deleted = finished
		},
		deleteConvByIDFn: func(ctx context.Context, id int64) error {
			close(deletedDone)
			return nil
		},
	}

	uService := &mockUserService{}
	s := conversations.NewConvServiceWithOptions(repo, uService, 5, 500, 1*time.Millisecond)

	parentCtx, parentCancel := context.WithCancel(context.Background())

	err := s.WipeConversation(parentCtx, 88, 1)
	if err != nil {
		t.Fatalf("WipeConversation failed: %v", err)
	}

	// Simulate HTTP response ending and cancelling parent context immediately
	parentCancel()

	select {
	case <-deletedDone:
		// Worker successfully completed despite parent context cancellation
	case <-time.After(2 * time.Second):
		t.Fatal("worker was aborted when parent context cancelled; WithoutCancel failed")
	}
}

// -----------------------------------------------------------------------------
// Other ConvService Methods Tests
// -----------------------------------------------------------------------------

func TestGetOrCreateDirectConversation_Success(t *testing.T) {
	repo := &mockConvRepo{
		getOrCreateDirectConvFn: func(ctx context.Context, input conversations.GetOrCreateDirectConvType) ([]conversations.GetConvParticipantType, error) {
			return []conversations.GetConvParticipantType{
				{ConvID: 1, UserID: 10, Name: "User 10", Username: "user10"},
				{ConvID: 1, UserID: 20, Name: "User 20", Username: "user20"},
			}, nil
		},
	}
	uService := &mockUserService{}
	s := conversations.NewConvService(repo, uService)

	participants, err := s.GetOrCreateDirectConversation(context.Background(), conversations.GetOrCreateDirectConvType{
		UserIDOne: 10,
		UserIDTwo: 20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(participants) != 2 {
		t.Fatalf("expected 2 participants, got %d", len(participants))
	}
}

func TestGetOrCreateDirectConversation_UserNotFound(t *testing.T) {
	repo := &mockConvRepo{}
	uService := &mockUserService{
		getUserByIDFn: func(ctx context.Context, id int64) (*users.GetUserResponseType, error) {
			if id == 999 {
				return nil, fmt.Errorf("user not found")
			}
			return &users.GetUserResponseType{ID: id}, nil
		},
	}
	s := conversations.NewConvService(repo, uService)

	_, err := s.GetOrCreateDirectConversation(context.Background(), conversations.GetOrCreateDirectConvType{
		UserIDOne: 10,
		UserIDTwo: 999,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *apierrors.APIError
	if !errors.As(err, &apiErr) || apiErr.Code != 404 {
		t.Fatalf("expected 404 NotFound error, got: %v", err)
	}
}

func TestGetConvsByUserID(t *testing.T) {
	repo := &mockConvRepo{
		getConvsByUserIDFn: func(ctx context.Context, input int64) []conversations.GetConvParticipantType {
			return []conversations.GetConvParticipantType{
				{ConvID: 1, UserID: 2, Name: "Alice", Username: "alice"},
			}
		},
	}
	uService := &mockUserService{}
	s := conversations.NewConvService(repo, uService)

	convs, err := s.GetConvsByUserID(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(convs) != 1 || convs[0].ConvID != 1 {
		t.Fatalf("expected 1 conv with ID 1, got %v", convs)
	}
}

func TestGetConvChats_PaginationBranches(t *testing.T) {
	unpaginatedCalled := false
	paginatedCalled := false

	repo := &mockConvRepo{
		getConvChatsFn: func(ctx context.Context, convID int64, input types.URLQueryParams) []conversations.GetConvChatResponseType {
			unpaginatedCalled = true
			return []conversations.GetConvChatResponseType{{ID: 1, Content: "First"}}
		},
		getConvChatsPaginatedFn: func(ctx context.Context, convID int64, input types.URLQueryParams) []conversations.GetConvChatResponseType {
			paginatedCalled = true
			return []conversations.GetConvChatResponseType{{ID: 2, Content: "Paginated"}}
		},
	}
	s := conversations.NewConvService(repo, &mockUserService{})

	// 1. Unpaginated (no cursor)
	res1 := s.GetConvChats(context.Background(), 1, types.URLQueryParams{})
	if !unpaginatedCalled || len(res1) != 1 || res1[0].Content != "First" {
		t.Fatalf("expected unpaginated path, got %v", res1)
	}

	// 2. Paginated (has cursor ID and Time)
	res2 := s.GetConvChats(context.Background(), 1, types.URLQueryParams{
		CursorID:   1,
		CursorTime: time.Now(),
	})
	if !paginatedCalled || len(res2) != 1 || res2[0].Content != "Paginated" {
		t.Fatalf("expected paginated path, got %v", res2)
	}
}

func TestCreateMessage(t *testing.T) {
	repo := &mockConvRepo{
		createMessageFn: func(ctx context.Context, convID, senderID int64, content string) (*conversations.CreateMessageResponseType, error) {
			return &conversations.CreateMessageResponseType{
				ID:       101,
				ConvID:   convID,
				SenderID: senderID,
				Content:  content,
			}, nil
		},
	}
	s := conversations.NewConvService(repo, &mockUserService{})

	msg, err := s.CreateMessage(context.Background(), 1, 2, "hello world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.ID != 101 || msg.Content != "hello world" {
		t.Fatalf("unexpected message: %+v", msg)
	}
}
