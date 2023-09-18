package handlers

import (
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/andrei0427/go-changediff/internal/app"
	"github.com/andrei0427/go-changediff/internal/app/middleware"
	"github.com/andrei0427/go-changediff/internal/app/models"
	"github.com/andrei0427/go-changediff/internal/app/services"
	"github.com/andrei0427/go-changediff/internal/data"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/exp/slices"
)

type AppHandler struct {
	ProjectService *services.ProjectService
	PostService    *services.PostService
	CDNService     *services.CDNService
	LabelService   *services.LabelService
	CacheService   *services.CacheService
}

func NewAppHandler(projectService *services.ProjectService,
	postService *services.PostService,
	cdnService *services.CDNService,
	labelService *services.LabelService,
	cacheService *services.CacheService) *AppHandler {
	return &AppHandler{ProjectService: projectService,
		PostService:  postService,
		CDNService:   cdnService,
		LabelService: labelService,
		CacheService: cacheService,
	}
}

func InitRoutes(app *app.App) {
	appHandler := NewAppHandler(app.ProjectService, app.PostService, app.CDNService, app.LabelService, app.CacheService)
	app.Fiber.Get("/", appHandler.Home)

	widget := app.Fiber.Group("/widget")
	widget.Use(middleware.UseLocale)
	widget.Get("/:key", appHandler.WidgetHome)

	changelog := widget.Group("/changelog")
	changelog.Get("/:key", appHandler.WidgetChangelog)
	changelog.Get("/posts/:key/:pageNo?", appHandler.WidgetChangelogPosts)

	widget.Get("/roadmap/:key", appHandler.WidgetRoadmap)
	widget.Get("/feedback/:key", appHandler.WidgetFeedback)

	admin := app.Fiber.Group("/admin")
	admin.Use(func(c *fiber.Ctx) error {
		return middleware.UseAuth(c, appHandler.CacheService, appHandler.ProjectService)
	})
	admin.Get("/dashboard", appHandler.Dashboard)
	admin.Get("/project", appHandler.GetProject)
	admin.Post("/onboarding", appHandler.PostOnboarding)

	posts := admin.Group("/posts")
	posts.Get("/", appHandler.Posts)
	posts.Get("/compose/:id?", appHandler.ComposePost)
	posts.Post("/save", appHandler.SavePost)
	posts.Get("/load", appHandler.LoadPosts)
	posts.Delete("/delete/:id", appHandler.DeletePost)
	posts.Delete("/confirm-delete/:id", appHandler.ConfirmDeletePost)

	settings := admin.Group("/settings")
	settings.Get("/", appHandler.Settings)
	settings.Get("/tab", appHandler.SettingsTab)
	settings.Post("/general/save", appHandler.SaveSettingsGeneralTab)
	settings.Post("/labels/save", appHandler.SaveSettingsLabelsTab)
	settings.Delete("/labels/delete/:id", appHandler.DeleteSettingsLabel)
	settings.Get("/labels/confirm-delete/:id", appHandler.ConfirmDeleteLabel)
	settings.Get("/labels/new", appHandler.NewSettingsLabel)
}

// Public Routes

func (a *AppHandler) Home(c *fiber.Ctx) error {
	return c.Render("index", fiber.Map{})
}

func (a *AppHandler) WidgetHome(c *fiber.Ctx) error {
	key := c.Params("key")
	project, err := a.ProjectService.GetProjectByKey(c.Context(), key)
	if err != nil {
		return fiber.NewError(404, "Project not found")
	}

	logoUrl := "/static/logo.png"
	if project.LogoUrl.Valid {
		logoUrl = project.LogoUrl.String
	}

	return c.Render("widget/index", fiber.Map{"Project": project, "LogoUrl": logoUrl, "activeTab": "changelog"})
}

func (a *AppHandler) WidgetChangelog(c *fiber.Ctx) error {
	key := c.Params("key")
	project, err := a.ProjectService.GetProjectByKey(c.Context(), key)
	if err != nil {
		return fiber.NewError(404, "Project not found")
	}

	return c.Render("widget/tabs/changelog", fiber.Map{"Project": project})
}

