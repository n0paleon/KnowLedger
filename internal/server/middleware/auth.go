package middleware

import (
	"KnowLedger/internal/server/helper"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/session"
)

func RequireAuth(c fiber.Ctx) error {
	sess := session.FromContext(c)
	if sess == nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	// Check if user is authenticated
	if sess.Get("authenticated") != true {
		return c.Redirect().Route("Show Admin Login")
	}

	userID, ok := sess.Get("user_id").(string)
	if userID == "" || !ok {
		return c.Redirect().Route("")
	}
	helper.SetUserID(c, userID)

	return c.Next()
}
