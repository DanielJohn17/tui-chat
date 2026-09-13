// Package conversations
package conversations

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/DanielJohn17/tui-chat/app/internal/api/database"
	"github.com/DanielJohn17/tui-chat/app/internal/api/types"
	"github.com/jackc/pgx/v5/pgtype"
)

type ConvRepositoryInt interface {
	GetOrCreateDirectConversation(
		ctx context.Context,
		input GetOrCreateDirectConvType,
	) ([]GetConvParticipantType, error)

	GetConvsByUserID(ctx context.Context, input int64) []GetConvParticipantType

	GetConvChats(
		ctx context.Context,
		convID int64,
		input types.URLQueryParams,
	) []GetConvChatResponseType

	GetConvChatsPaginated(
		ctx context.Context,
		convID int64,
		input types.URLQueryParams,
	) []GetConvChatResponseType

	CreateMessage(
		ctx context.Context,
		convID, senderID int64,
		content string,
	) (*CreateMessageResponseType, error)
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

	GetConvChats(
		ctx context.Context,
		arg database.GetConvChatsParams,
	) ([]database.GetConvChatsRow, error)

	GetConvChatsPaginated(
		ctx context.Context,
		arg database.GetConvChatsPaginatedParams,
	) ([]database.GetConvChatsPaginatedRow, error)

	CreateMessageAndGetRecipient(
		ctx context.Context,
		arg database.CreateMessageAndGetRecipientParams,
	) (database.CreateMessageAndGetRecipientRow, error)
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

func (r *ConvRepository) GetConvChats(
	ctx context.Context,
	convID int64,
	query types.URLQueryParams,
) []GetConvChatResponseType {
	limit := query.Limit
	if limit <= 0 {
		limit = 30
	}

	convParams := database.GetConvChatsParams{
		ConvID: convID,
		Limit:  limit,
	}

	chats, err := r.q.GetConvChats(ctx, convParams)
	if err != nil {
		log.Printf("===? GetConvChatsPaginated: %v\n", err)
		return []GetConvChatResponseType{}
	}

	convChats := make([]GetConvChatResponseType, len(chats))
	for i, v := range chats {
		convChats[i] = GetConvChatResponseType{
			ID:        v.ID,
			SenderID:  v.SenderID,
			Content:   v.Content,
			CreatedAt: v.CreatedAt.Time.Format(time.RFC3339),
			UpdatedAt: v.UpdatedAt.Time.Format(time.RFC3339),
		}
	}

	return convChats
}

func (r *ConvRepository) GetConvChatsPaginated(
	ctx context.Context,
	convID int64,
	query types.URLQueryParams,
) []GetConvChatResponseType {
	limit := query.Limit
	if limit <= 0 {
		limit = 30
	}

	timestamptzVal := pgtype.Timestamptz{
		Time:  query.CursorTime,
		Valid: true,
	}

	convParams := database.GetConvChatsPaginatedParams{
		ConvID:    convID,
		CreatedAt: timestamptzVal,
		ID:        query.CursorID,
		Limit:     limit,
	}

	chats, err := r.q.GetConvChatsPaginated(ctx, convParams)
	if err != nil {
		log.Printf("===> GetConvChatsPaginated: %v\n", err)
		return []GetConvChatResponseType{}
	}

	convChats := make([]GetConvChatResponseType, len(chats))
	for i, v := range chats {
		convChats[i] = GetConvChatResponseType{
			ID:        v.ID,
			SenderID:  v.SenderID,
			Content:   v.Content,
			CreatedAt: v.CreatedAt.Time.Format(time.RFC3339),
			UpdatedAt: v.UpdatedAt.Time.Format(time.RFC3339),
		}
	}

	return convChats
}

func (r *ConvRepository) CreateMessage(
	ctx context.Context,
	convID,
	senderID int64,
	content string,
) (*CreateMessageResponseType, error) {
	pgConvID := pgtype.Int8{Int64: convID, Valid: true}
	pgSenderID := pgtype.Int8{Int64: senderID, Valid: true}
	pgContent := pgtype.Text{String: content, Valid: true}

	messageParam := database.CreateMessageAndGetRecipientParams{
		SenderID: pgSenderID,
		ConvID:   pgConvID,
		Content:  pgContent,
	}

	chat, err := r.q.CreateMessageAndGetRecipient(ctx, messageParam)
	if err != nil {
		log.Printf("==> CreateMessage: %v", err)
		return nil, fmt.Errorf("error saving message")
	}

	return &CreateMessageResponseType{
		ID:          chat.ID,
		SenderID:    chat.SenderID,
		ConvID:      chat.ConvID,
		RecipientID: chat.RecipientID,
		Content:     chat.Content,
		CreatedAt:   chat.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:   chat.UpdatedAt.Time.Format(time.RFC3339),
	}, nil

}
