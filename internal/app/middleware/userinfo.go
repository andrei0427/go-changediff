package middleware

import (
	"encoding/json"

	"github.com/andrei0427/go-changediff/internal/app/models"
	"github.com/gofiber/fiber/v2"
)

func UseUserInfo(c *fiber.Ctx) error {
	userInfoStr := c.Cookies("user_info")

	var data models.UserInfo
	c.Locals("userInfo", &data)

	if len(userInfoStr) == 0 {
		return c.Next()
	}

	json.Unmarshal([]byte(userInfoStr), &data)

	return c.Next()
}
