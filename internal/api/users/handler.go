package users

import (
	"net/http"

	"github.com/DanielJohn17/tui-chat/internal/api/apierrs"
	"github.com/DanielJohn17/tui-chat/internal/api/helpers"
	"github.com/DanielJohn17/tui-chat/internal/api/types"
	"github.com/gin-gonic/gin"
)

type UserHandlerInt interface {
	SearchUsers(c *gin.Context)
	GetUserByUsername(c *gin.Context)
	GetUserByID(c *gin.Context)
	GetMe(c *gin.Context)
	UpdateProfile(c *gin.Context)
	RegenerateInviteToken(c *gin.Context)
	GetUserByInviteToken(c *gin.Context)
}

type UserHandler struct {
	s UserServiceInt
}

func NewUserHandler(s UserServiceInt) *UserHandler {
	return &UserHandler{s: s}
}

var _ UserHandlerInt = (*UserHandler)(nil)

func (h *UserHandler) SearchUsers(c *gin.Context) {
	userID, ok := c.Get("userId")
	id, valid := userID.(int64)
	if !ok || !valid || id <= 0 {
		helpers.WriteError(c, apierrs.NewUnauthorizedError("unauthorized"))
		return
	}
	result, err := h.s.SearchUsers(c.Request.Context(), id, c.Query("q"))
	if err != nil {
		helpers.WriteError(c, err)
		return
	}
	helpers.WriteJSON(c, http.StatusOK, result)
}

func (h *UserHandler) GetUserByUsername(c *gin.Context) {
	ctx := c.Request.Context()
	params, ok := c.Get("params")
	req, valid := params.(types.URLParamString)
	if !ok || !valid || req.Str == "" {
		helpers.WriteError(c, apierrs.NewBadRequestError("username is not present"))
		return
	}

	user, err := h.s.GetUserByUsername(ctx, req.Str)
	if err != nil {
		helpers.WriteError(c, err)
		return
	}

	helpers.WriteJSON(c, http.StatusOK, user)
}

func (h *UserHandler) GetUserByID(c *gin.Context) {
	ctx := c.Request.Context()
	params, ok := c.Get("params")
	req, valid := params.(types.URLParamInt)
	if !ok || !valid || req.ID <= 0 {
		helpers.WriteError(c, apierrs.NewBadRequestError("id is not present"))
		return
	}

	user, apiError := h.s.GetUserByID(ctx, int64(req.ID))
	if apiError != nil {
		helpers.WriteError(c, apiError)
		return
	}

	helpers.WriteJSON(c, http.StatusOK, user)
}

func (h *UserHandler) GetMe(c *gin.Context) {
	userID, ok := c.Get("userId")
	id, valid := userID.(int64)
	if !ok || !valid || id <= 0 {
		helpers.WriteError(c, apierrs.NewUnauthorizedError("unauthorized"))
		return
	}

	user, err := h.s.GetUserByID(c.Request.Context(), id)
	if err != nil {
		helpers.WriteError(c, err)
		return
	}

	helpers.WriteJSON(c, http.StatusOK, user)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, ok := c.Get("userId")
	id, valid := userID.(int64)
	if !ok || !valid || id <= 0 {
		helpers.WriteError(c, apierrs.NewUnauthorizedError("unauthorized"))
		return
	}

	var req UpdateUserProfileRequest
	if err := helpers.ParseJSON(c, &req); err != nil {
		helpers.WriteError(c, err)
		return
	}

	updated, err := h.s.UpdateProfile(c.Request.Context(), id, req)
	if err != nil {
		helpers.WriteError(c, err)
		return
	}

	helpers.WriteJSON(c, http.StatusOK, updated)
}

func (h *UserHandler) RegenerateInviteToken(c *gin.Context) {
	userID, ok := c.Get("userId")
	id, valid := userID.(int64)
	if !ok || !valid || id <= 0 {
		helpers.WriteError(c, apierrs.NewUnauthorizedError("unauthorized"))
		return
	}

	updated, err := h.s.RegenerateInviteToken(c.Request.Context(), id)
	if err != nil {
		helpers.WriteError(c, err)
		return
	}

	helpers.WriteJSON(c, http.StatusOK, updated)
}

func (h *UserHandler) GetUserByInviteToken(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		helpers.WriteError(c, apierrs.NewBadRequestError("invite token is required"))
		return
	}

	user, err := h.s.GetUserByInviteToken(c.Request.Context(), token)
	if err != nil {
		helpers.WriteError(c, err)
		return
	}

	helpers.WriteJSON(c, http.StatusOK, user)
}
