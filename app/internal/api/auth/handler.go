package auth

import (
	"net/http"

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

	if err := helpers.ParseJSON(c, &registerUser); err != nil {
		helpers.WriteError(c, err)
		return
	}

	userRegisterd, err := h.s.Register(ctx, registerUser)
	if err != nil {
		helpers.WriteError(c, err)
		return
	}

	helpers.WriteJSON(c, http.StatusCreated, userRegisterd)
}

func (h *AuthHandler) LoginUser(c *gin.Context) {
	var loginUser LoginUserType

	ctx := c.Request.Context()

	if err := helpers.ParseJSON(c, &loginUser); err != nil {
		helpers.WriteError(c, err)
		return
	}

	userLogin, err := h.s.Login(ctx, loginUser)
	if err != nil {
		helpers.WriteError(c, err)
		return
	}

	helpers.WriteJSON(c, http.StatusOK, userLogin)
}
