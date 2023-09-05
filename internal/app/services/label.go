package services

import (
	"context"

	"github.com/andrei0427/go-changediff/internal/data"
)

type LabelService struct {
	db *data.Queries
}

func NewLabelService(db *data.Queries) *LabelService {
	return &LabelService{db: db}
}

func (s *LabelService) GetLabels(ctx context.Context, projectId int32) ([]data.Label, error) {
	labels, err := s.db.GetLabels(ctx, projectId)
	return labels, err
}

func (s *LabelService) InsertLabel(ctx context.Context, label string, color string, project_id int32) (data.Label, error) {
	inserted, err := s.db.InsertLabel(ctx, data.InsertLabelParams{Label: label, Color: color, ProjectID: project_id})
	return inserted, err
}
