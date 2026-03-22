package middleware

import (
	"KnowLedger/internal/config"
	"KnowLedger/pkg/dto"

	"github.com/gofiber/fiber/v3"
)

type APIMarketplaceMiddleware struct {
	isDev                bool
	rapidApiProxySecret  string
	limitPearProxySecret string
}

func NewAPIMarketplaceMiddleware(cfg *config.Config) *APIMarketplaceMiddleware {
	return &APIMarketplaceMiddleware{
		isDev:                cfg.App.Dev,
		rapidApiProxySecret:  cfg.Api.RapidAPIProxySecret,
		limitPearProxySecret: cfg.Api.LimitPearProxySecret,
	}
}

func (m *APIMarketplaceMiddleware) Auth(c fiber.Ctx) error {
	// Skip auth in development
	if m.isDev {
		return c.Next()
	}

	rapidApiSecret := c.Get("X-RapidAPI-Proxy-Secret")
	limitPearSecret := c.Get("X-LIMITPEAR-PROXY-SECRET")

	switch {
	case rapidApiSecret != "":
		if rapidApiSecret != m.rapidApiProxySecret {
			return c.Status(fiber.StatusUnauthorized).JSON(dto.APIResponse{
				Error: "invalid rapidapi proxy secret",
			})
		}
		c.Locals("marketplace", "rapidapi")

	case limitPearSecret != "":
		if limitPearSecret != m.limitPearProxySecret {
			return c.Status(fiber.StatusUnauthorized).JSON(dto.APIResponse{
				Error: "invalid limitpear proxy secret",
			})
		}
		c.Locals("marketplace", "limitpear")

	default:
		return c.Status(fiber.StatusUnauthorized).JSON(dto.APIResponse{
			Error: "missing proxy secret",
		})
	}

	return c.Next()
}
