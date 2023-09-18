package services

import (
	"context"
	"database/sql"

	"github.com/andrei0427/go-changediff/internal/app/models"
	"github.com/andrei0427/go-changediff/internal/data"
	"github.com/google/uuid"
)

type ProjectService struct {
	db *data.Queries
}

func NewProjectService(db *data.Queries) *ProjectService {
	return &ProjectService{db: db}
}

func (s *ProjectService) GetProjectForUser(ctx context.Context, userId uuid.UUID) (*data.Project, error) {
	projects, err := s.db.GetProject(ctx, userId)

	if len(projects) == 0 {
		return nil, err

	}

	return &projects[0], err
}

func (s *ProjectService) GetProjectByKey(ctx context.Context, key string) (data.Project, error) {
	project, err := s.db.GetProjectByKey(ctx, key)
	return project, err
}

func (s *ProjectService) SaveProject(ctx context.Context, userId uuid.UUID, project models.ProjectModel, imageUrl *string) (data.Project, error) {
	if project.ID != nil {
		toUpdate := data.UpdateProjectParams{
			Name:        project.Name,
			Description: project.Description,
			AccentColor: project.AccentColor,
			ID:          *project.ID,
			LogoUrl:     sql.NullString{},
			UserID:      userId,
		}

		if imageUrl != nil {
			toUpdate.LogoUrl = sql.NullString{String: *imageUrl, Valid: true}
		}

		return s.db.UpdateProject(ctx, toUpdate)

	} else {
		toInsert := data.InsertProjectParams{
			Name:        project.Name,
			Description: project.Description,
			AccentColor: project.AccentColor,
			UserID:      userId,
			AppKey:      uuid.NewString(),
			LogoUrl:     sql.NullString{},
		}

		if imageUrl != nil {
			toInsert.LogoUrl = sql.NullString{String: *imageUrl}
		}

		inserted, err := s.db.InsertProject(ctx, toInsert)

		defaultLabels := []data.InsertLabelParams{
			{Label: "Release", Color: "#10B981", ProjectID: inserted.ID},
			{Label: "Fix", Color: "#EF4444", ProjectID: inserted.ID},
			{Label: "Announcement", Color: "#3B82F6", ProjectID: inserted.ID},
		}

		labelService := NewLabelService(s.db)
		for _, l := range defaultLabels {
			labelService.SaveLabel(ctx, models.LabelModel{Label: l.Label, Color: l.Color}, l.ProjectID)
		}

		return inserted, err
	}
}
