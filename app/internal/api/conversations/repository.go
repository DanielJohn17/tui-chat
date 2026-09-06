// Package conversations
package conversations

import (
	"context"
	"fmt"
	"log"

	"github.com/DanielJohn17/tui-chat/app/internal/api/database"
)

type ConvRepositoryInt interface {
	GetOrCreateDirectConversation(
		ctx context.Context,
		input GetOrCreateDirectConvType,
	) ([]GetConvParticipantType, error)
}

type ConvQuerier interface {
	GetOrCreateDirectConversation(
		ctx context.Context,
		arg database.GetOrCreateDirectConversationParams,
	) ([]database.GetOrCreateDirectConversationRow, error)
}

type ConvRepository struct {
	q ConvQuerier
}

func NewConvRepository(q ConvQuerier) *ConvRepository {
	return &ConvRepository{q: q}
}

var _ ConvRepositoryInt = (*ConvRepository)(nil)

func (r *ConvRepository) GetOrCreateDirectConversation(
	ctx context.Context,
	params GetOrCreateDirectConvType,
) ([]GetConvParticipantType, error) {

	convParams := database.GetOrCreateDirectConversationParams{
		UserID:   params.UserIDOne,
		UserID_2: params.UserIDTwo,
	}

	convParticipants, err := r.q.GetOrCreateDirectConversation(ctx, convParams)
	if err != nil {
		log.Printf("===> GetOrCreateDirectConversation: %v\n", err)
		return nil, fmt.Errorf("error creating conversation")
	}

	participants := make([]GetConvParticipantType, len(convParticipants))
	for i, v := range convParticipants {
		participants[i] = GetConvParticipantType(v)
	}

	return participants, nil
}
