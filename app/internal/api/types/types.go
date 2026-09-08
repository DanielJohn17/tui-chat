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
	CursorTime time.Time `form:"cursor_time" binding:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	CursorID   int64     `form:"cursor_id"   binding:"omitempty,gte=1"`
	Limit      int64     `form:"limit,default=30" binding:"omitempty,gte=1,lte=100"`
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
