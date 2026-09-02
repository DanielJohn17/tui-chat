// Package types
package types

type BaseResponse[T comparable] struct {
	Status  int
	Success bool `json:"success"`
	Data    T    `json:"data"`

	// Error   *ErrorInfo `json:"error,omitempty"`
}

type APIResponse[T comparable] struct {
	BaseResponse[T]
}

type APIListResponse[T comparable] struct {
	BaseResponse[T]
	Data []T   `json:"data"`
	Meta *Meta `json:"meta"`
}

type APIErrorResponse struct {
	Status  int
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type Meta struct {
	Page       uint `json:"page,omitempty"`
	Limit      uint `json:"limit,omitempty"`
	Total      uint `json:"total,omitempty"`
	TotalPages uint `json:"total_pages,omitempty"`
}
