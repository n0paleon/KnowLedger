package dto

type APIResponse struct {
	Error string `json:"error"`
	Data  any    `json:"data"`
}

type CreateFunFactAPIRequest struct {
	Content   string   `json:"content" validate:"required"`
	Tags      []string `json:"tags"`
	SourceURL string   `json:"source_url" validate:"omitempty,url"`
	MediaKey  string   `json:"media_key"`
}
