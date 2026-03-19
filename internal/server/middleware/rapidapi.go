package middleware

import (
	"KnowLedger/internal/config"
	"KnowLedger/pkg/dto"

	"github.com/gofiber/fiber/v3"
)

type RapidAPIMiddleware struct {
	proxySecret string
	isDev       bool // jika iya, ProxyAuthMiddleware akan diabaikan
}

func NewRapidAPIMiddleware(cfg *config.Config) *RapidAPIMiddleware {
	return &RapidAPIMiddleware{
		proxySecret: cfg.Api.RapidAPIProxySecret,
		isDev:       cfg.App.Dev,
	}
}

// ProxyAuthMiddleware memverifikasi apakah client mengirimkan request header yang dibutuhkan sebelum request diproses
func (m *RapidAPIMiddleware) ProxyAuthMiddleware(c fiber.Ctx) error {
	// jika dev mode aktif, langsung skip
	if m.isDev {
		return c.Next()
	}

	reqProxySecret := c.Get("X-RapidAPI-Proxy-Secret")
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
