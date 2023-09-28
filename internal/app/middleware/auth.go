package middleware

import (
	"fmt"
	"os"

	"github.com/andrei0427/go-changediff/internal/app/models"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

func UseAuth(c *fiber.Ctx) error {
	tokenString := c.Cookies("authUser")
	userTz := c.Cookies("user_tz")

	if tokenString == "" {
		return fiber.NewError(fiber.StatusUnauthorized)
	}

	token, err := jwt.ParseWithClaims(tokenString, &models.AuthTokenClaim{}, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(os.Getenv("SUPABASE_API_SECRET")), nil
	})

	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized)
	}

	// validate the essential claims
	if !token.Valid {
		return fiber.NewError(fiber.StatusUnauthorized)
	}

	claims := token.Claims.(*models.AuthTokenClaim)

	if userId, err := uuid.Parse(claims.Subject); err == nil {
		sessionUser := &models.SessionUser{
			Id:       userId,
			Email:    claims.Email,
			Metadata: claims.Metadata,
			Timezone: userTz,
		}

		c.Locals("user", sessionUser)
	}

	return c.Next()
}
