package handler

import (
	"KnowLedger/internal/service"
	"KnowLedger/pkg/dto"
	"fmt"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

type AdminHandler struct {
	funFactService *service.FunFactService
	log            *zap.Logger
}

func NewAdminHandler(funFactService *service.FunFactService, logger *zap.Logger) *AdminHandler {
	return &AdminHandler{
		funFactService: funFactService,
		log:            logger,
	}
}

func (h *AdminHandler) ShowCreateFunFact(c fiber.Ctx) error {
	return c.Render("pages/admin/create-funfact", dto.RenderData{
		Title: "Create Fun Fact",
	}.ToMap())
}

func (h *AdminHandler) CreateFunFact(c fiber.Ctx) error {
	req := new(dto.PostCreateFunFactRequest)
	renderData := new(dto.RenderData)
	renderData.Title = "Create Fun Fact"

	if err := c.Bind().Form(req); err != nil {
		renderData.Data = fiber.Map{
			"success": false,
			"msg":     err.Error(),
		}
		return c.Render("pages/admin/create-funfact", renderData.ToMap())
	}

	err := h.funFactService.CreateFact(c, req)
	if err != nil {
		h.log.Error("Create Fun Fact Error", zap.Error(err))
		renderData.Data = fiber.Map{
			"success": false,
			"msg":     err.Error(),
		}
		return c.Render("pages/admin/create-funfact", renderData.ToMap())
	}

	renderData.Data = fiber.Map{
		"success": true,
		"msg":     "Data created successfully",
	}
	return c.Render("pages/admin/create-funfact", renderData.ToMap())
}

func (h *AdminHandler) ShowDashboardIndex(c fiber.Ctx) error {
	page := fiber.Query[int](c, "page")
	limit := fiber.Query[int](c, "limit")

	req := &dto.GetFactsRequest{
		Page:  1,
		Limit: 20,
	}

	if page == 0 || limit == 0 {
		return c.Redirect().To(fmt.Sprintf("/admin?page=%d&limit=%d", req.Page, req.Limit))
	}

	if err := c.Bind().Query(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	facts, err := h.funFactService.GetFacts(c, req.Page, req.Limit)
	if err != nil {
		h.log.Error("GetFacts error", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Render("pages/admin/index",
		dto.RenderData{
			Title: "Dashboard",
			Data:  facts,
		}.ToMap(),
		"layouts/main",
	)
}

func (h *AdminHandler) ShowEditFunFact(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "id is required",
		})
	}

	fact, err := h.funFactService.GetOneFunFact(c, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Render("pages/admin/edit-funfact", dto.RenderData{
		Title: "Edit Fun Fact",
		Data: fiber.Map{
			"Fact": fact,
		},
	}.ToMap())
}

func (h *AdminHandler) EditFunFact(c fiber.Ctx) error {
	req := new(dto.PostEditFunFactRequest)

	if err := c.Bind().All(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if req.ID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "id is required",
		})
	}

	renderData := new(dto.RenderData)
	renderData.Title = "Edit Fun Fact"

	updatedFact, err := h.funFactService.UpdateFunFact(c, req.ID, req)
	if err != nil {
		renderData.Data = fiber.Map{
			"success": false,
			"error":   err.Error(),
		}
		return c.Render("pages/admin/edit-funfact", renderData.ToMap())
	}

	renderData.Data = fiber.Map{
		"success": true,
		"msg":     "Data updated successfully",
		"Fact":    updatedFact,
	}
	return c.Render("pages/admin/edit-funfact", renderData.ToMap())
}
