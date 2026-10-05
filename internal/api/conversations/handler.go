package conversations

import (
	"net/http"

	"github.com/DanielJohn17/tui-chat/internal/api/apierrs"
	"github.com/DanielJohn17/tui-chat/internal/api/helpers"
	"github.com/DanielJohn17/tui-chat/internal/api/types"
	"github.com/gin-gonic/gin"
)

type ConvHandlerInt interface {
	GetOrCreateDirectConversation(c *gin.Context)
	GetConvsByUserID(c *gin.Context)
	GetBulkChatsByUserID(c *gin.Context)
	GetConvChats(c *gin.Context)
	WipeConversation(c *gin.Context)
}

type ConvHandler struct {
	s ConvServiceInt
}

func NewConvHandler(s ConvServiceInt) *ConvHandler {
	return &ConvHandler{s: s}
}

var _ ConvHandlerInt = (*ConvHandler)(nil)

func (h *ConvHandler) GetOrCreateDirectConversation(c *gin.Context) {
	var params GetOrCreateDirectConvType
	ctx := c.Request.Context()
	userIDParam, ok := c.Get("userId")
	authUserID, valid := userIDParam.(int64)
	if !ok || !valid || authUserID <= 0 {
		helpers.WriteError(c, apierrs.NewUnauthorizedError("unauthorized"))
		return
	}

	if err := helpers.ParseJSON(c, &params); err != nil {
		helpers.WriteError(c, err)
		return
	}

	params.UserIDOne = authUserID

	participants, err := h.s.GetOrCreateDirectConversation(ctx, params)
	if err != nil {
		helpers.WriteError(c, err)
		return
	}

	helpers.WriteJSON(c, http.StatusOK, participants)
}

func (h *ConvHandler) GetConvsByUserID(c *gin.Context) {
	ctx := c.Request.Context()

	userIDParam, ok := c.Get("userId")
	userID, valid := userIDParam.(int64)
	if !ok || !valid || userID <= 0 {
		helpers.WriteError(c, apierrs.NewUnauthorizedError("unauthorized"))
		return
	}

	conversations, err := h.s.GetConvsByUserID(ctx, userID)
	if err != nil {
		helpers.WriteError(c, err)
		return
	}

	meta := &types.Meta{
		Total: int64(len(conversations)),
	}

	helpers.WriteJSONWithMeta(c, http.StatusOK, conversations, meta)
}

func (h *ConvHandler) GetBulkChatsByUserID(c *gin.Context) {
	userIDParam, ok := c.Get("userId")
	userID, valid := userIDParam.(int64)
	if !ok || !valid || userID <= 0 {
		helpers.WriteError(c, apierrs.NewUnauthorizedError("unauthorized"))
		return
	}

	ctx := c.Request.Context()

	var limit int64 = 50

	chats, err := h.s.GetBulkChatsByUserID(ctx, userID, limit)
	if err != nil {
		helpers.WriteError(c, err)
		return
	}

	helpers.WriteJSON(c, http.StatusOK, chats)
}

func (h *ConvHandler) GetConvChats(c *gin.Context) {
	paramVal, pOk := c.Get("params")
	params, pValid := paramVal.(types.URLParamInt)
	if !pOk || !pValid || params.ID <= 0 {
		helpers.WriteError(c, apierrs.NewBadRequestError("invalid conversation id"))
		return
	}

	userIDParam, uOk := c.Get("userId")
	userID, uValid := userIDParam.(int64)
	if !uOk || !uValid || userID <= 0 {
		helpers.WriteError(c, apierrs.NewUnauthorizedError("unauthorized"))
		return
	}

	var queries types.URLQueryParams
	if q, exists := c.Get("queries"); exists {
		if qParams, ok := q.(types.URLQueryParams); ok {
			queries = qParams
		}
	}
	ctx := c.Request.Context()

	convID := params.ID

	chats, err := h.s.GetConvChats(ctx, int64(convID), userID, queries)
	if err != nil {
		helpers.WriteError(c, err)
		return
	}

	var meta *types.Meta
	if len(chats) > 0 {
		lastChat := chats[len(chats)-1]
		meta = &types.Meta{
			Cursor: helpers.EncodeCursor(lastChat.CreatedAt, lastChat.ID),
			Limit:  queries.Limit,
		}
	} else {
		meta = &types.Meta{
			Limit: queries.Limit,
		}
	}

	helpers.WriteJSONWithMeta(c, http.StatusOK, chats, meta)
}

func (h *ConvHandler) WipeConversation(c *gin.Context) {
	userIDParam, uOk := c.Get("userId")
	userID, uValid := userIDParam.(int64)
	if !uOk || !uValid || userID <= 0 {
		helpers.WriteError(c, apierrs.NewUnauthorizedError("unauthorized"))
		return
	}

	convIDParam, cOk := c.Get("params")
	convParam, cValid := convIDParam.(types.URLParamInt)
	if !cOk || !cValid || convParam.ID <= 0 {
		helpers.WriteError(c, apierrs.NewBadRequestError("invalid conversation id"))
		return
	}
	convID := int64(convParam.ID)

	ctx := c.Request.Context()

	if err := h.s.WipeConversation(ctx, convID, userID); err != nil {
		helpers.WriteError(c, err)
		return
	}

	helpers.WriteJSON(c, http.StatusAccepted, gin.H{"message": "Conversation deleted"})
}