func (a *AppHandler) WidgetChangelogPosts(c *fiber.Ctx) error {
	key := c.Params("key")
	paramPageNo, _ := c.ParamsInt("pageNo")
	pageNo := 1
	if paramPageNo > 0 {
		pageNo = paramPageNo
	}

	posts, err := a.PostService.GetPublishedPagedPosts(c.Context(), key, int32(pageNo))
	if err != nil {
		return fiber.NewError(503, "Error fetching posts")
	}

	return c.Render("widget/components/posts", fiber.Map{"Posts": posts})
}

func (a *AppHandler) WidgetRoadmap(c *fiber.Ctx) error {
	key := c.Params("key")
	project, err := a.ProjectService.GetProjectByKey(c.Context(), key)
	if err != nil {
		return fiber.NewError(404, "Project not found")
	}

	return c.Render("widget/tabs/roadmap", fiber.Map{"Project": project})
}

func (a *AppHandler) WidgetFeedback(c *fiber.Ctx) error {
	key := c.Params("key")
	project, err := a.ProjectService.GetProjectByKey(c.Context(), key)
	if err != nil {
		return fiber.NewError(404, "Project not found")
	}

	return c.Render("widget/tabs/feedback", fiber.Map{"Project": project})
}

// Protected Routes

func (a *AppHandler) Dashboard(c *fiber.Ctx) error {
	return c.Render("dashboard", fiber.Map{})
}

func (a *AppHandler) Posts(c *fiber.Ctx) error {
	return c.Render("posts", fiber.Map{})
}

func (a *AppHandler) Settings(c *fiber.Ctx) error {
	return c.Render("settings", fiber.Map{})
}

func (a *AppHandler) SettingsTab(c *fiber.Ctx) error {
	curUser := c.Locals("user").(*middleware.SessionUser)
	tabName := c.Query("hash", "general")

	switch tabName {
	case "general":
		return c.Render("partials/components/settings/general_tab", fiber.Map{"Form": curUser.Project})

	case "labels":
		labels, err := a.LabelService.GetLabels(c.Context(), curUser.Project.ID)

		var message string
		if err != nil {
			message = err.Error()
		}

		return c.Render("partials/components/settings/labels_tab", fiber.Map{"Labels": labels, "Message": message})

	default:
		return c.Render("partials/components/settings/general_tab", fiber.Map{"Form": curUser.Project})
	}
}

func (a *AppHandler) NewSettingsLabel(c *fiber.Ctx) error {
	curUser := c.Locals("user").(*middleware.SessionUser)
	return c.Render("partials/components/settings/label_row", fiber.Map{"label": data.Label{Label: "", Color: curUser.Project.AccentColor}})
}

func (a *AppHandler) ConfirmDeleteLabel(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(400, "Invalid id parameter supplied")
	}

	return c.Render("partials/components/delete_confirm_modal", fiber.Map{"Title": "Confirm deletion",
		"Body":           "Are you sure you want to delete this label?",
		"EndpointUri":    "/admin/settings/labels/delete/" + fmt.Sprint(id),
		"TargetSelector": "#tab-content",
		"Swap":           "innerHtml",
	})
}

func (a *AppHandler) DeleteSettingsLabel(c *fiber.Ctx) error {
	viewPath := "partials/components/settings/labels_tab"
	curUser := c.Locals("user").(*middleware.SessionUser)
	currentLabels, err := a.LabelService.GetLabels(c.Context(), curUser.Project.ID)
	if err != nil {
		return c.Render(viewPath, fiber.Map{"Error": err.Error()})
	}

	labelIdToDelete, err := c.ParamsInt("id")
	if err != nil {
		return c.Render(viewPath, fiber.Map{"Labels": currentLabels, "Error": err.Error()})
	}

	err = a.LabelService.DeleteLabel(c.Context(), int32(labelIdToDelete), curUser.Project.ID)
	if err != nil {
		return c.Render(viewPath, fiber.Map{"Labels": currentLabels, "Error": err.Error()})
	}

	labelIdxToDelete := slices.IndexFunc(currentLabels, func(l data.Label) bool {
		return l.ID == int32(labelIdToDelete)
	})

	updatedLabels := make([]data.Label, 0)
	updatedLabels = append(updatedLabels, currentLabels[:labelIdxToDelete]...)
	updatedLabels = append(updatedLabels, currentLabels[labelIdxToDelete+1:]...)

	return c.Render(viewPath, fiber.Map{"Labels": updatedLabels, "Success": true, "Message": "Label successfully deleted"})
}

