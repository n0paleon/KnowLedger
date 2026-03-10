package dto

import "KnowLedger/internal/model"

type GetFactsRequest struct {
	Page  int `query:"page" validate:"numeric,min=1,gte=1"`
	Limit int `query:"limit" validate:"numeric,min=1,gte=1,lte=20"`
}

type PostCreateFunFactRequest struct {
	Content   string           `form:"content" validate:"required"`
	Tags      string           `form:"tags" validate:""`
	Status    model.FactStatus `form:"status" validate:"required,oneof=draft published"`
	SourceURL string           `form:"source_url" validate:"omitempty,url"`
	MediaKey  string           `form:"media_key" validate:""`
}

type PostEditFunFactRequest struct {
	ID        string           `uri:"id"`
	Content   string           `form:"content" validate:"required"`
	Tags      string           `form:"tags" validate:""`
	Status    model.FactStatus `form:"status" validate:"required,oneof=draft published"`
	SourceURL string           `form:"source_url" validate:"omitempty,url"`
	MediaKey  string           `form:"media_key" validate:""`
}
