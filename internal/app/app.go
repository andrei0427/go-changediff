package app

import (
	"os"

	"github.com/andrei0427/go-changediff/internal/app/services"
	"github.com/andrei0427/go-changediff/internal/data"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/django/v3"
)

type App struct {
	DB    *data.Queries
	Fiber *fiber.App

	ProjectService *services.ProjectService
}

func NewApp() *App {
	dbConn := data.InitPostgresDb()

	engine := django.New("web", ".html")
	engine.Reload(os.Getenv("ENV") == "development")
	// engine.AddFuncMap()

	fiber := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		Views:                 engine,
		// ErrorHandler:          handlers.HandleError,
	})

	fiber.Static("/static", "web/static")

	projectService := services.NewProjectService(dbConn)

	return &App{
		DB:             dbConn,
		Fiber:          fiber,
		ProjectService: projectService,
	}
}
