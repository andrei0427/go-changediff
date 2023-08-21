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

func (s *ProjectService) SaveProject(ctx context.Context, userId uuid.UUID, project models.OnboardingModel, imageUrl *string) (data.Project, error) {
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

	return inserted, err
}
