package services

import (
	"context"

	"github.com/andrei0427/go-changediff/internal/app/models"
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

func (s *LabelService) DeleteLabel(ctx context.Context, labelId int32, projectId int32) error {
	if _, err := s.db.UnsetLabels(ctx, data.UnsetLabelsParams{ID: labelId, ProjectID: projectId}); err != nil {
		return err
	}

	if _, err := s.db.DeleteLabel(ctx, data.DeleteLabelParams{ID: labelId, ProjectID: projectId}); err != nil {
		return err
	}

	return nil

}

func (s *LabelService) SaveLabel(ctx context.Context, model models.LabelModel, project_id int32) (data.Label, error) {
	if model.ID != nil {
		toUpdate := data.UpdateLabelParams{
			ID:        *model.ID,
			Label:     model.Label,
			Color:     model.Color,
			ProjectID: project_id,
		}

		return s.db.UpdateLabel(ctx, toUpdate)
	} else {
		toInsert := data.InsertLabelParams{
			Label:     model.Label,
			Color:     model.Color,
			ProjectID: project_id,
		}

		return s.db.InsertLabel(ctx, toInsert)
	}
}
