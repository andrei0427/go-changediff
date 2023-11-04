package handlers

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/url"
	"strings"
	"time"

	"github.com/andrei0427/go-changediff/internal/app"
	"github.com/andrei0427/go-changediff/internal/app/middleware"
	"github.com/andrei0427/go-changediff/internal/app/models"
	"github.com/andrei0427/go-changediff/internal/app/services"
	"github.com/andrei0427/go-changediff/internal/data"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/exp/slices"
)

type AppHandler struct {
	AuthorService  *services.AuthorService
	ProjectService *services.ProjectService
	PostService    *services.PostService
	CDNService     *services.CDNService
	LabelService   *services.LabelService
	CacheService   *services.CacheService
	RoadmapService *services.RoadmapService
}

func NewAppHandler(
	authorService *services.AuthorService,
	projectService *services.ProjectService,
	postService *services.PostService,
	cdnService *services.CDNService,
	labelService *services.LabelService,
	cacheService *services.CacheService,
	roadmapService *services.RoadmapService,
) *AppHandler {
	return &AppHandler{
		AuthorService:  authorService,
		ProjectService: projectService,
		PostService:    postService,
		CDNService:     cdnService,
		LabelService:   labelService,
		CacheService:   cacheService,
		RoadmapService: roadmapService,
	}
}

func InitRoutes(app *app.App) {
	appHandler := NewAppHandler(app.AuthorService,
		app.ProjectService,
		app.PostService,
		app.CDNService,
		app.LabelService,
		app.CacheService,
		app.RoadmapService)
	// app.Fiber.Get("/", appHandler.Home)

	widget := app.Fiber.Group("/widget", middleware.UseLocale, middleware.UseUserId, middleware.UseUserInfo)
	widget.Get("/:key", appHandler.WidgetHome)

	widget.Use(middleware.UseWithUserId)
	changelog := widget.Group("/changelog")
	changelog.Get("/:key", appHandler.WidgetChangelog)
	changelog.Get("/posts/:key/:pageNo?", appHandler.WidgetChangelogPosts)
	changelog.Post("/posts/:key/:pageNo?", appHandler.WidgetChangelogPosts)
	changelog.Put("/posts/view/:key/:postId/:reaction?", appHandler.WidgetChangelogReaction)
	changelog.Put("/posts/comment/:key/:postId", appHandler.WidgetChangelogComment)

	widget.Get("/roadmap/:key", appHandler.WidgetRoadmap)
	widget.Get("/ideas/:key", appHandler.WidgetFeedback)

	admin := app.Fiber.Group("/admin", middleware.UseAuth)
	admin.Get("/dashboard", appHandler.Dashboard)
	admin.Use(func(c *fiber.Ctx) error {
		return middleware.UseProject(c, appHandler.CacheService, appHandler.ProjectService)
	})
	admin.Get("/project", appHandler.GetProject)
	admin.Post("/onboarding", appHandler.PostOnboarding)

	admin.Use(func(c *fiber.Ctx) error {
		return middleware.UseAuthor(c, appHandler.CacheService, appHandler.AuthorService)
	})

	analytics := admin.Group("/analytics")
	analytics.Get("/", appHandler.GetUserAnalytics)
	analytics.Get("/user", appHandler.GetAnalyticsByUser)

	billing := admin.Group("/billing")
	billing.Get("/", appHandler.GetBilling)

	posts := admin.Group("/posts")
	posts.Get("/", appHandler.Posts)
	posts.Get("/compose/:id?", appHandler.ComposePost)
	posts.Post("/save", appHandler.SavePost)
	posts.Get("/load", appHandler.LoadPosts)
	posts.Get("/load-reactions/:postId", appHandler.LoadPostReactions)
	posts.Delete("/delete/:id", appHandler.DeletePost)
	posts.Delete("/confirm-delete/:id", appHandler.ConfirmDeletePost)

	settings := admin.Group("/settings")
	settings.Get("/", appHandler.Settings)
	settings.Get("/tab", appHandler.SettingsTab)
	settings.Post("/general/save", appHandler.SaveSettingsGeneralTab)

	settings.Post("/changelog/labels/save", appHandler.SaveSettingsLabelsTab)
	settings.Delete("/changelog/labels/delete/:id", appHandler.DeleteSettingsLabel)
	settings.Get("/changelog/labels/confirm-delete/:id", appHandler.ConfirmDeleteLabel)
	settings.Get("/changelog/labels/new", appHandler.NewSettingsLabel)

	settings.Get("/roadmap/boards/open/:id?", appHandler.RoadmapBoardOpen)
	settings.Post("/roadmap/boards/save", appHandler.RoadmapBoardSave)
	settings.Get("/roadmap/boards/confirm-delete/:id", appHandler.ConfirmDeleteBoard)
	settings.Delete("/roadmap/boards/delete/:id", appHandler.DeleteSettingsBoard)

	settings.Get("/roadmap/status/open/:id?", appHandler.RoadmapStatusOpen)
	settings.Post("/roadmap/status/save", appHandler.RoadmapStatusSave)
	settings.Get("/roadmap/status/confirm-delete/:id", appHandler.ConfirmDeleteStatus)
	settings.Delete("/roadmap/status/delete/:id", appHandler.DeleteSettingsStatus)
}

