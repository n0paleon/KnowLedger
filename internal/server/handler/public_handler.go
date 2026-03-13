package handler

import (
	"KnowLedger/internal/service"
	"KnowLedger/pkg/dto"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

type PublicHandler struct {
	funFactService *service.FunFactService
	log            *zap.Logger
}

func NewPublicHandler(fs *service.FunFactService, logger *zap.Logger) *PublicHandler {
	return &PublicHandler{
		funFactService: fs,
		log:            logger,
	}
}

func (h *PublicHandler) PublicShowIndex(c fiber.Ctx) error {
	fact, err := h.funFactService.GetOneRandomFunFact(c)
	if err != nil {
		h.log.Error("failed to get one random fun fact", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	//return c.JSON(fact)

	return c.Render("pages/index", dto.RenderData{
		Title: "Daily Fun Fact | KnowLedger",
		Data: fiber.Map{
			"Fact": fact,
		},
	}.ToMap(), "")
}
