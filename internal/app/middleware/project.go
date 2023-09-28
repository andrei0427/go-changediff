package middleware

import (
	"fmt"

	"github.com/andrei0427/go-changediff/internal/app/models"
	"github.com/andrei0427/go-changediff/internal/app/services"
	"github.com/andrei0427/go-changediff/internal/data"
	"github.com/gofiber/fiber/v2"
)

func UseProject(c *fiber.Ctx, cacheService *services.CacheService, projectService *services.ProjectService) error {
	curUser := c.Locals("user").(*models.SessionUser)

	userProjectCacheKey := "user-" + fmt.Sprint(curUser.Id) + "project"
	cachedProject, ok := cacheService.Get(userProjectCacheKey)

	if !ok {
		project, _ := projectService.GetProjectForUser(c.Context(), curUser.Id)

		if project != nil {
			cachedProject = project
			cacheService.Set(userProjectCacheKey, cachedProject, nil)
		}
	}

	if cachedProject != nil {
		cachedProject := cachedProject.(*data.Project)
		curUser.Project = cachedProject
	}

	c.Locals("user", curUser)

	return c.Next()
}