// Public Routes

// func (a *AppHandler) Home(c *fiber.Ctx) error {
// 	return c.Render("index", fiber.Map{})
// }

func (a *AppHandler) WidgetHome(c *fiber.Ctx) error {
	key := c.Params("key")
	isEmbedded := c.Query("embed") == "1"
	userUuid := c.Locals("userUuid")

	project, err := a.ProjectService.GetProjectByKey(c.Context(), a.CacheService, key)
	if err != nil {
		return fiber.NewError(404, "Project not found")
	}

	logoUrl := ""
	if project.LogoUrl.Valid {
		logoUrl = project.LogoUrl.String
	}

	newUserId := ""
	if userUuid == nil {
		newUserId = uuid.New().String()
	}

	return c.Render("widget/index", fiber.Map{"Project": project, "LogoUrl": logoUrl, "activeTab": "changelog", "newUserId": newUserId, "isEmbedded": isEmbedded})
}

func (a *AppHandler) WidgetChangelog(c *fiber.Ctx) error {
	key := c.Params("key")
	project, err := a.ProjectService.GetProjectByKey(c.Context(), a.CacheService, key)
	if err != nil {
		return fiber.NewError(404, "Project not found")
	}

	return c.Render("widget/tabs/changelog", fiber.Map{"Project": project})
}

func (a *AppHandler) WidgetChangelogPosts(c *fiber.Ctx) error {
	key := c.Params("key")
	paramPageNo, _ := c.ParamsInt("pageNo")

	model := new(models.Search)
	c.BodyParser(model)

	userUuid := c.Locals("userUuid").(*uuid.UUID)
	pageNo := 1
	if paramPageNo > 0 {
		pageNo = paramPageNo
	}

	posts, err := a.PostService.GetPublishedPagedPosts(c.Context(), key, int32(pageNo), model.Search, *userUuid)
	if err != nil {
		return fiber.NewError(503, "Error fetching posts")
	}

	project, err := a.ProjectService.GetProjectByKey(c.Context(), a.CacheService, key)
	if err != nil {
		return fiber.NewError(503, "Error fetching project")
	}

	return c.Render("widget/components/posts", fiber.Map{"Posts": posts, "ProjectKey": key, "Project": project})
}

