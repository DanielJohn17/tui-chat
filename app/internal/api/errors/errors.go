// Package errors
package errors

import (
	"fmt"
	"net/http"

	"github.com/DanielJohn17/tui-chat/app/internal/api/config"
)

type APIError struct {
	Code    int    `json:"-"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

var (
	goEnv string
)

func init() {
	goEnv = config.ENV.GoEnv
}

func (e *APIError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}

	return e.Message
}

func (e *APIError) Unwrap() error {
	return e.Err
}

func NewNotFoundError(msg string) *APIError {
	return &APIError{Code: http.StatusNotFound, Message: msg}
}

func NewConflictError(msg string) *APIError {
	return &APIError{Code: http.StatusConflict, Message: msg}
}

func NewUnauthorizedError(msg string) *APIError {
	return &APIError{Code: http.StatusUnauthorized, Message: msg}
}

func NewForbiddenError(msg string) *APIError {
	return &APIError{Code: http.StatusForbidden, Message: msg}
}

func NewBadRequestError(msg string) *APIError {
	return &APIError{Code: http.StatusBadRequest, Message: msg}
}

func NewInternalServerError(msg string, err error) *APIError {
	if goEnv != "development" {
		err = nil
	}
	return &APIError{Code: http.StatusInternalServerError, Message: msg, Err: err}
}
