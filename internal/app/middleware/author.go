package middleware

import (
	"fmt"

	"github.com/andrei0427/go-changediff/internal/app/models"
	"github.com/andrei0427/go-changediff/internal/app/services"
	"github.com/andrei0427/go-changediff/internal/data"
	"github.com/gofiber/fiber/v2"
)

func UseAuthor(c *fiber.Ctx, cacheService *services.CacheService, authorService *services.AuthorService) error {
	curUser := c.Locals("user").(*models.SessionUser)

	if curUser.Project == nil {
		return c.Redirect("/admin/dashboard")
	}

	userAuthorCacheKey := "user-" + fmt.Sprint(curUser.Id) + "author"
	cachedAuthor, ok := cacheService.Get(userAuthorCacheKey)

	if !ok {
		author, _ := authorService.GetAuthorByUser(c.Context(), curUser.Id)

		if author != nil {
			cachedAuthor = author
			cacheService.Set(userAuthorCacheKey, cachedAuthor, nil)
		}
	}

	if cachedAuthor != nil {
		cachedAuthor := cachedAuthor.(*data.GetAuthorByUserRow)
		curUser.Author = cachedAuthor
	}

	c.Locals("user", curUser)

	return c.Next()
}