func (a *AppHandler) WidgetChangelogReaction(c *fiber.Ctx) error {
	userUuid := c.Locals("userUuid").(*uuid.UUID)
	userLocale := c.Locals("timezone").(string)
	userInfo := c.Locals("userInfo").(*models.UserInfo)

	projectKey := c.Params("key")
	reaction := c.Params("reaction")
	postId, err := c.ParamsInt("postId")

	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Post ID is required")
	}

	if projectKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Project key is required")
	}

	project, err := a.ProjectService.GetProjectByKey(c.Context(), a.CacheService, projectKey)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Project not found")
	}

	post, err := a.PostService.GetPost(c.Context(), int32(postId), project.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Post not found")
	}

	var Reaction = sql.NullString{String: "", Valid: false}
	if reaction != "" {
		parsed, err := url.QueryUnescape(reaction)

		if err == nil {
			Reaction = sql.NullString{String: parsed, Valid: true}
		}
	}

	_, err = a.PostService.SaveReaction(c.Context(), data.InsertReactionParams{
		UserUuid:  *userUuid,
		PostID:    post.ID,
		IpAddr:    c.IP(),
		UserAgent: c.Get("User-Agent"),
		Locale:    userLocale,
		Reaction:  Reaction,
	}, userInfo)

	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Something went wrong")
	}

	return c.SendStatus(fiber.StatusOK)
}

func (a *AppHandler) WidgetChangelogComment(c *fiber.Ctx) error {
	userUuid := c.Locals("userUuid").(*uuid.UUID)
	projectKey := c.Params("key")
	postId, err := c.ParamsInt("postId")
	body := new(models.ChangelogComment)

	if err := c.BodyParser(body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Comment is required")
	}

	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Post ID is required")
	}

	if projectKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Project key is required")
	}

	project, err := a.ProjectService.GetProjectByKey(c.Context(), a.CacheService, projectKey)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Project not found")
	}

	_, err = a.PostService.GetPost(c.Context(), int32(postId), project.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Post not found")
	}

	_, err = a.PostService.InsertPostComment(c.Context(), *userUuid, body.Comment, int32(postId))
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Something went wrong")
	}

	return c.SendStatus(fiber.StatusOK)
}

func (a *AppHandler) WidgetRoadmap(c *fiber.Ctx) error {
	key := c.Params("key")
	project, err := a.ProjectService.GetProjectByKey(c.Context(), a.CacheService, key)
	if err != nil {
		return fiber.NewError(404, "Project not found")
	}

	return c.Render("widget/tabs/roadmap", fiber.Map{"Project": project})
}

func (a *AppHandler) WidgetFeedback(c *fiber.Ctx) error {
	key := c.Params("key")
	project, err := a.ProjectService.GetProjectByKey(c.Context(), a.CacheService, key)
	if err != nil {
		return fiber.NewError(404, "Project not found")
	}

	return c.Render("widget/tabs/ideas", fiber.Map{"Project": project})
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
	curUser := c.Locals("user").(*models.SessionUser)
	tabName := c.Query("hash", "general")

	switch tabName {
	case "general":
		return c.Render("partials/components/settings/general_tab", fiber.Map{"Form": curUser.Project})

	case "changelog":
		labels, err := a.LabelService.GetLabels(c.Context(), curUser.Project.ID)

		var message string
		if err != nil {
			message = err.Error()
		}

		return c.Render("partials/components/settings/changelog_tab", fiber.Map{"Labels": labels, "Message": message})

	case "roadmap":
		boards, boardsErr := a.RoadmapService.GetBoards(c.Context(), curUser.Project.ID)
		statuses, statusesErr := a.RoadmapService.GetStatuses(c.Context(), curUser.Project.ID)

		var message = ""
		if boardsErr != nil {
			message = boardsErr.Error()
		}

		if statusesErr != nil {
			message = message + "\n" + statusesErr.Error()
		}

		return c.Render("partials/components/settings/roadmap_tab", fiber.Map{"Boards": boards, "BoardsEmpty": len(boards) == 0, "Statuses": statuses, "StatusesEmpty": len(statuses) == 0, "Message": message})

	default:
		return c.Render("partials/components/settings/general_tab", fiber.Map{"Form": curUser.Project})
	}
}

