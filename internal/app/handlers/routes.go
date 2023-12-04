package handlers

import (
	"fmt"
	"html/template"
	"net/url"
	"sort"
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
	ViewerService  *services.ViewerService
	ProjectService *services.ProjectService
	PostService    *services.PostService
	CDNService     *services.CDNService
	LabelService   *services.LabelService
	CacheService   *services.CacheService
	RoadmapService *services.RoadmapService
}

func NewAppHandler(
	authorService *services.AuthorService,
	viewerService *services.ViewerService,
	projectService *services.ProjectService,
	postService *services.PostService,
	cdnService *services.CDNService,
	labelService *services.LabelService,
	cacheService *services.CacheService,
	roadmapService *services.RoadmapService,
) *AppHandler {
	return &AppHandler{
		AuthorService:  authorService,
		ViewerService:  viewerService,
		ProjectService: projectService,
		PostService:    postService,
		CDNService:     cdnService,
		LabelService:   labelService,
		CacheService:   cacheService,
		RoadmapService: roadmapService,
	}
}

func InitRoutes(app *app.App) {
	appHandler := NewAppHandler(
		app.AuthorService,
		app.ViewerService,
		app.ProjectService,
		app.PostService,
		app.CDNService,
		app.LabelService,
		app.CacheService,
		app.RoadmapService)
	// app.Fiber.Get("/", appHandler.Home)

	widget := app.Fiber.Group("/widget/:key", middleware.UseLocale, middleware.UseUserId, middleware.UseUserInfo)
	widget.Get("/", appHandler.WidgetHome)

	widget.Use(func(c *fiber.Ctx) error {
		return middleware.UseWithViewer(c, appHandler.CacheService, appHandler.ProjectService, appHandler.ViewerService)
	})

	changelog := widget.Group("/changelog")
	changelog.Get("/", appHandler.WidgetChangelog)
	changelog.Get("/posts/:pageNo?", appHandler.WidgetChangelogPosts)
	changelog.Put("/posts/view/:postId/:reaction?", appHandler.WidgetChangelogReaction)
	changelog.Put("/posts/comment/:postId", appHandler.WidgetChangelogComment)

	widget.Get("/roadmap", appHandler.WidgetRoadmap)
	widget.Get("/ideas", appHandler.WidgetFeedback)

	admin := app.Fiber.Group("/admin", middleware.UseAuth)
	admin.Get("/dashboard", appHandler.Dashboard)
	admin.Use(func(c *fiber.Ctx) error {
		return middleware.UseProject(c, appHandler.CacheService, appHandler.ProjectService, false)
	})
	admin.Get("/project", appHandler.GetProject)
	admin.Post("/onboarding", appHandler.PostOnboarding)

	admin.Use(func(c *fiber.Ctx) error {
		return middleware.UseAuthor(c, appHandler.CacheService, appHandler.AuthorService)
	})

	analytics := admin.Group("/analytics")
	analytics.Get("/", appHandler.GetUserAnalytics)
	analytics.Get("/user/:viewerId", appHandler.GetAnalyticsByUser)

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

	roadmap := admin.Group("/roadmap")
	roadmap.Get("/", appHandler.GetRoadmap)
	roadmap.Get("/board", appHandler.GetBoard)
	roadmapPost := roadmap.Group("/post")
	roadmapPost.Get("/compose", appHandler.GetRoadmapComposeForm)
	roadmapPost.Get("/activity/:postId/:commentId?", appHandler.GetRoadmapPostActivity)
	roadmapPost.Post("/activity/:postId/comment/:commentId?", appHandler.PostRoadmapPostComment)
	roadmapPost.Post("/save", appHandler.SaveRoadmapPost)
	roadmapPost.Delete("/delete/:id", appHandler.DeleteRoadmapPost)
	roadmapPost.Delete("/confirm-delete/:id", appHandler.ConfirmDeleteRoadmapPost)
	roadmapPost.Post("/save-status/:boardId/:id/:statusId", appHandler.SaveRoadmapPostStatus)

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
	settings.Post("/roadmap/status/order/up/:id", func(c *fiber.Ctx) error {
		return appHandler.ChangeStatusOrder(c, true)
	})
	settings.Post("/roadmap/status/order/down/:id", func(c *fiber.Ctx) error {
		return appHandler.ChangeStatusOrder(c, false)
	})
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

	viewer := c.Locals("viewer").(data.Viewer)
	pageNo := 1
	if paramPageNo > 0 {
		pageNo = paramPageNo
	}

	project, err := a.ProjectService.GetProjectByKey(c.Context(), a.CacheService, key)
	if err != nil {
		return fiber.NewError(503, "Error fetching project")
	}

	posts, err := a.PostService.GetPublishedPagedPosts(c.Context(), key, int32(pageNo), model.Search, viewer.ID)
	if err != nil {
		return fiber.NewError(503, "Error fetching posts")
	}

	return c.Render("widget/components/posts", fiber.Map{"Posts": posts, "ProjectKey": key, "Project": project})
}