func (a *AppHandler) SaveSettingsLabelsTab(c *fiber.Ctx) error {
	viewPath := "partials/components/settings/labels_tab"
	curUser := c.Locals("user").(*middleware.SessionUser)
	currentLabels, err := a.LabelService.GetLabels(c.Context(), curUser.Project.ID)
	if err != nil {
		return c.Render(viewPath, fiber.Map{"Error": err.Error()})
	}

	form := new(models.LabelModel)
	if err := c.BodyParser(form); err != nil {
		return c.Render(viewPath, fiber.Map{"Labels": currentLabels, "Error": err.Error()})
	}

	if len(strings.TrimSpace(form.Label)) == 0 {
		return c.Render(viewPath, fiber.Map{"Error": "Label is required"})
	}

	if len(strings.TrimSpace(form.Color)) == 0 {
		return c.Render(viewPath, fiber.Map{"Error": "Color is required"})
	}

	savedLabel, err := a.LabelService.SaveLabel(c.Context(), *form, curUser.Project.ID)
	if err != nil {
		return c.Render(viewPath, fiber.Map{"Labels": currentLabels, "Error": err.Error()})
	}

	if form.ID == nil {
		currentLabels = append(currentLabels, savedLabel)
	} else {
		labelIdx := slices.IndexFunc(currentLabels, func(l data.Label) bool {
			return l.ID == *form.ID
		})

		if labelIdx > -1 {
			currentLabels[labelIdx] = savedLabel
		}
	}

	return c.Render(viewPath, fiber.Map{"Labels": currentLabels, "Success": true, "Message": "Label saved successfully."})

}

func (a *AppHandler) SaveSettingsGeneralTab(c *fiber.Ctx) error {
	viewPath := "partials/components/settings/general_tab"
	curUser := c.Locals("user").(*middleware.SessionUser)

	form := new(models.ProjectModel)
	if err := c.BodyParser(form); err != nil {
		return c.Render(viewPath, fiber.Map{"Error": err.Error()})
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
		return c.Render(viewPath, fiber.Map{"Form": form, "Errors": errs})
	}

	var fileName *string

	if file, err := c.FormFile("photo"); err == nil {
		uploadedFile, err := a.CDNService.UploadImage(file, 250)
		if err != nil {
			return c.Render(viewPath, fiber.Map{"Error": err.Error(), "Form": form})
		}

		fileName = uploadedFile
	} else if curUser.Project.LogoUrl.Valid {
		fileName = &curUser.Project.LogoUrl.String
	}

	savedProject, err := a.ProjectService.SaveProject(c.Context(), curUser.Id, *form, fileName)
	if err != nil {
		return c.Render(viewPath, fiber.Map{"Error": err.Error(), "Form": form})
	}

	a.CacheService.Set("user-"+fmt.Sprint(curUser.Id)+"project", &savedProject, nil)

	return c.Render(viewPath, fiber.Map{"Form": savedProject, "Success": true, "Message": "Settings saved successfully!"})
}