func (a *AppHandler) RoadmapStatusSave(c *fiber.Ctx) error {
	viewPath := "partials/components/settings/roadmap_status_slideover_form"
	curUser := c.Locals("user").(*models.SessionUser)

	form := new(models.RoadmapStatusModel)
	if err := c.BodyParser(form); err != nil {
		return c.Render(viewPath, fiber.Map{"Error": err.Error()})
	}

	errs := make(map[string]string)
	if len(strings.TrimSpace(form.Status)) == 0 {
		errs["Status"] = "Status is required"
	}

	if len(strings.TrimSpace(form.Color)) == 0 {
		errs["Name"] = "Color is required"
	}

	if len(errs) > 0 {
		return c.Render(viewPath, fiber.Map{"Status": form, "Errors": errs})
	}

	savedStatus, err := a.RoadmapService.SaveStatus(c.Context(), *form, curUser.Project.ID)
	if err != nil {
		return c.Render(viewPath, fiber.Map{"Status": form, "Error": err.Error()})
	}

	return c.Render(viewPath, fiber.Map{"Status": savedStatus, "Success": true, "Message": "Status saved successfully.", "Close": true})
}

func (a *AppHandler) ConfirmDeleteStatus(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(400, "Invalid id parameter supplied")
	}

	return c.Render("partials/components/delete_confirm_modal", fiber.Map{"Title": "Confirm deletion",
		"Body":           "Are you sure you want to delete this status? All posts within must also be deleted for this action to happen.",
		"EndpointUri":    "/admin/settings/roadmap/status/delete/" + fmt.Sprint(id),
		"TargetSelector": "#status-content",
		"Swap":           "innerHtml",
	})
}

func (a *AppHandler) DeleteSettingsStatus(c *fiber.Ctx) error {
	viewPath := "partials/components/settings/roadmap_tab_status"
	curUser := c.Locals("user").(*models.SessionUser)

	currentStatuses, err := a.RoadmapService.GetStatuses(c.Context(), curUser.Project.ID)
	if err != nil {
		return c.Render(viewPath, fiber.Map{"Error": err.Error()})
	}

	statusIdToDelete, err := c.ParamsInt("id")
	if err != nil {
		return c.Render(viewPath, fiber.Map{"Statuses": currentStatuses, "Error": err.Error()})
	}

	err = a.RoadmapService.DeleteStatus(c.Context(), int32(statusIdToDelete), curUser.Project.ID)
	if err != nil {
		return c.Render(viewPath, fiber.Map{"Statuses": currentStatuses, "Error": err.Error()})
	}

	statusIdxToDelete := slices.IndexFunc(currentStatuses, func(l data.GetStatusesRow) bool {
		return l.ID == int32(statusIdToDelete)
	})

	updatedStatuses := make([]data.GetStatusesRow, 0)
	updatedStatuses = append(updatedStatuses, currentStatuses[:statusIdxToDelete]...)
	updatedStatuses = append(updatedStatuses, currentStatuses[statusIdxToDelete+1:]...)

	return c.Render(viewPath, fiber.Map{"Statuses": updatedStatuses, "Success": true, "Message": "Status successfully deleted"})
}

func (a *AppHandler) RoadmapStatusOpen(c *fiber.Ctx) error {
	curUser := c.Locals("user").(*models.SessionUser)
	id, err := c.ParamsInt("id", -1)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid id parameter supplied")
	}

	var status data.GetStatusRow

	if id > 0 {
		status, err = a.RoadmapService.GetStatus(c.Context(), int32(id), curUser.Project.ID)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Error fetching board")
		}
	}

	return c.Render("partials/components/settings/roadmap_status_slideover", fiber.Map{"Status": status})
}