func (a *AppHandler) WidgetChangelogReaction(c *fiber.Ctx) error {
	viewer := c.Locals("viewer").(data.Viewer)
	reaction := c.Params("reaction")
	postId, err := c.ParamsInt("postId")

	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Post ID is required")
	}

	post, err := a.PostService.GetPost(c.Context(), int32(postId), viewer.ProjectID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Post not found")
	}

	var Reaction *string
	if reaction != "" {
		parsed, err := url.QueryUnescape(reaction)

		if err == nil {
			Reaction = &parsed
		}
	}

	interactionType := models.InteractionTypeView
	if Reaction != nil && len(*Reaction) > 0 {
		interactionType = models.InteractionTypeReaction
	}

	_, err = a.PostService.SaveInteraction(c.Context(), post.ID, viewer.ID, viewer.ProjectID, interactionType, Reaction)

	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Something went wrong")
	}

	return c.SendStatus(fiber.StatusOK)
}

func (a *AppHandler) WidgetChangelogComment(c *fiber.Ctx) error {
	viewer := c.Locals("viewer").(data.Viewer)
	postId, err := c.ParamsInt("postId")
	body := new(models.ChangelogComment)

	if err := c.BodyParser(body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Comment is required")
	}

	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Post ID is required")
	}

	post, err := a.PostService.GetPost(c.Context(), int32(postId), viewer.ProjectID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Post not found")
	}

	_, err = a.PostService.InsertPostComment(c.Context(), viewer.ID, post.ID, viewer.ProjectID, body.Comment)
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
	middleware.UseProject(c, a.CacheService, a.ProjectService, true)

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
		form := models.GeneralSettingsModel{
			Name:        curUser.Project.Name,
			Description: curUser.Project.Description,
			AccentColor: curUser.Project.AccentColor,
			FirstName:   curUser.Author.FirstName,
			LastName:    curUser.Author.LastName,
		}

		return c.Render("partials/components/settings/general_tab", fiber.Map{"Form": form, "LogoUrl": curUser.Project.LogoUrl})

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

