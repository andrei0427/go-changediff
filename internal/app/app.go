package app

import (
	"log"
	"os"

	"github.com/andrei0427/go-changediff/internal/app/services"
	"github.com/andrei0427/go-changediff/internal/data"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/django/v3"
)

type App struct {
	DB    *data.Queries
	Fiber *fiber.App

	CDNService     *services.CDNService
	ProjectService *services.ProjectService
}

func NewApp() *App {
	dbConn := data.InitPostgresDb()

	engine := django.New("web/views", ".html")
	engine.Reload(os.Getenv("ENV") == "development")
	// engine.AddFuncMap()

	fiber := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		Views:                 engine,
		ErrorHandler:          handleError,
		PassLocalsToViews:     true,
	})

	fiber.Static("/static", "web/static")

	projectService := services.NewProjectService(dbConn)
	cdnService := services.NewCDNService()

	return &App{
		DB:             dbConn,
		Fiber:          fiber,
		ProjectService: projectService,
		CDNService:     cdnService,
	}
}

func handleError(c *fiber.Ctx, err error) error {
	e, ok := err.(*fiber.Error)

	log.Println(e.Message, e.Code)

	if ok {
		return c.Render("error", fiber.Map{"Code": e.Code, "Message": e.Message})
	}

	return c.Render("error", fiber.Map{"Code": fiber.StatusInternalServerError, "Message": err.Error()})
}