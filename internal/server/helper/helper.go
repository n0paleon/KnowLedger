package helper

import "github.com/gofiber/fiber/v3"

const (
	UserIDKey = "user_id"
)

func SetUserID(c fiber.Ctx, userID string) {
	c.Locals(UserIDKey, userID)
}

func GetUserID(c fiber.Ctx) string {
	userID, ok := c.Locals(UserIDKey).(string)
	if !ok {
		return ""
	}
	return userID
}