func (a *AppHandler) RoadmapBoardSave(c *fiber.Ctx) error {
	viewPath := "partials/components/settings/roadmap_board_slideover_form"
	curUser := c.Locals("user").(*models.SessionUser)

	form := new(models.RoadmapBoardModel)
	if err := c.BodyParser(form); err != nil {
		return c.Render(viewPath, fiber.Map{"Error": err.Error()})
	}

	errs := make(map[string]string)
	if len(strings.TrimSpace(form.Name)) == 0 {
		errs["Name"] = "Name is required"
	}

	if len(errs) > 0 {
		return c.Render(viewPath, fiber.Map{"Board": form, "Errors": errs})
	}

	savedBoard, err := a.RoadmapService.SaveBoard(c.Context(), *form, curUser.Project.ID)
	if err != nil {
		return c.Render(viewPath, fiber.Map{"Board": form, "Error": err.Error()})
	}

	return c.Render(viewPath, fiber.Map{"Board": savedBoard, "Success": true, "Message": "Board saved successfully.", "Close": true})
}

func (a *AppHandler) ConfirmDeleteBoard(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(400, "Invalid id parameter supplied")
	}

	return c.Render("partials/components/delete_confirm_modal", fiber.Map{"Title": "Confirm deletion",
		"Body":           "Are you sure you want to delete this board? All posts within must also be deleted for this action to happen.",
		"EndpointUri":    "/admin/settings/roadmap/boards/delete/" + fmt.Sprint(id),
		"TargetSelector": "#board-content",
		"Swap":           "innerHtml",
	})
}

func (a *AppHandler) DeleteSettingsBoard(c *fiber.Ctx) error {
	viewPath := "partials/components/settings/roadmap_tab_boards"
	curUser := c.Locals("user").(*models.SessionUser)

	currentBoards, err := a.RoadmapService.GetBoards(c.Context(), curUser.Project.ID)
	if err != nil {
		return c.Render(viewPath, fiber.Map{"Error": err.Error()})
	}

	boardIdToDelete, err := c.ParamsInt("id")
	if err != nil {
		return c.Render(viewPath, fiber.Map{"Boards": currentBoards, "Error": err.Error()})
	}

	err = a.RoadmapService.DeleteBoard(c.Context(), int32(boardIdToDelete), curUser.Project.ID)
	if err != nil {
		return c.Render(viewPath, fiber.Map{"Boards": currentBoards, "Error": err.Error()})
	}

	boardIdxToDelete := slices.IndexFunc(currentBoards, func(l data.GetBoardsRow) bool {
		return l.ID == int32(boardIdToDelete)
	})

	updatedBoards := make([]data.GetBoardsRow, 0)
	updatedBoards = append(updatedBoards, currentBoards[:boardIdxToDelete]...)
	updatedBoards = append(updatedBoards, currentBoards[boardIdxToDelete+1:]...)

	return c.Render(viewPath, fiber.Map{"Boards": updatedBoards, "Success": true, "Message": "Board successfully deleted"})
}

func (a *AppHandler) RoadmapBoardOpen(c *fiber.Ctx) error {
	curUser := c.Locals("user").(*models.SessionUser)
	id, err := c.ParamsInt("id", -1)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid id parameter supplied")
	}

	var board data.GetBoardRow
	if id > 0 {
		board, err = a.RoadmapService.GetBoard(c.Context(), int32(id), curUser.Project.ID)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Error fetching board")
		}
	}

	return c.Render("partials/components/settings/roadmap_board_slideover", fiber.Map{"Board": board})

}

func (a *AppHandler) NewSettingsLabel(c *fiber.Ctx) error {
	curUser := c.Locals("user").(*models.SessionUser)
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
	viewPath := "partials/components/settings/changelog_tab"
	curUser := c.Locals("user").(*models.SessionUser)
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
	viewPath := "partials/components/settings/changelog_tab"
	curUser := c.Locals("user").(*models.SessionUser)
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
	curUser := c.Locals("user").(*models.SessionUser)

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
	curUser := c.Locals("user").(*models.SessionUser)

	id, _ := c.ParamsInt("id")

	form := new(models.PostModel)
	if id > 0 {
		post, err := a.PostService.GetPost(c.Context(), int32(id), curUser.Project.ID)
		if err != nil {
			return fiber.NewError(404, "Post not found")
		}

		postId := post.ID
		publishedOn := post.PublishedOn.Format(time.DateTime)

		form.Content = template.HTMLEscapeString(post.Body)
		form.ID = &postId
		form.Title = post.Title
		form.PublishedOn = publishedOn
		form.IsPublished = post.IsPublished.Bool

		if post.ExpiresOn.Valid {
			expiresOn := post.ExpiresOn.Time.Format(time.DateTime)
			form.ExpiresOn = expiresOn
		}

		if post.LabelID.Valid {
			labelId := int(post.LabelID.Int32)
			form.LabelId = &labelId
		}
	} else {
		form.PublishedOn = time.Now().UTC().Format(time.DateTime)
	}

	labels, err := a.LabelService.GetLabels(c.Context(), curUser.Project.ID)
	if err != nil {
		return fiber.NewError(500, "Error when fetching labels")
	}

	return c.Render("post", fiber.Map{"form": form, "Id": id, "Labels": labels})
}

