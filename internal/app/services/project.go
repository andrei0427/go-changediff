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

func (s *ProjectService) GetProjectsForUser(ctx context.Context, userId uuid.UUID) ([]data.Project, error) {
	projects, err := s.db.GetProjects(ctx, userId)
	return projects, err
}
