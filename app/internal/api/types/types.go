// Package types
package types

import "time"

type URLParamString struct {
	Str string `uri:"username" binding:"required,alphanum,min=3,max=20"`
}

type URLParamInt struct {
	ID int `uri:"id" binding:"required,gt=0"`
}

type URLQueryParams struct {
	Cursor     string    `form:"cursor"           binding:"omitempty"`
	CursorTime time.Time `form:"-"`
	CursorID   int64     `form:"-"`
	Limit      int64     `form:"limit,default=30" binding:"omitempty,gte=1,lte=100"`
}

type APIResponse[T any] struct {
	Success bool  `json:"success"`
	Data    T     `json:"data,omitempty"`
	Meta    *Meta `json:"meta,omitempty"`
}

type Meta struct {
	Page       int64  `json:"page,omitempty"`
	Cursor     string `json:"cursor,omitempty"`
	Limit      int64  `json:"limit,omitempty"`
	Total      int64  `json:"total,omitempty"`
	TotalPages int64  `json:"total_pages,omitempty"`
}
