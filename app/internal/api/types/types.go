// Package types
package types

type URLParamString struct {
	Str string `uri:"username" binding:"required,alphanum,min=3,max=20"`
}

type URLParamInt struct {
	ID int `uri:"id" binding:"required,gt=0"`
}

type APIResponse[T any] struct {
	Success bool  `json:"success"`
	Data    T     `json:"data,omitempty"`
	Meta    *Meta `json:"meta,omitempty"`
}

type Meta struct {
	Page       uint `json:"page,omitempty"`
	Limit      uint `json:"limit,omitempty"`
	Total      uint `json:"total,omitempty"`
	TotalPages uint `json:"total_pages,omitempty"`
}
