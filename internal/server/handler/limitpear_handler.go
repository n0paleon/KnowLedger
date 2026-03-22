package handler

import (
	"KnowLedger/internal/service"
	"KnowLedger/pkg/dto"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

type LimitPearHandler struct {
	log            *zap.Logger
	funFactService *service.FunFactService
}

func NewLimitPearHandler(log *zap.Logger, f *service.FunFactService) *LimitPearHandler {
	return &LimitPearHandler{
		log:            log,
		funFactService: f,
	}
}

func (h *LimitPearHandler) GetOneRandomFunFact(c fiber.Ctx) error {
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
