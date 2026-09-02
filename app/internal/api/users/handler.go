package users

import (
	"github.com/DanielJohn17/tui-chat/app/internal/api/helpers"
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

	req := c.MustGet("params").(URLParam)

	apiResponse, apiError := h.s.GetUserByUsername(c, req.Username)
	if apiError != nil {
		helpers.WriteError(c, *apiError)
		return
	}

	helpers.WriteJSON(c, *apiResponse)
}

func (h *UserHandler) GetUserByID(c *gin.Context) {

}
