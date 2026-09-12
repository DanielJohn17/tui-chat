package conversations

import (
	"context"
	"sync"

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
}

type ConvService struct {
	r ConvRepositoryInt
	u users.UserServiceInt
}

func NewConvService(r ConvRepositoryInt, u users.UserServiceInt) *ConvService {
	return &ConvService{r: r, u: u}
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
