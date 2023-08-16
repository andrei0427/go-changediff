package handlers

import (
	"github.com/andrei0427/go-changediff/internal/app"
	"github.com/andrei0427/go-changediff/internal/app/middleware"
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

	app.Fiber.Use(middleware.UseAuth)
	app.Fiber.Get("/dashboard", appHandler.Dashboard)
}

func (a *AppHandler) Home(c *fiber.Ctx) error {
	return c.Render("index", fiber.Map{})
}

func (a *AppHandler) Dashboard(c *fiber.Ctx) error {
	return c.Render("dashboard", fiber.Map{})
}