func (a *AppHandler) ChangeStatusOrder(c *fiber.Ctx, up bool) error {
	viewPath := "partials/components/settings/roadmap_tab_status"
	curUser := c.Locals("user").(*models.SessionUser)

	currentStatuses, err := a.RoadmapService.GetStatuses(c.Context(), curUser.Project.ID)
	if err != nil {
		return c.Render(viewPath, fiber.Map{"Error": err.Error()})
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Render(viewPath, fiber.Map{"Statuses": currentStatuses, "Error": err.Error()})
	}

	updatedStatuses, err := a.RoadmapService.UpdateStatusSortOrder(c.Context(), up, int32(id), curUser.Project.ID)
	if err != nil {
		return c.Render(viewPath, fiber.Map{"Statuses": currentStatuses, "Error": err.Error()})
	}

	return c.Render(viewPath, fiber.Map{"Statuses": updatedStatuses, "Success": true, "Message": "Status order successfully updated"})

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

	form := new(models.GeneralSettingsModel)
	if err := c.BodyParser(form); err != nil {
		return c.Render(viewPath, fiber.Map{"Error": err.Error()})
	}

	errs := make(map[string]string)
	if len(strings.TrimSpace(form.Name)) == 0 {
		errs["Name"] = "Name is required"
	}

	if len(strings.TrimSpace(form.FirstName)) == 0 {
		errs["FirstName"] = "First name is required"
	}

	if len(strings.TrimSpace(form.LastName)) == 0 {
		errs["LastName"] = "Last name is required"
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

	var profilePictureUrl *string
	if file, err := c.FormFile("display_picture"); err == nil {
		uploadedFile, err := a.CDNService.UploadImage(file, 250)
		if err != nil {
			return c.Render(viewPath, fiber.Map{"Error": err.Error(), "Form": form})
		}

		profilePictureUrl = uploadedFile
	} else if curUser.Author.PictureUrl.Valid {
		profilePictureUrl = &curUser.Author.PictureUrl.String
	}

	_, err := a.AuthorService.UpdateAuthorForUser(c.Context(), curUser.Id, curUser.Project.ID, *form, profilePictureUrl)
	if err != nil {
		return c.Render(viewPath, fiber.Map{"Error": err.Error(), "Form": form})
	}
	savedAuthor, err := a.AuthorService.GetAuthorByUser(c.Context(), curUser.Id)
	if err != nil {
		return c.Render(viewPath, fiber.Map{"Error": err.Error(), "Form": form})
	}

	projectModel := models.ProjectModel{
		Name:        form.Name,
		Description: form.Description,
		AccentColor: form.AccentColor,
		ID:          &curUser.Project.ID,
	}
	savedProject, err := a.ProjectService.SaveProject(c.Context(), curUser.Id, projectModel, fileName)
	if err != nil {
		return c.Render(viewPath, fiber.Map{"Error": err.Error(), "Form": form, "LogoUrl": curUser.Project.LogoUrl})
	}

	a.CacheService.Set("user-"+fmt.Sprint(curUser.Id)+"project", &savedProject, nil)
	a.CacheService.Set("user-"+fmt.Sprint(curUser.Id)+"author", savedAuthor, nil)

	return c.Render(viewPath, fiber.Map{"Form": form, "LogoUrl": savedProject.LogoUrl, "SavedProfilePictureUrl": *profilePictureUrl, "Success": true, "Message": "Settings saved successfully!"})
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
			form.ExpiresOn = post.ExpiresOn.Time.Format(time.DateTime)
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

	reactions, err := a.PostService.GetPostReactions(c.Context(), curUser.Project.ID, &postId, nil)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "something went wrong")
	}

	comments, err := a.PostService.GetPostComments(c.Context(), curUser.Project.ID, &postId, nil)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "something went wrong")
	}

	return c.Render("partials/components/post_reactions_slideover", fiber.Map{"Title": "Post Reactions", "Reactions": reactions, "Comments": comments, "CommentCount": len(comments)})
}

