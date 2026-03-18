package dto

import "KnowLedger/internal/model"

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

type ListFactsParams struct {
	Page    int              `query:"page" validate:"numeric,gte=0"`
	Limit   int              `query:"limit" validate:"numeric,gte=0,lte=100"`
	Search  string           `query:"search" validate:"omitempty,lte=200"`
	Status  model.FactStatus `query:"status" validate:"omitempty,oneof=draft published"`
	SortBy  string           `query:"sortBy" validate:"omitempty,oneof=created_at updated_at"`
	SortDir string           `query:"sortDir" validate:"omitempty,oneof=asc desc"`
}

type ListTagsParams struct {
	Page      int    `query:"page" validate:"numeric,gte=0"`
	Limit     int    `query:"limit" validate:"numeric,gte=0,lte=500"`
	SortOrder string `query:"sortOrder" validate:"omitempty,oneof=asc desc"`
}

type PostChangePasswordRequest struct {
	Password        string `json:"password" validate:"required,min=6,max=50"`
	ConfirmPassword string `json:"confirm_password" validate:"required,min=6,max=50"`
}

type GCJobListParams struct {
	Page    int                `query:"page" validate:"numeric,gte=0"`
	Limit   int                `query:"limit" validate:"numeric,gte=0"`
	Status  model.GCJobStatus  `query:"status" validate:"omitempty,oneof=pending running completed failed"`
	Trigger model.GCJobTrigger `query:"trigger" validate:"omitempty,oneof=automatic manual"`
	SortDir string             `query:"sortDir" validate:"omitempty,oneof=asc desc"`
}
