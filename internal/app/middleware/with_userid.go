package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func UseWithUserId(c *fiber.Ctx) error {
	userUuid := c.Locals("userUuid").(*uuid.UUID)

	if userUuid == nil {
		return fiber.NewError(fiber.StatusBadRequest, "User uuid not found")

	}

	return c.Next()
}
