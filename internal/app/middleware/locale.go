package middleware

import (
	"github.com/gofiber/fiber/v2"
)

func UseLocale(c *fiber.Ctx) error {
	userTz := c.Cookies("user_tz")
	c.Locals("timezone", userTz)

	return c.Next()
}