func (a *AppHandler) LoadPosts(c *fiber.Ctx) error {
	curUser := c.Locals("user").(*models.SessionUser)

	posts, err := a.PostService.GetPosts(c.Context(), curUser.Project.ID)

	if err != nil {
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
			form.Content = template.HTMLEscapeString(form.Content)
			return c.Render("partials/components/post_form", fiber.Map{"error": err.Error(), "form": form, "firstForm": form.First})
		}
	} else {
		_, err := a.PostService.InsertPost(c.Context(), *form, curUser.Author.ID, curUser.Project.ID, loc)
		if err != nil {
			form.Content = template.HTMLEscapeString(form.Content)
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

func (a *AppHandler) GetRoadmap(c *fiber.Ctx) error {
	curUser := c.Locals("user").(*models.SessionUser)

	boards, err := a.RoadmapService.GetBoards(c.Context(), curUser.Project.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "error when fetching boards")
	}

	var publicBoards, privateBoards []data.GetBoardsRow

	for _, item := range boards {
		if item.IsPrivate {
			privateBoards = append(privateBoards, item)
		} else {
			publicBoards = append(publicBoards, item)
		}
	}

	var firstBoardId int32
	if len(publicBoards) > 0 {
		firstBoardId = publicBoards[0].ID
	} else if len(privateBoards) > 0 {
		firstBoardId = privateBoards[0].ID
	}

	return c.Render("roadmap", fiber.Map{"FirstBoardID": firstBoardId, "PublicBoards": publicBoards, "PrivateBoards": privateBoards, "HasPrivateBoards": len(privateBoards), "BoardCount": len(boards)})
}

func (a *AppHandler) GetBoard(c *fiber.Ctx) error {
	curUser := c.Locals("user").(*models.SessionUser)
	boardId := c.QueryInt("id")

	board, err := a.RoadmapService.GetBoard(c.Context(), int32(boardId), curUser.Project.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "board not found")
	}

	statuses, err := a.RoadmapService.GetStatuses(c.Context(), curUser.Project.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "error fetching statuses")
	}

	posts, err := a.RoadmapService.GetPostsForBoard(c.Context(), board.ID, curUser.Project.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "error fetching posts")
	}

	var unassignedPosts []data.GetPostsForBoardRow
	for _, p := range posts {
		if !p.BoardID.Valid || !p.StatusID.Valid {
			unassignedPosts = append(unassignedPosts, p)
		}
	}

	var statusesWithPosts []models.RoadmapBoardStatusWithPosts
	statusesWithPosts = append(statusesWithPosts, models.RoadmapBoardStatusWithPosts{
		Status: data.GetStatusesRow{ID: -1, Status: "Unassigned", Color: "#DDDDDD"},
		Posts:  unassignedPosts,
	})

	for _, s := range statuses {
		var statusPosts []data.GetPostsForBoardRow

		for _, p := range posts {
			if p.StatusID.Int32 == s.ID {
				statusPosts = append(statusPosts, p)
			}
		}

		statusesWithPosts = append(statusesWithPosts, models.RoadmapBoardStatusWithPosts{Status: s, Posts: statusPosts})
	}

	return c.Render("partials/components/roadmap/board", fiber.Map{"Board": board, "StatusesWithPosts": statusesWithPosts})
}

func (a *AppHandler) GetRoadmapComposeForm(c *fiber.Ctx) error {
	curUser := c.Locals("user").(*models.SessionUser)
	id := c.QueryInt("id", 0)
	statusId := c.QueryInt("statusId", 0)
	boardId := c.QueryInt("boardId", 0)

	statuses, err := a.RoadmapService.GetStatuses(c.Context(), curUser.Project.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "error when fetching statuses")
	}

	boards, err := a.RoadmapService.GetBoards(c.Context(), curUser.Project.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "error when fetching boards")
	}

	form := models.RoadmapPostModel{StatusID: 0}

	if id <= 0 {
		if statusId > 0 {
			statusID := int32(statusId)
			hasStatus := slices.ContainsFunc(statuses, func(s data.GetStatusesRow) bool {
				return s.ID == statusID
			})

			if hasStatus {
				form.StatusID = statusID
			}
		}

		if boardId > 0 {
			boardID := int32(boardId)
			hasBoard := slices.ContainsFunc(boards, func(b data.GetBoardsRow) bool {
				return b.ID == boardID
			})

			if hasBoard {
				form.BoardID = &boardID
			}
		}
	} else {
		post, err := a.RoadmapService.GetPostById(c.Context(), int32(id), curUser.Project.ID)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "error fetching post")
		}

		if post.BoardID.Valid {
			form.BoardID = &post.BoardID.Int32
		}

		if post.StatusID.Valid {
			form.StatusID = post.StatusID.Int32
		}

		if post.DueDate.Valid {
			form.DueDate = post.DueDate.Time.Format(time.DateTime)
		}

		form.Content = template.HTMLEscapeString(post.Body)
		form.Title = post.Title
		form.IsPrivate = post.IsPrivate
		form.ID = &post.ID
	}

	statuses = append([]data.GetStatusesRow{{ID: 0, Status: "Unassigned", SortOrder: -1, Color: "#DDDDDD"}}, statuses...)

	return c.Render("partials/components/roadmap/post_slideover", fiber.Map{"form": form, "Statuses": statuses, "Boards": boards})

}

