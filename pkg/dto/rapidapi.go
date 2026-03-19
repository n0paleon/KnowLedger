package dto

import (
	"KnowLedger/internal/model"
	"KnowLedger/pkg/utils"
	"html"
)

type FactResponse struct {
	Content     string             `json:"content"`
	ContentHTML string             `json:"content_html"`
	SourceURL   string             `json:"source_url"`
	Media       *FactMediaResponse `json:"media"`
	Tags        []FactTagsResponse `json:"tags"`
}

type FactTagsResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type FactMediaResponse struct {
	URL         string `json:"url"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
}

func ConvertFactModelToDto(fact *model.Fact) FactResponse {
	f := FactResponse{}
	f.Content = utils.StripHTML(fact.Content)
	f.ContentHTML = html.UnescapeString(fact.Content)
	f.SourceURL = fact.SourceURL

	if fact.Media != nil {
		f.Media = &FactMediaResponse{
			URL:         fact.Media.URL,
			ContentType: fact.Media.ContentType,
			Size:        fact.Media.Size,
		}
	}

	tags := make([]FactTagsResponse, len(fact.Tags))
	for i, t := range fact.Tags {
		tags[i] = FactTagsResponse{
			ID:   t.ID,
			Name: t.Name,
		}
	}
	f.Tags = tags

	return f
}
