package auth

import (
	"github.com/DanielJohn17/tui-chat/app/internal/api/helpers"
	"github.com/gin-gonic/gin"
)

type AuthHandlerInt interface {
	RegisterUser(c *gin.Context)
	LoginUser(c *gin.Context)
}

type AuthHandler struct {
	s AuthServiceInt
}

func NewAuthHander(as AuthServiceInt) *AuthHandler {
	return &AuthHandler{s: as}
}

var _ AuthHandlerInt = (*AuthHandler)(nil)

func (h *AuthHandler) RegisterUser(c *gin.Context) {
	var registerUser RegisterUserType

	ctx := c.Request.Context()

	if apiError := helpers.ParseJSON(c, &registerUser); apiError != nil {
		helpers.WriteError(c, *apiError)
		return
	}

	apiResp, apiError := h.s.Register(ctx, registerUser)
	if apiError != nil {
		helpers.WriteError(c, *apiError)
		return
	}

	helpers.WriteJSON(c, *apiResp)
}

func (h *AuthHandler) LoginUser(c *gin.Context) {
	var loginUser LoginUserType

	ctx := c.Request.Context()

	if apiError := helpers.ParseJSON(c, &loginUser); apiError != nil {
		helpers.WriteError(c, *apiError)
		return
	}

	apiResp, apiError := h.s.Login(ctx, loginUser)
	if apiError != nil {
		helpers.WriteError(c, *apiError)
		return
	}

	helpers.WriteJSON(c, *apiResp)
}
