package model

import "math"

type Paginated[T any] struct {
	Data       []T   `json:"data"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"total_pages"`
}

func NewPaginated[T any](data []T, total int64, page, limit int) *Paginated[T] {
	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	return &Paginated[T]{
		Data:       data,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: int(math.Ceil(float64(total) / float64(limit))),
	}
}
