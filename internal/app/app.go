package app

import (
	"log"
	"os"

	"github.com/andrei0427/go-changediff/internal/app/services"
	"github.com/andrei0427/go-changediff/internal/data"
	"github.com/andrei0427/go-changediff/web/views"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/django/v3"
)

type App struct {
	DB    *data.Queries
	Fiber *fiber.App

	AuthorService  *services.AuthorService
	CDNService     *services.CDNService
	ProjectService *services.ProjectService
	PostService    *services.PostService
	LabelService   *services.LabelService
	CacheService   *services.CacheService
	RoadmapService *services.RoadmapService
}

func NewApp() *App {
	db := data.InitPostgresDb()
	dbConn := data.New(db)

	engine := django.New("web/views", ".html")
	engine.AddFuncMap(views.HelperFuncMap)
	engine.Reload(os.Getenv("ENV") == "development")
	// engine.AddFuncMap()

	fiber := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		Views:                 engine,
		ErrorHandler:          handleError,
		PassLocalsToViews:     true,
	})

	fiber.Static("/static", "web/static")

	authorService := services.NewAuthorService(dbConn)
	projectService := services.NewProjectService(dbConn)
	postService := services.NewPostService(dbConn, db)
	labelService := services.NewLabelService(dbConn)
	roadmapService := services.NewRoadmapService(dbConn, db)
	cdnService := services.NewCDNService()
	cacheService := services.NewCacheService()

	return &App{
		DB:             dbConn,
		Fiber:          fiber,
		AuthorService:  authorService,
		ProjectService: projectService,
		PostService:    postService,
		CDNService:     cdnService,
		LabelService:   labelService,
		CacheService:   cacheService,
		RoadmapService: roadmapService,
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
