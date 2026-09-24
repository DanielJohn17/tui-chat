// Package conversations
package conversations

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/DanielJohn17/tui-chat/internal/api/database"
	"github.com/DanielJohn17/tui-chat/internal/api/types"
	"github.com/jackc/pgx/v5/pgtype"
)

type ConvRepositoryInt interface {
	GetOrCreateDirectConversation(
		ctx context.Context,
		input GetOrCreateDirectConvType,
	) ([]GetConvParticipantType, error)

	GetConvsByUserID(ctx context.Context, input int64) []GetConvParticipantType

	GetBulkChatsByUserID(ctx context.Context, userID, limit int64) ([]BulkChatsResponseType, error)

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

	IsUserInConversation(ctx context.Context, convID, userID int64) bool

	MarkConvForDeleting(ctx context.Context, convID int64) error

	DeleteMessagesAsBatch(ctx context.Context, convID, batchSize int64) (int64, error)

	DeleteConvByID(ctx context.Context, id int64) error

	GetUnreadCount(
		ctx context.Context,
		userID, convID int64,
	) (int, error)

	MarkAsRead(ctx context.Context, messageID, userID, convID int64) error
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

	GetBulkChatsByUserID(
		ctx context.Context,
		arg database.GetBulkChatsByUserIDParams,
	) ([]database.GetBulkChatsByUserIDRow, error)

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

	IsUserInConversation(ctx context.Context, arg database.IsUserInConversationParams) (bool, error)

	MarkConversationDeleting(ctx context.Context, id int64) error

	DeleteMessagesAsBatch(
		ctx context.Context,
		arg database.DeleteMessagesAsBatchParams,
	) (int64, error)

	DeleteConversationById(ctx context.Context, id int64) error

	GetUnreadCountForUser(
		ctx context.Context,
		arg database.GetUnreadCountForUserParams,
	) (int32, error)

	MarkConversationRead(ctx context.Context, arg database.MarkConversationReadParams) error
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
		participants[i] = GetConvParticipantType{
			ConvID:   v.ConvID,
			UserID:   v.UserID,
			Name:     v.Name,
			Username: v.Username,
		}
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
		convParticipants[i] = GetConvParticipantType{
			ConvID:          v.ConvID,
			UserID:          v.UserID,
			Name:            v.Name,
			Username:        v.Username,
			LastMessage:     v.LastMessage,
			LastMessageTime: v.LastMessageTime.Time.Format(time.RFC3339),
			UnreadCount:     int(v.UnreadCount),
		}
	}

	return convParticipants
}

func (r *ConvRepository) GetBulkChatsByUserID(ctx context.Context, userID, limit int64) ([]BulkChatsResponseType, error) {
	params := database.GetBulkChatsByUserIDParams{
		UserID:   userID,
		MsgLimit: int32(limit),
	}
	chats, err := r.q.GetBulkChatsByUserID(ctx, params)
	if err != nil {
		log.Printf("===>GetBulkChatsByUserID: %v", err)
		return nil, fmt.Errorf("error getting messages")
	}

	convChats := make([]BulkChatsResponseType, len(chats))
	for i, chat := range chats {
		convChats[i] = BulkChatsResponseType{
			ID:        chat.ID,
			SenderID:  chat.SenderID,
			ConvID:    chat.ConvID,
			Content:   chat.Content,
			CreatedAt: chat.CreatedAt.Time.Format(time.RFC3339),
			UpdatedAt: chat.UpdatedAt.Time.Format(time.RFC3339),
		}
	}

	return convChats, nil
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

func (r *ConvRepository) IsUserInConversation(
	ctx context.Context,
	convID, userID int64,
) bool {
	params := database.IsUserInConversationParams{
		ConvID: convID,
		UserID: userID,
	}

	found, err := r.q.IsUserInConversation(ctx, params)
	if err != nil {
		return false
	}

	return found
}

func (r *ConvRepository) MarkConvForDeleting(ctx context.Context, convID int64) error {
	if err := r.q.MarkConversationDeleting(ctx, convID); err != nil {
		return fmt.Errorf("error marking conversation for delete")
	}

	return nil
}

func (r *ConvRepository) DeleteMessagesAsBatch(
	ctx context.Context,
	convID, batchSize int64,
) (int64, error) {
	params := database.DeleteMessagesAsBatchParams{
		ConvID:    convID,
		BatchSize: batchSize,
	}

	rowsDeleted, err := r.q.DeleteMessagesAsBatch(ctx, params)
	if err != nil {
		return 0, err
	}

	return rowsDeleted, nil
}

func (r *ConvRepository) DeleteConvByID(ctx context.Context, id int64) error {
	if err := r.q.DeleteConversationById(ctx, id); err != nil {
		return err
	}

	return nil
}

func (r *ConvRepository) GetUnreadCount(
	ctx context.Context,
	userID, convID int64,
) (int, error) {
	params := database.GetUnreadCountForUserParams{
		UserID: userID,
		ConvID: convID,
	}

	unreadCount, err := r.q.GetUnreadCountForUser(ctx, params)
	if err != nil {
		return 0, err
	}

	return int(unreadCount), nil
}

func (r *ConvRepository) MarkAsRead(ctx context.Context, messageID, userID, convID int64) error {
	params := database.MarkConversationReadParams{
		MessageID: pgtype.Int8{Int64: messageID},
		ConvID:    pgtype.Int8{Int64: convID},
		UserID:    pgtype.Int8{Int64: userID},
	}

	if err := r.q.MarkConversationRead(ctx, params); err != nil {
		return err
	}

	return nil
}
