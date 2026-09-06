package users

import (
	"net/http"

	"github.com/DanielJohn17/tui-chat/app/internal/api/helpers"
	"github.com/DanielJohn17/tui-chat/app/internal/api/types"
	"github.com/gin-gonic/gin"
)

type UserHandlerInt interface {
	GetUserByUsername(c *gin.Context)
	GetUserByID(c *gin.Context)
}

type UserHandler struct {
	s UserServiceInt
}

func NewUserHandler(s UserServiceInt) *UserHandler {
	return &UserHandler{s: s}
}

var _ UserHandlerInt = (*UserHandler)(nil)

func (h *UserHandler) GetUserByUsername(c *gin.Context) {

	ctx := c.Request.Context()
	req := c.MustGet("params").(types.URLParamString)

	user, err := h.s.GetUserByUsername(ctx, req.Str)
	if err != nil {
		helpers.WriteError(c, err)
		return
	}

	helpers.WriteJSON(c, http.StatusOK, user)
}

func (h *UserHandler) GetUserByID(c *gin.Context) {
	ctx := c.Request.Context()
	req := c.MustGet("params").(types.URLParamInt)

	user, apiError := h.s.GetUserByID(ctx, int64(req.ID))
	if apiError != nil {
		helpers.WriteError(c, apiError)
		return
	}

	helpers.WriteJSON(c, http.StatusOK, user)
}
