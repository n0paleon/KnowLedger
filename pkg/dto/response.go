package dto

import (
	"github.com/gofiber/fiber/v3"
	"github.com/iancoleman/orderedmap"
)

type RenderData struct {
	Title string
	Data  any
	Error string
}

func (r RenderData) ToMap() fiber.Map {
	o := orderedmap.New()

	o.Set("Title", r.Title)
	o.Set("Data", r.Data)
	o.Set("Error", r.Error)

	return o.Values()
}
