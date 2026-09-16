package conversations

import (
	"context"
	"log/slog"
	"runtime/debug"
	"sync"
	"time"

	"github.com/DanielJohn17/tui-chat/app/internal/api/errors"
	"github.com/DanielJohn17/tui-chat/app/internal/api/types"
	"github.com/DanielJohn17/tui-chat/app/internal/api/users"
)

type ConvServiceInt interface {
	GetOrCreateDirectConversation(
		ctx context.Context,
		input GetOrCreateDirectConvType,
	) ([]GetConvParticipantType, error)

	GetConvsByUserID(ctx context.Context, input int64) ([]GetConvParticipantType, error)

	GetConvChats(
		ctx context.Context,
		convID int64,
		query types.URLQueryParams,
	) []GetConvChatResponseType

	CreateMessage(
		ctx context.Context,
		convID, senderID int64, content string,
	) (*CreateMessageResponseType, error)

	WipeConversation(ctx context.Context, convID, userID int64) error
}

type ConvService struct {
	r          ConvRepositoryInt
	u          users.UserServiceInt
	sem        chan struct{}
	batchSize  int64
	sleepDelay time.Duration
}

func NewConvService(r ConvRepositoryInt, u users.UserServiceInt) *ConvService {
	return NewConvServiceWithOptions(r, u, 5, 500, time.Millisecond*15)
}

// NewConvServiceWithOptions used for testing purpose to configure zero-sleep and custom batch sizes and worker limits.
func NewConvServiceWithOptions(
	r ConvRepositoryInt,
	u users.UserServiceInt,
	workerSlots int,
	batchSize int64,
	sleepDelay time.Duration,
) *ConvService {
	if workerSlots <= 0 {
		workerSlots = 5
	}
	if batchSize <= 0 {
		batchSize = 500
	}
	return &ConvService{
		r:          r,
		u:          u,
		sem:        make(chan struct{}, workerSlots),
		batchSize:  batchSize,
		sleepDelay: sleepDelay,
	}
}

var _ ConvServiceInt = (*ConvService)(nil)

func (s *ConvService) GetOrCreateDirectConversation(
	ctx context.Context,
	input GetOrCreateDirectConvType,
) ([]GetConvParticipantType, error) {
	errChan := make(chan error, 2)
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		_, err := s.u.GetUserByID(ctx, input.UserIDOne)
		errChan <- err
	}()

	go func() {
		defer wg.Done()
		_, err := s.u.GetUserByID(ctx, input.UserIDTwo)
		errChan <- err
	}()

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return nil, errors.NewNotFoundError(err.Error())
		}
	}

	participants, err := s.r.GetOrCreateDirectConversation(ctx, input)
	if err != nil {
		return nil, errors.NewInternalServerError(err.Error(), err)
	}

	return participants, nil
}

func (s *ConvService) GetConvsByUserID(
	ctx context.Context,
	input int64,
) ([]GetConvParticipantType, error) {
	_, err := s.u.GetUserByID(ctx, input)
	if err != nil {
		return nil, err
	}

	participants := s.r.GetConvsByUserID(ctx, input)

	return participants, nil
}

func (s *ConvService) GetConvChats(
	ctx context.Context,
	convID int64,
	query types.URLQueryParams,
) []GetConvChatResponseType {
	if query.CursorTime.IsZero() || query.CursorID == 0 {
		chats := s.r.GetConvChats(ctx, convID, query)
		return chats
	}

	chats := s.r.GetConvChatsPaginated(ctx, convID, query)

	return chats
}

func (s *ConvService) CreateMessage(
	ctx context.Context,
	convID, senderID int64, content string,
) (*CreateMessageResponseType, error) {
	chat, err := s.r.CreateMessage(ctx, convID, senderID, content)
	if err != nil {
		return nil, errors.NewInternalServerError(err.Error(), err)
	}

	return chat, nil
}

func (s *ConvService) WipeConversation(ctx context.Context, convID, userID int64) error {
	if found := s.r.IsUserInConversation(ctx, convID, userID); !found {
		return errors.NewForbiddenError("forbidden")
	}

	if err := s.r.MarkConvForDeleting(ctx, convID); err != nil {
		return errors.NewInternalServerError(err.Error(), err)
	}

	bgCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 15*time.Minute)

	go func() {
		defer cancel()
		defer func() {
			if err := recover(); err != nil {
				stackTree := string(debug.Stack())
				slog.Error("Panic recoverd", "error", err, "stackTreee", stackTree)
			}
		}()

		select {
		case s.sem <- struct{}{}:
			defer func() {
				<-s.sem
			}()
		case <-bgCtx.Done():
			return
		}

		for {
			if err := bgCtx.Err(); err != nil {
				slog.WarnContext(bgCtx, "conversation wipeout aborted: context expired",
					"conv_id", convID,
					"error", err,
				)
				return
			}

			rowsDeleted, err := s.r.DeleteMessagesAsBatch(bgCtx, convID, s.batchSize)
			if err != nil {
				slog.ErrorContext(bgCtx, "failed to delete message batch during wipeout",
					"conv_id", convID,
					"batch_size", s.batchSize,
					"error", err,
				)
				return // critical: return immediatly
			}

			if rowsDeleted < s.batchSize {
				break
			}

			select {
			case <-bgCtx.Done():
				return
			case <-time.After(s.sleepDelay):
			}
		}

		if err := s.r.DeleteConvByID(bgCtx, convID); err != nil {
			slog.ErrorContext(bgCtx, "failed to delete conversation row after messages purged",
				"conv_id", convID,
				"error", err,
			)
			return
		}

		slog.InfoContext(bgCtx, "conversation wipeout completed successfully",
			"conv_id", convID,
		)

	}()

	return nil
}