func (a *AppHandler) GetRoadmapPostActivity(c *fiber.Ctx) error {
	curUser := c.Locals("user").(*models.SessionUser)
	postId, err := c.ParamsInt("postId")
	if err != nil || postId <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid post id provided")
	}

	var commentId *int32
	paramCommentId, err := c.ParamsInt("commentId")
	if err == nil {
		if paramCommentId > 0 {
			i32CommentId := int32(paramCommentId)
			commentId = &i32CommentId
		}
	}

	post, err := a.RoadmapService.GetRoadmapPostActivity(c.Context(), int32(postId), curUser.Project.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "error retrieving post")
	}

	firstActivity := models.RoadmapPostActivityModel{
		ID:        post.ID,
		CreatedOn: post.Createdon,
		Type:      models.ActivityTypeCreation,
		Who:       post.Who,
	}
	if post.Whopictureurl.Valid {
		firstActivity.WhoPictureUrl = post.Whopictureurl.String
	}

	activity := []models.RoadmapPostActivityModel{
		firstActivity,
	}

	statusUpdates, err := a.RoadmapService.GetRoadmapPostStatusActivity(c.Context(), int32(postId), curUser.Project.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "error retrieving status updates")
	}
	unassignedStatus := &models.RoadmapStatusModel{
		Color:  "#DDDDDD",
		Status: "Unassigned",
	}

	for _, u := range statusUpdates {
		statusUpdateActivity := &models.RoadmapPostStatusActivityModel{
			FromStatus: unassignedStatus,
			ToStatus:   unassignedStatus,
		}

		if u.Statusfromid.Valid {
			statusUpdateActivity.FromStatus = &models.RoadmapStatusModel{
				ID:          &u.Statusfromid.Int32,
				Status:      u.Statusfrom.String,
				Color:       u.Statusfromcolor.String,
				Description: u.Statusfromdescription.String,
			}
		}

		if u.Statustoid.Valid {
			statusUpdateActivity.ToStatus = &models.RoadmapStatusModel{
				ID:          &u.Statustoid.Int32,
				Status:      u.Statusto.String,
				Color:       u.Statustocolor.String,
				Description: u.Statustodescription.String,
			}
		}

		newActivity := models.RoadmapPostActivityModel{
			ID:                   u.ID,
			CreatedOn:            u.Createdon,
			Who:                  u.Who,
			Type:                 models.ActivityTypeStatusUpdate,
			StatusUpdateActivity: statusUpdateActivity,
		}

		if u.Whopictureurl.Valid {
			newActivity.WhoPictureUrl = u.Whopictureurl.String
		}

		activity = append(activity, newActivity)
	}

	comments, err := a.RoadmapService.GetRoadmapPostComments(c.Context(), int32(postId), curUser.Project.ID, commentId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "error retrieving comments")
	}
	reactions, err := a.RoadmapService.GetRoadmapPostReactions(c.Context(), int32(postId), curUser.Project.ID, commentId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "error retrieving reactions")
	}

	var postComments []models.RoadmapPostActivityModel
	for _, c := range comments {
		newActivity := models.RoadmapPostActivityModel{
			ID:        c.ID,
			CreatedOn: c.Createdon,
			Who:       c.Who,
			Type:      models.ActivityTypeComment,
			CommentActivity: &models.RoadmapPostCommentModel{
				IsPinned:        c.Ispinned,
				ReplyCount:      c.Replycount,
				ParentCommentID: commentId,
			},
		}

		if !c.Isdeleted {
			comment := c.Comment
			newActivity.CommentActivity.Comment = &comment
		}

		if c.Whopictureurl.Valid {
			newActivity.WhoPictureUrl = c.Whopictureurl.String
		}

		var postReactions []models.RoadmapPostReactionActivityModel
		for _, r := range reactions {
			newReaction := models.RoadmapPostReactionActivityModel{
				Who:             r.Who,
				Reaction:        r.Reaction,
				ParentCommentID: commentId,
			}

			if r.Whopictureurl.Valid {
				newReaction.WhoPictureUrl = &r.Whopictureurl.String
			}

			postReactions = append(postReactions, newReaction)
		}
		newActivity.CommentActivity.Reactions = postReactions

		activity = append(activity, newActivity)
		postComments = append(postComments, newActivity)
	}

	sort.Slice(activity, func(i int, j int) bool {
		return activity[i].CreatedOn.After(activity[j].CreatedOn)
	})

	if commentId != nil {
		return c.Render("partials/components/roadmap/post_comment_list", fiber.Map{"postId": postId, "activity": activity, "comments": postComments})
	}

	return c.Render("partials/components/roadmap/post_slideover_timeline", fiber.Map{"postId": postId, "activity": activity})

}

