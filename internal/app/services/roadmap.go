package services

import (
	"context"
	"database/sql"
	"errors"

	"github.com/andrei0427/go-changediff/internal/app/models"
	"github.com/andrei0427/go-changediff/internal/data"
)

type RoadmapService struct {
	db  *data.Queries
	sql *sql.DB
}

func NewRoadmapService(db *data.Queries, sql *sql.DB) *RoadmapService {
	return &RoadmapService{db: db, sql: sql}
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

func (s *RoadmapService) UpdateStatusSortOrder(ctx context.Context, up bool, statusId int32, projectId int32) (*[]data.UpdateStatusOrderRow, error) {
	statuses, err := s.GetStatuses(ctx, projectId)
	if err != nil {
		return nil, err
	}

	statusToMove := data.GetStatusesRow{ID: 0}
	for _, status := range statuses {
		if status.ID == statusId {
			statusToMove = status
		}
	}

	newSortOrder := statusToMove.SortOrder
	if up {
		newSortOrder--
	} else {
		newSortOrder++
	}

	if statusToMove.ID == 0 {
		return nil, errors.New("status not found")
	}

	statuses = append(statuses[:statusToMove.SortOrder], statuses[statusToMove.SortOrder+1:]...)
	statuses = append(statuses[:newSortOrder], append([]data.GetStatusesRow{statusToMove}, statuses[newSortOrder:]...)...)

	tx, err := s.sql.Begin()
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	qtx := s.db.WithTx(tx)
	var updatedStatuses []data.UpdateStatusOrderRow
	for i, status := range statuses {
		updated, err := qtx.UpdateStatusOrder(ctx, data.UpdateStatusOrderParams{SortOrder: int32(i), ID: status.ID, ProjectID: projectId})

		if err != nil {
			return nil, errors.New("could not update status")
		}

		updatedStatuses = append(updatedStatuses, updated)
	}

	tx.Commit()
	return &updatedStatuses, nil
}

func (s *RoadmapService) SaveStatus(ctx context.Context, model models.RoadmapStatusModel, projectId int32) (*data.RoadmapStatus, error) {
	if model.ID != nil {
		toUpdate := data.UpdateStatusParams{
			ID:          *model.ID,
			Status:      model.Status,
			Color:       model.Color,
			Description: model.Description,
			ProjectID:   projectId,
		}

		updated, err := s.db.UpdateStatus(ctx, toUpdate)
		return &updated, err

	}

	nextSortOrder, err := s.db.GetNextSortOrderForStatus(ctx, projectId)
	if err != nil {
		return nil, err
	}

	toInsert := data.InsertStatusParams{
		Status:      model.Status,
		Color:       model.Color,
		Description: model.Description,
		ProjectID:   projectId,
		SortOrder:   nextSortOrder,
	}

	inserted, err := s.db.InsertStatus(ctx, toInsert)
	return &inserted, err
}

func (s *RoadmapService) DeleteStatus(ctx context.Context, boardId int32, projectId int32) error {
	posts, err := s.db.HasPostsForStatus(ctx, sql.NullInt32{Int32: boardId, Valid: true})
	if err != nil {
		return err
	}

	if posts > 0 {
		return errors.New("there are posts having this status, please delete them first")
	}

	_, err = s.db.DeleteStatus(ctx, data.DeleteStatusParams{ID: boardId, ProjectID: projectId})
	return err
}

func (s *RoadmapService) GetPostsForBoard(ctx context.Context, boardId int32, projectId int32) ([]data.GetPostsForBoardRow, error) {
	return s.db.GetPostsForBoard(ctx, data.GetPostsForBoardParams{BoardID: sql.NullInt32{Int32: boardId, Valid: true}, ProjectID: projectId})
}
