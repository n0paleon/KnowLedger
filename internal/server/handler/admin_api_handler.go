package handler

import (
	"KnowLedger/internal/service"
	"io"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

type AdminApiHandler struct {
	funFactService *service.FunFactService
	mediaService   *service.MediaService
	log            *zap.Logger
}

const MaxFileUpload = 5 * 1024 * 1024 // 5MB

func NewAdminApiHandler(fs *service.FunFactService, ms *service.MediaService, logger *zap.Logger) *AdminApiHandler {
	return &AdminApiHandler{
		funFactService: fs,
		mediaService:   ms,
		log:            logger,
	}
}

func (h *AdminApiHandler) DeleteFunFact(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "id is required",
		})
	}

	if err := h.funFactService.DeleteFact(c, id); err != nil {
		h.log.Error("DeleteFunFact error", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
	})
}

func (h *AdminApiHandler) DeleteTag(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "id is required",
		})
	}

	if err := h.funFactService.DeleteTag(c, id); err != nil {
		h.log.Error("DeleteTag error", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
	})
}

func (h *AdminApiHandler) UploadMedia(c fiber.Ctx) error {
	file, err := c.FormFile("media")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"error": "invalid media file",
		})
	}

	if file.Size > MaxFileUpload {
		return c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"error": "file too big",
		})
	}

	f, err := file.Open()
	if err != nil {
		h.log.Error("UploadMedia file.Open() error", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"error": "invalid media file",
		})
	}
	defer f.Close()

	limitedReader := io.LimitReader(f, int64(MaxFileUpload))

	data, err := io.ReadAll(limitedReader)
	if err != nil {
		h.log.Error("UploadMedia io.ReadAll() error", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"error": err.Error(),
		})
	}

	result, err := h.mediaService.SaveMedia(c, data)
	if err != nil {
		h.log.Error("UploadMedia SaveMedia() error", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(&fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(&fiber.Map{
		"key":  result.Key,
		"url":  result.URL,
		"size": result.SizeHumanized(),
	})
}

func (h *AdminApiHandler) GetTagSuggestions(c fiber.Ctx) error {
	q := c.Query("q")
	if q == "" {
		return c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"error": "invalid tag query",
		})
	}

	tags := h.funFactService.GetTagSuggestions(c, q)

	return c.JSON(tags)
}
