package keys

import (
	"net/http"
	"strconv"

	"github.com/DanielJohn17/tui-chat/internal/api/apierrs"
	"github.com/DanielJohn17/tui-chat/internal/api/helpers"
	"github.com/DanielJohn17/tui-chat/internal/api/types"
	gin "github.com/gin-gonic/gin"
)

type KeyHandlerInt interface {
	UploadPublicKey(c *gin.Context)
	GetPublicKey(c *gin.Context)
}

type KeyHandler struct {
	s KeyServiceInt
}

func NewKeyHandler(s KeyServiceInt) *KeyHandler {
	return &KeyHandler{s: s}
}

var _ KeyHandlerInt = (*KeyHandler)(nil)

func (h *KeyHandler) UploadPublicKey(c *gin.Context) {
	userIDAny, exists := c.Get("userId")
	if !exists {
		helpers.WriteError(c, apierrs.NewUnauthorizedError("unauthorized"))
		return
	}
	userID, ok := userIDAny.(int64)
	if !ok {
		helpers.WriteError(c, apierrs.NewUnauthorizedError("unauthorized"))
		return
	}

	var req UploadPublicKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helpers.WriteError(c, apierrs.NewBadRequestError(err.Error()))
		return
	}

	if err := h.s.UploadPublicKey(c.Request.Context(), userID, req.PublicKey); err != nil {
		helpers.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, types.APIResponse[any]{Success: true})
}

func (h *KeyHandler) GetPublicKey(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil || userID <= 0 {
		helpers.WriteError(c, apierrs.NewBadRequestError("invalid user_id"))
		return
	}

	resp, err := h.s.GetPublicKey(c.Request.Context(), userID)
	if err != nil {
		helpers.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, types.APIResponse[PublicKeyResponse]{Success: true, Data: *resp})
}
