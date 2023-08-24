package handlers

import (
	"strings"

	"github.com/andrei0427/go-changediff/internal/app"
	"github.com/andrei0427/go-changediff/internal/app/middleware"
	"github.com/andrei0427/go-changediff/internal/app/models"
	"github.com/andrei0427/go-changediff/internal/app/services"
	"github.com/gofiber/fiber/v2"
)

type AppHandler struct {
	ProjectService *services.ProjectService
	PostService    *services.PostService
	CDNService     *services.CDNService
}

func NewAppHandler(projectService *services.ProjectService, postService *services.PostService, cdnService *services.CDNService) *AppHandler {
	return &AppHandler{ProjectService: projectService, PostService: postService, CDNService: cdnService}
}

func InitRoutes(app *app.App) {
	appHandler := NewAppHandler(app.ProjectService, app.PostService, app.CDNService)

	app.Fiber.Get("/", appHandler.Home)

	app.Fiber.Use(middleware.UseAuth)
	admin := app.Fiber.Group("/admin")
	admin.Get("/dashboard", appHandler.Dashboard)
	admin.Get("/project", appHandler.GetProject)
	admin.Post("/onboarding", appHandler.PostOnboarding)
	admin.Post("/save-post", appHandler.SavePost)
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

	posts, err := a.PostService.GetPostCountForUser(c.Context(), curUser.Id)
	if err != nil {
		return fiber.NewError(503, "Could not get posts for user")
	}

	if posts == 0 {
		return c.Render("first_post", fiber.Map{"project": project})
	}

	return c.Render("partials/components/dashboard_widget", fiber.Map{"Project": project})
}

func (a *AppHandler) PostOnboarding(c *fiber.Ctx) error {
	curUser := c.Locals("user").(*middleware.SessionUser)

	form := new(models.OnboardingModel)
	if err := c.BodyParser(form); err != nil {
		return c.Render("get_started", fiber.Map{"error": err.Error()})
	}

	errs := make(map[string]string)
	if len(strings.TrimSpace(form.Name)) == 0 {
		errs["Name"] = "Name is required"
	}

	if len(strings.TrimSpace(form.Description)) == 0 {
		errs["Description"] = "Description is required"
	}

	if len(strings.TrimSpace(form.AccentColor)) == 0 {
		errs["AccentColor"] = "Accent color is required"
	}

	if len(errs) > 0 {
		return c.Render("get_started", fiber.Map{"form": form, "errors": errs})
	}

	var fileName *string

	if file, err := c.FormFile("photo"); err == nil {
		uploadedFile, err := a.CDNService.UploadImage(file, 250)
		if err != nil {
			return c.Render("get_started", fiber.Map{"error": err.Error(), "form": form})
		}

		fileName = uploadedFile
	}

	savedProject, err := a.ProjectService.SaveProject(c.Context(), curUser.Id, *form, fileName)
	if err != nil {
		return c.Render("get_started", fiber.Map{"error": err.Error(), "form": form})
	}

	return c.Render("first_post", fiber.Map{"project": savedProject})
}

func (a *AppHandler) SavePost(c *fiber.Ctx) error {
	curUser := c.Locals("user").(*middleware.SessionUser)

	form := new(models.PostModel)
	if err := c.BodyParser(form); err != nil {
		return c.Render("partials/components/post_form", fiber.Map{"error": err.Error()})
	}

	errs := make(map[string]string)
	if len(strings.TrimSpace(form.Title)) == 0 {
		errs["Title"] = "Title is required"
	}

	if len(strings.TrimSpace(form.Content)) == 0 {
		errs["Content"] = "Post content is required"
	}

	if len(errs) > 0 {
		return c.Render("partials/components/post_form", fiber.Map{"form": form, "errors": errs})
	}

	var fileName *string
	if file, err := c.FormFile("banner"); err == nil {
		uploadedFile, err := a.CDNService.UploadImage(file, 250)
		if err != nil {
			return c.Render("partials/components/post_form", fiber.Map{"error": err.Error(), "form": form})
		}

		fileName = uploadedFile
	}

	_, err := a.PostService.InsertPost(c.Context(), *form, fileName, curUser.Id, 1)
	if err != nil {
		return c.Render("partials/components/post_form", fiber.Map{"error": err.Error(), "form": form})
	}

	return c.Render("dashboard", fiber.Map{})
}
