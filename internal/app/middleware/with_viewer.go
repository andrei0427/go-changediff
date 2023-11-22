package middleware

import (
	"fmt"

	"github.com/andrei0427/go-changediff/internal/app/models"
	"github.com/andrei0427/go-changediff/internal/app/services"
	"github.com/andrei0427/go-changediff/internal/data"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func UseWithViewer(c *fiber.Ctx,
	cacheService *services.CacheService,
	projectService *services.ProjectService,
	viewerService *services.ViewerService,
) error {
	projectKey := c.Params("key")
	userUuid := c.Locals("userUuid").(*uuid.UUID)
	userLocale := c.Locals("timezone").(string)
	userInfo := c.Locals("userInfo").(*models.UserInfo)

	if userUuid == nil {
		return fiber.NewError(fiber.StatusBadRequest, "User uuid not found")
	}

	var userId *string
	if userInfo != nil && userInfo.ID != nil {
		strUserID := string(*userInfo.ID)
		userId = &strUserID
	}

	viewerCacheKey := fmt.Sprint("viewer-", projectKey, "-", userUuid.String())
	cachedViewer, ok := cacheService.Get(viewerCacheKey)

	if !ok {
		viewer, _ := viewerService.GetViewer(c.Context(), *userUuid, userId)

		if viewer == nil {
			project, err := projectService.GetProjectByKey(c.Context(), cacheService, projectKey)
			if err != nil {
				return fiber.NewError(fiber.StatusBadRequest, "project not found")
			}

			savedViewer, err := viewerService.SaveViewer(c.Context(),
				*userUuid,
				c.IP(),
				c.Get("User-Agent"),
				userLocale,
				userInfo,
				project.ID,
			)

			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, "error saving viewer")
			}

			viewer = &savedViewer
		}

		cachedViewer = *viewer
		cacheService.Set(viewerCacheKey, cachedViewer, nil)
	}

	viewer := cachedViewer.(data.Viewer)
	if (!viewer.UserID.Valid && userInfo.ID != nil) ||
		(!viewer.UserName.Valid && userInfo.Name != nil) ||
		(!viewer.UserEmail.Valid && userInfo.Email != nil) ||
		(!viewer.UserRole.Valid && userInfo.Role != nil) ||
		(!viewer.UserData.Valid && userInfo.Info != nil) {
		updatedViewer, err := viewerService.SaveViewer(c.Context(), *userUuid, c.IP(), c.Get("User-Agent"), userLocale, userInfo, viewer.ProjectID)

		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "error saving viewer")
		}

		cachedViewer = &updatedViewer
		cacheService.Set(viewerCacheKey, updatedViewer, nil)
	}

	c.Locals("viewer", cachedViewer)

	return c.Next()
}
