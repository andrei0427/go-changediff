package services

import (
	"context"
	"database/sql"
	"errors"

	"github.com/andrei0427/go-changediff/internal/app/models"
	"github.com/andrei0427/go-changediff/internal/data"
)

type RoadmapService struct {
	db *data.Queries
}

func NewRoadmapSercice(db *data.Queries) *RoadmapService {
	return &RoadmapService{db: db}
}

func (s *RoadmapService) GetBoards(ctx context.Context, projectId int32) ([]data.GetBoardsRow, error) {
	return s.db.GetBoards(ctx, projectId)
}

func (s *RoadmapService) GetBoard(ctx context.Context, id int32, projectId int32) (data.GetBoardRow, error) {
	return s.db.GetBoard(ctx, data.GetBoardParams{ID: id, ProjectID: projectId})
}

func (s *RoadmapService) SaveBoard(ctx context.Context, model models.RoadmapBoardModel, project_id int32) (data.RoadmapBoard, error) {
	if model.ID != nil {
		toUpdate := data.UpdateBoardParams{
			ID:          *model.ID,
			Name:        model.Name,
			IsPrivate:   model.IsPrivate,
			Description: model.Description,
			ProjectID:   project_id,
		}

		return s.db.UpdateBoard(ctx, toUpdate)

	}

	toInsert := data.InsertBoardParams{
		Name:        model.Name,
		IsPrivate:   model.IsPrivate,
		Description: model.Description,
		ProjectID:   project_id,
	}

	return s.db.InsertBoard(ctx, toInsert)
}

func (s *RoadmapService) DeleteBoard(ctx context.Context, boardId int32, projectId int32) error {
	posts, err := s.db.HasPostsForBoard(ctx, sql.NullInt32{Int32: boardId, Valid: true})
	if err != nil {
		return err
	}

	if posts > 0 {
		return errors.New("there are posts inside this board, please delete them first")
	}

	_, err = s.db.DeleteBoard(ctx, data.DeleteBoardParams{ID: boardId, ProjectID: projectId})
	return err
}

func (s *RoadmapService) GetStatuses(ctx context.Context, projectId int32) ([]data.GetStatusesRow, error) {
	return s.db.GetStatuses(ctx, projectId)
}

func (s *RoadmapService) GetStatus(ctx context.Context, id int32, projectId int32) (data.GetStatusRow, error) {
	return s.db.GetStatus(ctx, data.GetStatusParams{ID: id, ProjectID: projectId})
}

func (s *RoadmapService) SaveStatus(ctx context.Context, model models.RoadmapStatusModel, project_id int32) (data.RoadmapStatus, error) {
	if model.ID != nil {
		toUpdate := data.UpdateStatusParams{
			ID:          *model.ID,
			Status:      model.Status,
			Color:       model.Color,
			Description: model.Description,
			ProjectID:   project_id,
		}

		return s.db.UpdateStatus(ctx, toUpdate)

	}

	toInsert := data.InsertStatusParams{
		Status:      model.Status,
		Color:       model.Color,
		Description: model.Description,
		ProjectID:   project_id,
	}

	return s.db.InsertStatus(ctx, toInsert)
}

func (s *RoadmapService) DeleteStatus(ctx context.Context, boardId int32, projectId int32) error {
	posts, err := s.db.HasPostsForStatus(ctx, boardId)
	if err != nil {
		return err
	}

	if posts > 0 {
		return errors.New("there are posts having this status, please delete them first")
	}

	_, err = s.db.DeleteStatus(ctx, data.DeleteStatusParams{ID: boardId, ProjectID: projectId})
	return err
}
