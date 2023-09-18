package middleware

import (
	"fmt"
	"os"

	"github.com/andrei0427/go-changediff/internal/app/services"
	"github.com/andrei0427/go-changediff/internal/data"
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
	Timezone string
	Project  *data.Project
}

func UseAuth(c *fiber.Ctx, cacheService *services.CacheService, projectService *services.ProjectService) error {
	tokenString := c.Cookies("authUser")
	userTz := c.Cookies("user_tz")

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
		return fiber.NewError(fiber.StatusUnauthorized)
	}

	// validate the essential claims
	if !token.Valid {
		return fiber.NewError(fiber.StatusUnauthorized)
	}

	claims := token.Claims.(*AuthTokenClaim)

	if userId, err := uuid.Parse(claims.Subject); err == nil {
		cacheKey := "user-" + fmt.Sprint(userId) + "project"
		cachedProject, ok := cacheService.Get(cacheKey)

		if !ok {
			project, _ := projectService.GetProjectForUser(c.Context(), userId)

			if project != nil {
				cachedProject = project
				cacheService.Set(cacheKey, cachedProject, nil)
			}
		}

		sessionUser := &SessionUser{
			Id:       userId,
			Email:    claims.Email,
			Metadata: claims.Metadata,
			Timezone: userTz,
		}

		if cachedProject != nil {
			cachedProject := cachedProject.(*data.Project)
			sessionUser.Project = cachedProject
		}

		c.Locals("user", sessionUser)
	}

	return c.Next()
}
