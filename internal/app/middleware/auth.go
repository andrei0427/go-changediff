package middleware

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

type UserMetadata struct {
	AvatarUrl string `json:"avatar_url"`
	FullName  string `json:"full_name"`
}

type AuthTokenClaim struct {
	jwt.StandardClaims

	Email    string       `json:"email"`
	Metadata UserMetadata `json:"user_metadata"`
}

type SessionUser struct {
	Id       uuid.UUID
	Email    string
	Metadata UserMetadata
}

func UseAuth(c *fiber.Ctx) error {
	tokenString := c.Cookies("authUser")

	if tokenString == "" {
		return fiber.NewError(fiber.StatusUnauthorized)
	}

	token, err := jwt.ParseWithClaims(tokenString, &AuthTokenClaim{}, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(os.Getenv("SUPABASE_API_SECRET")), nil
	})

	if err != nil {
		log.Fatal(err)
		return fiber.NewError(fiber.StatusUnauthorized)
	}

	// validate the essential claims
	if !token.Valid {
		log.Println("pee")
		return fiber.NewError(fiber.StatusUnauthorized)
	}

	claims := token.Claims.(*AuthTokenClaim)

	if userId, err := uuid.Parse(claims.Subject); err == nil {
		c.Locals("user", &SessionUser{
			Id:       userId,
			Email:    claims.Email,
			Metadata: claims.Metadata,
		})
	}

	return c.Next()
}
