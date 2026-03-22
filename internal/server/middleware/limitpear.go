package middleware

import (
	"KnowLedger/internal/config"
	"KnowLedger/pkg/dto"

	"github.com/gofiber/fiber/v3"
)

type LimitPearMiddleware struct {
	proxySecret string
	isDev       bool
}

func NewLimitPearMiddleware(cfg *config.Config) *LimitPearMiddleware {
	return &LimitPearMiddleware{
		proxySecret: cfg.Api.LimitPearProxySecret,
		isDev:       cfg.App.Dev,
	}
}

func (m *LimitPearMiddleware) ProxyAuthMiddleware(c fiber.Ctx) error {
	if m.isDev {
		return c.Next()
	}

	reqProxySecret := c.Get("X-LIMITPEAR-PROXY-SECRET")
	if reqProxySecret == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.APIResponse{
			Error: "unauthorized",
		})
	}

	if reqProxySecret != m.proxySecret {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.APIResponse{
			Error: "invalid proxy secret",
		})
	}

	return c.Next()
}

func (m *LimitPearMiddleware) HealthCheckMiddleware(c fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(dto.APIResponse{
		Data: "KnowLedger is ready!",
	})
}
