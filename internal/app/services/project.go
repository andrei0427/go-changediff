package services

import (
	"context"

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