func (a *AppHandler) SaveRoadmapPostStatus(c *fiber.Ctx) error {
	curUser := c.Locals("user").(*models.SessionUser)

	postId, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "could not read post id")
	}

	post, err := a.RoadmapService.GetPostById(c.Context(), int32(postId), curUser.Project.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusForbidden, "post id does not belong to this user")
	}

	statusId, err := c.ParamsInt("statusId")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "could not read status id")
	}

	if (!post.StatusID.Valid && statusId <= 0) || (post.StatusID.Int32 == int32(statusId)) {
		return c.Render("partials/components/roadmap/board_post", fiber.Map{"p": post, "StatusUpdated": true})
	}

	statuses, err := a.RoadmapService.GetStatuses(c.Context(), curUser.Project.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "error when validating status")
	}
	hasStatus := slices.ContainsFunc[data.GetStatusesRow](statuses, func(row data.GetStatusesRow) bool {
		return row.ID == int32(statusId)
	})
	if !hasStatus && statusId > 0 {
		return fiber.NewError(fiber.StatusForbidden, "status does not belong to this user")
	}

	boardId, err := c.ParamsInt("boardId")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "could not read board id")
	}

	var BoardID *int32
	isSettingToUnassigned := post.StatusID.Valid && statusId == -1
	isSettingToAssigned := !post.StatusID.Valid && statusId > -1

	if !isSettingToAssigned && !isSettingToUnassigned && post.BoardID.Valid {
		BoardID = &post.BoardID.Int32
	} else if isSettingToAssigned {
		i32BoardId := int32(boardId)
		_, err := a.RoadmapService.GetBoard(c.Context(), i32BoardId, curUser.Project.ID)
		if err != nil {
			return fiber.NewError(fiber.StatusForbidden, "board does not belong to this user")
		}
		BoardID = &i32BoardId
	}

	saved, err := a.RoadmapService.UpdatePostStatus(c.Context(), int32(max(0, statusId)), BoardID, int32(postId), curUser.Project.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "error saving status")
	}

	var fromStatusId *int32
	var toStatusId *int32

	if post.StatusID.Valid {
		fromStatusId = &post.StatusID.Int32
	}

	if statusId > 0 {
		i32StatusId := int32(statusId)
		toStatusId = &i32StatusId
	}

	a.RoadmapService.InsertPostActivity(c.Context(), fromStatusId, toStatusId, post.ID, curUser.Author.ID)

	return c.Render("partials/components/roadmap/board_post", fiber.Map{"p": saved, "StatusUpdated": true})
}

