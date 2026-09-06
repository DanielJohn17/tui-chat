package conversations

import (
	"context"
	"sync"

	"github.com/DanielJohn17/tui-chat/app/internal/api/errors"
	"github.com/DanielJohn17/tui-chat/app/internal/api/users"
)

type ConvServiceInt interface {
	GetOrCreateDirectConversation(
		ctx context.Context,
		input GetOrCreateDirectConvType,
	) ([]GetConvParticipantType, error)
}

type ConvService struct {
	r ConvRepositoryInt
	u users.UserServiceInt
}

func NewConvService(r ConvRepositoryInt) *ConvService {
	return &ConvService{r: r}
}

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
