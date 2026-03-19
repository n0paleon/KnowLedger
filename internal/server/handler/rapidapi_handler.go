package handler

import (
	"KnowLedger/internal/service"
	"KnowLedger/pkg/dto"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

type RapidAPIHandler struct {
	log            *zap.Logger
	funFactService *service.FunFactService
}

func NewRapidAPIHandler(log *zap.Logger, fs *service.FunFactService) *RapidAPIHandler {
	return &RapidAPIHandler{
		log:            log,
		funFactService: fs,
	}
}

func (h *RapidAPIHandler) GetOneRandomFunFact(c fiber.Ctx) error {
	response := new(dto.APIResponse)

	fact, err := h.funFactService.GetOneRandomFunFact(c)
	if err != nil {
		h.log.Error("Error getting random fun fact", zap.Error(err))
		response.Error = "something went wrong"
		return c.Status(http.StatusServiceUnavailable).JSON(response)
	}

	factDto := dto.ConvertFactModelToDto(fact)
	response.Data = factDto
	return c.Status(http.StatusOK).JSON(response)
}
