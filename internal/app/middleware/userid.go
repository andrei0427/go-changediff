package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func UseUserId(c *fiber.Ctx) error {
	userId := c.Cookies("user_id")

	userUuid, err := uuid.Parse(userId)

	if err == nil {
		c.Locals("userUuid", &userUuid)
	}

	return c.Next()
}