func (a *AppHandler) LoadPostReactions(c *fiber.Ctx) error {
	curUser := c.Locals("user").(*models.SessionUser)
	iPostId, err := c.ParamsInt("postId")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "post id is required")
	}

	postId := int32(iPostId)

	reactions, err := a.PostService.GetPostReactions(c.Context(), &postId, curUser.Project.ID, nil, nil)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "something went wrong")
	}

	comments, err := a.PostService.GetPostComments(c.Context(), &postId, curUser.Project.ID, nil, nil)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "something went wrong")
	}

	return c.Render("partials/components/post_reactions_slideover", fiber.Map{"Title": "Post Reactions", "Reactions": reactions, "Comments": comments, "CommentCount": len(comments)})
}

func (a *AppHandler) LoadPosts(c *fiber.Ctx) error {
	curUser := c.Locals("user").(*models.SessionUser)

	posts, err := a.PostService.GetPosts(c.Context(), curUser.Project.ID)

	if err != nil {
		fmt.Println(err.Error())
		return fiber.NewError(503, "Could not get posts for user")
	}

	return c.Render("partials/components/post_table", fiber.Map{"Posts": posts, "Empty": len(posts) == 0})
}

func (a *AppHandler) DeletePost(c *fiber.Ctx) error {
	curUser := c.Locals("user").(*models.SessionUser)

	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Render("partials/components/banner", fiber.Map{"Message": "Invalid id parameter supplied"})
	}

	if _, err := a.PostService.DeletePost(c.Context(), int32(id), curUser.Project.ID); err != nil {
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
	curUser := c.Locals("user").(*models.SessionUser)

	if curUser.Project == nil {
		return c.Render("get_started", fiber.Map{})
	}

	posts, err := a.PostService.GetPostCountForProject(c.Context(), curUser.Project.ID)
	if err != nil {
		return fiber.NewError(503, "Could not get posts for user")
	}

	if posts == 0 {
		labels, err := a.LabelService.GetLabels(c.Context(), curUser.Project.ID)
		if err != nil {
			return fiber.NewError(500, "Error when fetching labels")
		}

		return c.Render("partials/components/post_form", fiber.Map{"project": curUser.Project, "firstPost": true, "Labels": labels})
	}

	return c.Render("partials/components/dashboard_widget", fiber.Map{"Project": curUser.Project, "PostCount": posts})
}

func (a *AppHandler) PostOnboarding(c *fiber.Ctx) error {
	curUser := c.Locals("user").(*models.SessionUser)

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

	curUser.Project = &savedProject
	a.AuthorService.InsertAuthorForUser(c.Context(), *curUser)

	labels, err := a.LabelService.GetLabels(c.Context(), curUser.Project.ID)
	if err != nil {
		return fiber.NewError(500, "Error when fetching labels")
	}

	return c.Render("partials/components/post_form", fiber.Map{"project": savedProject, "firstPost": true, "Labels": labels})
}

