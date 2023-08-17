package handlers

import (
	"github.com/andrei0427/go-changediff/internal/app"
	"github.com/andrei0427/go-changediff/internal/app/middleware"
	"github.com/andrei0427/go-changediff/internal/app/services"
	"github.com/gofiber/fiber/v2"
)

type AppHandler struct {
	ProjectService *services.ProjectService
	CDNService     *services.CDNService
}

func NewAppHandler(projectService *services.ProjectService, cdnService *services.CDNService) *AppHandler {
	return &AppHandler{ProjectService: projectService, CDNService: cdnService}
}

func InitRoutes(app *app.App) {
	appHandler := NewAppHandler(app.ProjectService, app.CDNService)

	app.Fiber.Get("/", appHandler.Home)

	app.Fiber.Use(middleware.UseAuth)
	app.Fiber.Get("/dashboard", appHandler.Dashboard)
	app.Fiber.Get("/project", appHandler.GetProject)

}

func (a *AppHandler) Home(c *fiber.Ctx) error {
	return c.Render("index", fiber.Map{})
}

func (a *AppHandler) Dashboard(c *fiber.Ctx) error {
	return c.Render("dashboard", fiber.Map{})
}

func (a *AppHandler) GetProject(c *fiber.Ctx) error {
	curUser := c.Locals("user").(*middleware.SessionUser)

	project, err := a.ProjectService.GetProjectForUser(c.Context(), curUser.Id)
	if err != nil {
		return fiber.NewError(503, "Could not get project for user")
	}

	if project == nil {
		return c.Render("get_started", fiber.Map{})

	}

	return c.Render("partials/components/dashboard_widget", fiber.Map{"Project": project})
}
