package repository

import "gorm.io/gorm"

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
