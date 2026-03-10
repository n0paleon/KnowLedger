package handler

import (
	"KnowLedger/internal/repository"
	"KnowLedger/pkg/dto"
	"KnowLedger/pkg/utils"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/session"
	"go.uber.org/zap"
)

type AuthHandler struct {
	adminRepository *repository.AdminRepository
	log             *zap.Logger
}

func NewAuthHandler(adminRepo *repository.AdminRepository, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		adminRepository: adminRepo,
		log:             logger,
	}
}

func (h *AuthHandler) ShowLogin(c fiber.Ctx) error {
	return c.Render("pages/admin/login", dto.RenderData{
		Title: "Login",
	}.ToMap(), "layouts/auth")
}

func (h *AuthHandler) Login(c fiber.Ctx) error {
	sess := session.FromContext(c)
	username := c.FormValue("username")
	password := c.FormValue("password")

	renderData := dto.RenderData{
		Title: "Login",
	}

	admin, err := h.adminRepository.FindByUsername(c, username)
	if err != nil {
		renderData.Error = "invalid username or password"
		return c.Render("pages/admin/login", renderData.ToMap(), "layouts/auth")
	}

	if ok := utils.CheckPasswordHash(admin.Password, password); !ok {
		renderData.Error = "invalid username or password"
		return c.Render("pages/admin/login", renderData.ToMap(), "layouts/auth")
	}

	if err := sess.Regenerate(); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "session error",
		})
	}

	sess.Set("user_id", admin.ID)
	sess.Set("username", admin.Username)
	sess.Set("authenticated", true)

	return c.Redirect().Route("Show Admin Dashboard")
}
