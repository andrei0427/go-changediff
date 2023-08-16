package handlers

import (
	"github.com/andrei0427/go-changediff/internal/app"
	"github.com/andrei0427/go-changediff/internal/app/services"
	"github.com/gofiber/fiber/v2"
)

type AppHandler struct {
	ProjectService *services.ProjectService
}

func NewAppHandler(projectService *services.ProjectService) *AppHandler {
	return &AppHandler{ProjectService: projectService}
}

func InitRoutes(app *app.App) {
	appHandler := NewAppHandler(app.ProjectService)

	app.Fiber.Get("/", appHandler.Home)
}

func (a *AppHandler) Home(c *fiber.Ctx) error {
	return c.Render("views/index", fiber.Map{})
}
