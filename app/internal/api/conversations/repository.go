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

	GetConvsByUserID(ctx context.Context, input int64) []GetConvParticipantType
}

type ConvQuerier interface {
	GetOrCreateDirectConversation(
		ctx context.Context,
		arg database.GetOrCreateDirectConversationParams,
	) ([]database.GetOrCreateDirectConversationRow, error)

	GetConversationsByUserId(
		ctx context.Context,
		userID int64,
	) ([]database.GetConversationsByUserIdRow, error)
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

func (r *ConvRepository) GetConvsByUserID(
	ctx context.Context,
	input int64,
) []GetConvParticipantType {
	conversations, err := r.q.GetConversationsByUserId(ctx, input)
	if err != nil {
		log.Printf("===> GetConvsByUserId: %v\n", err)
		return []GetConvParticipantType{}
	}

	convParticipants := make([]GetConvParticipantType, len(conversations))
	for i, v := range conversations {
		convParticipants[i] = GetConvParticipantType(v)
	}

	return convParticipants
}