func (a *AppHandler) ComposePost(c *fiber.Ctx) error {
	curUser := c.Locals("user").(*middleware.SessionUser)

	id, _ := c.ParamsInt("id")

	form := new(models.PostModel)
	if id > 0 {
		post, err := a.PostService.GetPost(c.Context(), int32(id), curUser.Id)
		if err != nil {
			return fiber.NewError(404, "Post not found")
		}

		postId := int64(post.ID)
		publishedOn := post.PublishedOn.Format(time.DateOnly)

		form.Content = template.HTMLEscapeString(post.Body)
		form.Id = &postId
		form.Title = post.Title
		form.PublishedOn = &publishedOn
	}

	labels, err := a.LabelService.GetLabels(c.Context(), curUser.Project.ID)
	if err != nil {
		return fiber.NewError(500, "Error when fetching labels")
	}

	return c.Render("post", fiber.Map{"form": form, "Id": id, "Labels": labels})
}

func (a *AppHandler) LoadPosts(c *fiber.Ctx) error {
	curUser := c.Locals("user").(*middleware.SessionUser)

	posts, err := a.PostService.GetPosts(c.Context(), curUser.Id)

	if err != nil {
		return fiber.NewError(503, "Could not get posts for user")
	}

	return c.Render("partials/components/post_table", fiber.Map{"Posts": posts, "Empty": len(posts) == 0})
}

func (a *AppHandler) DeletePost(c *fiber.Ctx) error {
	curUser := c.Locals("user").(*middleware.SessionUser)

	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Render("partials/components/banner", fiber.Map{"Message": "Invalid id parameter supplied"})
	}

	if _, err := a.PostService.DeletePost(c.Context(), int32(id), curUser.Id); err != nil {
		return c.Render("partials/components/banner", fiber.Map{"Message": "An error occured when deleting the post"})
	}

	return c.Render("partials/components/banner", fiber.Map{"Message": "Post successfully deleted", "Success": true})
}

func (a *AppHandler) ConfirmDeletePost(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(400, "Invalid id parameter supplied")
	}

	return c.Render("partials/components/delete_confirm_modal", fiber.Map{"Title": "Confirm deletion",
		"Body":        "Are you sure you want to delete this post",
		"EndpointUri": "/admin/posts/delete/" + fmt.Sprint(id),
		"ElementType": "table",
		"ElementId":   "post-" + fmt.Sprint(id),
	})
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
		return c.Render("partials/components/post_form", fiber.Map{"project": project, "firstPost": true})
	}

	return c.Render("partials/components/dashboard_widget", fiber.Map{"Project": project})
}

func (a *AppHandler) PostOnboarding(c *fiber.Ctx) error {
	curUser := c.Locals("user").(*middleware.SessionUser)

	form := new(models.ProjectModel)
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

	return c.Render("partials/components/post_form", fiber.Map{"project": savedProject, "firstPost": true})
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
		form.Content = template.HTMLEscapeString(form.Content)
		return c.Render("partials/components/post_form", fiber.Map{"form": form, "errors": errs, "firstPost": form.First})
	}

	var fileName *string
	if file, err := c.FormFile("banner"); err == nil {
		uploadedFile, err := a.CDNService.UploadImage(file, 250)
		if err != nil {
			return c.Render("partials/components/post_form", fiber.Map{"error": err.Error(), "form": form, "firstPost": form.First})
		}

		fileName = uploadedFile
	}

	project, err := a.ProjectService.GetProjectForUser(c.Context(), curUser.Id)
	if err != nil {
		return c.Render("partials/components/post_form", fiber.Map{"error": "Could not locate project for user", "form": form, "firstPost": form.First})
	}

	if form.Id != nil && *form.Id > 0 {
		_, err := a.PostService.UpdatePost(c.Context(), *form, fileName, curUser.Id, project.ID)
		if err != nil {
			return c.Render("partials/components/post_form", fiber.Map{"error": err.Error(), "form": form, "firstForm": form.First})
		}
	} else {
		_, err := a.PostService.InsertPost(c.Context(), *form, fileName, curUser.Id, project.ID)
		if err != nil {
			return c.Render("partials/components/post_form", fiber.Map{"error": err.Error(), "form": form, "firstForm": form.First})
		}
	}

	c.Response().Header.Add("HX-Redirect", "/admin/posts")
	return c.Render("partials/components/dashboard_widget", fiber.Map{})
}