func (a *AppHandler) PostRoadmapPostComment(c *fiber.Ctx) error {
	viewPath := "partials/components/roadmap/post_comment"
	curUser := c.Locals("user").(*models.SessionUser)
	viewerAny := c.Locals("viewer")

	postId, err := c.ParamsInt("postId")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "bad post id given")
	}

	commentId, _ := c.ParamsInt("commentId")
	var i32CommentId *int32
	if commentId > 0 {
		parsedCommentId := int32(commentId)
		i32CommentId = &parsedCommentId
	}

	form := new(models.RoadmapPostCommentFormModel)
	err = c.BodyParser(form)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid input")
	}

	var Who *string
	var authorId *int32
	var WhoPictureUrl *string
	if curUser != nil {
		authorId = &curUser.Author.ID
		Who = &curUser.Author.FirstName
		if curUser.Author.PictureUrl.Valid {
			WhoPictureUrl = &curUser.Author.PictureUrl.String
		}
	}
	var viewerId *int32
	if viewerAny != nil {
		viewer := viewerAny.(data.Viewer)
		viewerId = &viewer.ID

		if viewer.UserName.Valid {
			Who = &viewer.UserName.String
		}
	}

	if authorId == nil && viewerId == nil {
		return fiber.NewError(fiber.StatusForbidden, "author or viewer id required")
	}

	comment, err := a.RoadmapService.InsertRoadmapPostComment(c.Context(), int32(postId), i32CommentId, authorId, viewerId, form.Comment)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "error saving comment")
	}

	activity := models.RoadmapPostActivityModel{
		ID:            comment.ID,
		CreatedOn:     comment.CreatedOn,
		Who:           *Who,
		WhoPictureUrl: *WhoPictureUrl,
		Type:          models.ActivityTypeComment,
		CommentActivity: &models.RoadmapPostCommentModel{
			Comment:         &comment.Content,
			ReplyCount:      0,
			ParentCommentID: i32CommentId,
			IsPinned:        false,
			Reactions:       []models.RoadmapPostReactionActivityModel{},
		},
	}

	return c.Render(viewPath, fiber.Map{"a": activity, "postId": postId})
}

func (a *AppHandler) SaveRoadmapPost(c *fiber.Ctx) error {
	curUser := c.Locals("user").(*models.SessionUser)
	viewPath := "partials/components/roadmap/post_slideover_form"

	form := new(models.RoadmapPostModel)
	if err := c.BodyParser(form); err != nil {
		return c.Render(viewPath, fiber.Map{"error": err.Error()})
	}

	errs := make(map[string]string)
	if len(strings.TrimSpace(form.Title)) == 0 {
		errs["Title"] = "Title is required"
	}

	if len(strings.TrimSpace(form.Content)) == 0 {
		errs["Content"] = "Post content is required"
	}

	statuses, err := a.RoadmapService.GetStatuses(c.Context(), curUser.Project.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "error when fetching statuses")
	}

	if form.StatusID > 0 {
		hasStatus := slices.ContainsFunc[data.GetStatusesRow](statuses, func(d data.GetStatusesRow) bool {
			return d.ID == form.StatusID
		})

		if !hasStatus {
			errs["Status"] = "Status not found"
		}
	}

	boards, err := a.RoadmapService.GetBoards(c.Context(), curUser.Project.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "error when fetching boards")
	}

	if form.BoardID != nil {
		hasBoard := slices.ContainsFunc[data.GetBoardsRow](boards, func(d data.GetBoardsRow) bool {
			return d.ID == *form.BoardID
		})

		if !hasBoard {
			errs["Board"] = "Board not found"
		}
	}

	if len(errs) > 0 {
		form.Content = template.HTMLEscapeString(form.Content)
		return c.Render(viewPath, fiber.Map{"form": form, "errors": errs, "Statuses": statuses, "Boards": boards})
	}

	loc, err := time.LoadLocation(curUser.Timezone)
	if err != nil {
		return c.Render(viewPath, fiber.Map{"error": "Could not locate user timezone", "form": form, "Statuses": statuses, "Boards": boards})
	}

	var savedPost data.RoadmapPost
	if form.ID != nil && *form.ID > 0 {
		savedPost, err = a.RoadmapService.UpdatePost(c.Context(), *form, curUser.Project.ID, loc)
		if err != nil {
			form.Content = template.HTMLEscapeString(form.Content)
			return c.Render(viewPath, fiber.Map{"error": err.Error(), "form": form, "Statuses": statuses, "Boards": boards})
		}
	} else {
		author := curUser.Author
		var userUuid *uuid.UUID
		if c.Locals("userUuid") != nil {
			userUuid = c.Locals("userUuid").(*uuid.UUID)
		}

		if author == nil && userUuid == nil {
			return c.Render(viewPath, fiber.Map{"error": "Author or user uuid not supplied", "form": form, "Statuses": statuses, "Boards": boards})

		}

		savedPost, err = a.RoadmapService.InsertPost(c.Context(), *form, &author.ID, nil, curUser.Project.ID, loc, false)
		if err != nil {
			form.Content = template.HTMLEscapeString(form.Content)
			return c.Render(viewPath, fiber.Map{"error": err.Error(), "form": form, "Statuses": statuses, "Boards": boards})
		}
	}

	savedPost.Body = template.HTMLEscapeString(form.Content)
	return c.Render(viewPath, fiber.Map{"Post": savedPost, "Statuses": statuses, "Boards": boards, "Success": true, "Message": "Post saved successfully.", "Close": true})
}

