package conversations

import (
	"net/http"

	"github.com/DanielJohn17/tui-chat/app/internal/api/helpers"
	"github.com/gin-gonic/gin"
)

type ConvHandlerInt interface {
	GetOrCreateDirectConversation(c *gin.Context)
}

type ConvHandler struct {
	s ConvServiceInt
}

func NewConvHandler(s ConvServiceInt) *ConvHandler {
	return &ConvHandler{s: s}
}

func (h *ConvHandler) GetOrCreateDirectConversation(c *gin.Context) {
	var params GetOrCreateDirectConvType
	ctx := c.Request.Context()

	if err := helpers.ParseJSON(c, &params); err != nil {
		helpers.WriteError(c, err)
		return
	}

	participants, err := h.s.GetOrCreateDirectConversation(ctx, params)
	if err != nil {
		helpers.WriteError(c, err)
		return
	}

	helpers.WriteJSON(c, http.StatusOK, participants)
}
