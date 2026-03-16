package handler

import (
	"KnowLedger/internal/service"
	"KnowLedger/pkg/dto"
	"bytes"
	"io"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

type InternalAPIHandler struct {
	log        *zap.Logger
	apiService *service.InternalApiService
}

func NewInternalAPIHandler(log *zap.Logger, apiService *service.InternalApiService) *InternalAPIHandler {
	return &InternalAPIHandler{
		log:        log,
		apiService: apiService,
	}
}

func (h *InternalAPIHandler) UploadMedia(c fiber.Ctx) error {
	response := dto.APIResponse{}

	file, err := c.FormFile("media")
	if err != nil {
		response.Error = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(response)
	}

	if file.Size > MaxFileUpload {
		response.Error = "file too big"
		return c.Status(fiber.StatusBadRequest).JSON(response)
	}

	f, err := file.Open()
	if err != nil {
		h.log.Error("UploadMedia file.Open() error", zap.Error(err))
		response.Error = err.Error()
		return c.Status(fiber.StatusInternalServerError).JSON(response)
	}
	defer f.Close()

	buf := bytes.NewBuffer(make([]byte, 0, file.Size))

	if _, err := io.Copy(buf, f); err != nil {
		h.log.Error("UploadMedia io.Copy() error", zap.Error(err))
		response.Error = err.Error()
		return c.Status(fiber.StatusInternalServerError).JSON(response)
	}

	result, err := h.apiService.SaveMedia(c, buf.Bytes())
	if err != nil {
		h.log.Error("UploadMedia SaveMedia() error", zap.Error(err))
		response.Error = err.Error()
		return c.Status(fiber.StatusInternalServerError).JSON(response)
	}

	response.Data = result
	return c.Status(fiber.StatusOK).JSON(response)
}

func (h *InternalAPIHandler) CreateFunFact(c fiber.Ctx) error {
	response := dto.APIResponse{}
	var req []*dto.CreateFunFactAPIRequest

	if err := c.Bind().JSON(&req); err != nil || len(req) == 0 {
		response.Error = "invalid payload"
		return c.Status(fiber.StatusBadRequest).JSON(response)
	}

	if len(req) > 2000 {
		response.Error = "payload too big"
		return c.Status(fiber.StatusBadRequest).JSON(response)
	}

	total, err := h.apiService.CreateFunFact(c, req)
	if err != nil {
		h.log.Error("CreateFunFact error", zap.Error(err))
		response.Error = err.Error()
	}
	response.Data = fiber.Map{
		"total": total,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

func (h *InternalAPIHandler) AuthMiddleware(c fiber.Ctx) error {
	response := dto.APIResponse{}
	accountID := c.Get("X-Account-Id")
	apiKey := c.Get("X-Api-Key")

	if accountID == "" || apiKey == "" {
		response.Error = "unauthorized"
		return c.Status(fiber.StatusUnauthorized).JSON(response)
	}

	if err := h.apiService.ValidateApiKey(c, accountID, apiKey); err != nil {
		response.Error = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(response)
	}

	h.log.Info("user authenticated to use internal api",
		zap.String("accountId", accountID),
		zap.String("apiKey", apiKey),
	)
	return c.Next()
}
