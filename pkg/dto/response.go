package dto

import "github.com/gofiber/fiber/v3"

type RenderData struct {
	Title string `handlebars:"Title"`
	Data  any    `handlebars:"Data"`
	Error string `handlebars:"Error"`
}

func (r RenderData) ToMap() fiber.Map {
	return fiber.Map{
		"Title": r.Title,
		"Data":  r.Data,
		"Error": r.Error,
	}
}
