package repository

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Repository[T any] struct {
	db      *gorm.DB
	factory func(*gorm.DB) *T
}

func (r *Repository[T]) WithTx(tx *gorm.DB) *T {
	if tx == nil {
		return r.factory(r.db)
	}
	return r.factory(tx)
}

func WithPagination(page, limit int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page <= 0 {
			page = 1
		}
		hardLimit := 300
		fallbackLimit := 20
		if limit <= 0 || limit > hardLimit {
			zap.L().Warn("pagination limit out of range, falling back to default",
				zap.Int("request_limit", limit),
				zap.Int("fallback_limit", fallbackLimit),
				zap.Int("hard_limit", hardLimit),
			)
			limit = fallbackLimit
		}
		return db.Limit(limit).Offset((page - 1) * limit)
	}
}
