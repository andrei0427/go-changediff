package services

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/andrei0427/go-changediff/internal/app/models"
	"github.com/andrei0427/go-changediff/internal/data"
	"github.com/google/uuid"
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

func (s *RoadmapService) GetPostById(ctx context.Context, postId int32, projectId int32) (data.RoadmapPost, error) {
	return s.db.GetRoadmapPost(ctx, data.GetRoadmapPostParams{ID: postId, ProjectID: projectId})
}

func (s *RoadmapService) InsertPost(ctx context.Context, post models.RoadmapPostModel, authorId *int32, userUuid *uuid.UUID, projectId int32, userLocation *time.Location, isIdea bool) (data.RoadmapPost, error) {
	toInsert := data.InsertRoadmapPostParams{
		Title:     post.Title,
		Body:      post.Content,
		ProjectID: projectId,
		IsPrivate: post.IsPrivate,
		IsIdea:    isIdea,
		DueDate:   sql.NullTime{Valid: false},
	}

	if len(post.DueDate) > 0 {
		parsedDate, err := time.ParseInLocation("2006-01-02T15:04", post.DueDate, userLocation)
		if err != nil {
			return data.RoadmapPost{}, errors.New("error parsing publish date")
		}

		toInsert.DueDate = sql.NullTime{Time: parsedDate.UTC(), Valid: true}
	}

	if post.StatusID > 0 {
		toInsert.StatusID = sql.NullInt32{Int32: int32(post.StatusID), Valid: true}
	}

	if post.BoardID != nil && *post.BoardID > 0 {
		toInsert.BoardID = sql.NullInt32{Int32: int32(*post.BoardID), Valid: true}
	}

	if authorId != nil {
		toInsert.AuthorID = sql.NullInt32{Int32: *authorId, Valid: true}
	} else if userUuid != nil {
		toInsert.UserUuid = uuid.NullUUID{UUID: *userUuid, Valid: true}
	} else {
		return data.RoadmapPost{}, errors.New("eithor author id or user uuid are required")
	}

	return s.db.InsertRoadmapPost(ctx, toInsert)
}

func (s *RoadmapService) UpdatePost(ctx context.Context, post models.RoadmapPostModel, projectId int32, userLocation *time.Location) (data.RoadmapPost, error) {
	if post.ID == nil {
		return data.RoadmapPost{}, errors.New("ID is required when updating")
	}

	toUpdate := data.UpdateRoadmapPostParams{
		ID:        *post.ID,
		Title:     post.Title,
		Body:      post.Content,
		ProjectID: projectId,
		IsPrivate: post.IsPrivate,
		DueDate:   sql.NullTime{Valid: false},
	}

	if len(post.DueDate) > 0 {
		parsedDate, err := time.ParseInLocation("2006-01-02T15:04", post.DueDate, userLocation)
		if err != nil {
			return data.RoadmapPost{}, errors.New("error parsing publish date")
		}

		toUpdate.DueDate = sql.NullTime{Time: parsedDate.UTC(), Valid: true}
	}

	if post.StatusID > 0 {
		toUpdate.StatusID = sql.NullInt32{Int32: int32(post.StatusID), Valid: true}
	}

	if post.BoardID != nil && *post.BoardID > 0 {
		toUpdate.BoardID = sql.NullInt32{Int32: int32(*post.BoardID), Valid: true}
	}

	return s.db.UpdateRoadmapPost(ctx, toUpdate)
}