func (a *AppHandler) DeleteRoadmapPost(c *fiber.Ctx) error {
	curUser := c.Locals("user").(*models.SessionUser)

	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Render("partials/components/banner", fiber.Map{"Message": "Invalid id parameter supplied"})
	}

	if _, err := a.RoadmapService.DeletePost(c.Context(), int32(id), curUser.Project.ID); err != nil {
		fmt.Println(err.Error())
		return c.Render("partials/components/banner", fiber.Map{"Message": "An error occured when deleting the post"})
	}

	c.Append("HX-Trigger-After-Swap", fmt.Sprintf("{\"reset-counters\": \"post-%d\"}", id))
	return c.Render("partials/components/banner", fiber.Map{"Message": "Post successfully deleted", "Success": true})
}

func (a *AppHandler) ConfirmDeleteRoadmapPost(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(400, "Invalid id parameter supplied")
	}

	return c.Render("partials/components/delete_confirm_modal", fiber.Map{"Title": "Confirm deletion",
		"Body":        "Are you sure you want to delete this post",
		"EndpointUri": "/admin/roadmap/post/delete/" + fmt.Sprint(id),
		"IsSlideOver": true,
	})
}

func (a *AppHandler) GetUserAnalytics(c *fiber.Ctx) error {
	curUser := c.Locals("user").(*models.SessionUser)

	data, err := a.PostService.GetAnalytics(c.Context(), curUser.Project.ID, nil)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "error when fetching analytics")
	}

	return c.Render("analytics", fiber.Map{"data": data, "AnalyticsEmpty": len(data) == 0})
}

func (a *AppHandler) GetAnalyticsByUser(c *fiber.Ctx) error {
	curUser := c.Locals("user").(*models.SessionUser)
	qViewerId, err := c.ParamsInt("viewerId")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid viewer id")
	}

	i32ViewerId := int32(qViewerId)

	data, err := a.PostService.GetAnalytics(c.Context(), curUser.Project.ID, &i32ViewerId)
	if err != nil || len(data) == 0 {
		return fiber.NewError(fiber.StatusInternalServerError, "error when fetching analytics")
	}

	comments, err := a.PostService.GetPostComments(c.Context(), curUser.Project.ID, nil, &i32ViewerId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "error when fetching analytics comments")
	}

	reactions, err := a.PostService.GetPostReactions(c.Context(), curUser.Project.ID, nil, &i32ViewerId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "error when fetching analytics reactons")
	}

	return c.Render("partials/components/analytics/user_reactions_slideover", fiber.Map{"Data": data[0], "Comments": comments, "CommentCount": len(comments), "Reactions": reactions})
}