func (a *AppHandler) SavePost(c *fiber.Ctx) error {
	curUser := c.Locals("user").(*models.SessionUser)

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

	if len(strings.TrimSpace(form.PublishedOn)) == 0 {
		errs["PublishedOn"] = "Publish date is required"
	}

	labels, err := a.LabelService.GetLabels(c.Context(), curUser.Project.ID)
	if err != nil {
		return fiber.NewError(500, "Error when fetching labels")
	}

	if len(errs) > 0 {
		form.Content = template.HTMLEscapeString(form.Content)
		return c.Render("partials/components/post_form", fiber.Map{"form": form, "errors": errs, "firstPost": form.First, "Labels": labels})
	}

	loc, err := time.LoadLocation(curUser.Timezone)
	if err != nil {
		return c.Render("partials/components/post_form", fiber.Map{"error": "Could not locate user timezone", "form": form, "firstPost": form.First, "Labels": labels})
	}

	if form.ID != nil && *form.ID > 0 {
		_, err := a.PostService.UpdatePost(c.Context(), *form, curUser.Project.ID, loc)
		if err != nil {
			return c.Render("partials/components/post_form", fiber.Map{"error": err.Error(), "form": form, "firstForm": form.First})
		}
	} else {
		_, err := a.PostService.InsertPost(c.Context(), *form, curUser.Author.ID, curUser.Project.ID, loc)
		if err != nil {
			return c.Render("partials/components/post_form", fiber.Map{"error": err.Error(), "form": form, "firstForm": form.First})
		}
	}

	c.Response().Header.Add("HX-Redirect", "/admin/posts")
	return c.Render("partials/components/dashboard_widget", fiber.Map{})
}

func (a *AppHandler) GetBilling(c *fiber.Ctx) error {
	curUser := c.Locals("user").(*models.SessionUser)

	proPrice := 19
	proAnnualPrice := ((proPrice * 10) / 12)

	maxPrice := 39
	maxAnnualPrice := ((maxPrice * 10) / 12)

	expirationDate := ""
	if curUser.Author.Tier.Valid && curUser.Author.Tier.Int32 >= 1 && curUser.Author.ExpiresOn != nil {
		if t, ok := curUser.Author.ExpiresOn.(time.Time); ok {
			expirationDate = fmt.Sprintf("Expires on %s, %d %s %d.", t.Weekday(), t.Day(), t.Month(), t.Year())
		}
	}

	return c.Render("billing", fiber.Map{"user": curUser, "ProPrice": proPrice, "ProAnnualPrice": proAnnualPrice, "MaxPrice": maxPrice, "MaxAnnualPrice": maxAnnualPrice, "ExpirationDate": expirationDate})
}

func (a *AppHandler) GetUserAnalytics(c *fiber.Ctx) error {
	curUser := c.Locals("user").(*models.SessionUser)

	data, err := a.PostService.GetAnalytics(c.Context(), curUser.Project.ID, nil, nil)
	if err != nil {
		fmt.Println(err)
		return fiber.NewError(fiber.StatusInternalServerError, "error when fetching analytics")
	}

	return c.Render("analytics", fiber.Map{"data": data, "AnalyticsEmpty": len(data) == 0})
}

func (a *AppHandler) GetAnalyticsByUser(c *fiber.Ctx) error {
	curUser := c.Locals("user").(*models.SessionUser)
	qUserUuid := c.Query("uuid")
	qUserId := c.Query("userId")

	userUuid := string(qUserUuid)
	userId := string(qUserId)

	data, err := a.PostService.GetAnalytics(c.Context(), curUser.Project.ID, &userUuid, &userId)
	if err != nil || len(data) == 0 {
		return fiber.NewError(fiber.StatusInternalServerError, "error when fetching analytics")
	}

	comments, err := a.PostService.GetPostComments(c.Context(), nil, curUser.Project.ID, &userUuid, &userId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "error when fetching analytics comments")
	}

	reactions, err := a.PostService.GetPostReactions(c.Context(), nil, curUser.Project.ID, &userUuid, &userId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "error when fetching analytics reactons")
	}

	return c.Render("partials/components/analytics/user_reactions_slideover", fiber.Map{"Data": data[0], "Comments": comments, "CommentCount": len(comments), "Reactions": reactions})
}
