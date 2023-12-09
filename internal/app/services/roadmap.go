package services

import (
	"context"
	"database/sql"
	"errors"
	"time"

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
		nextSortOrder = 0
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

func (s *RoadmapService) GetPostById(ctx context.Context, postId int32, projectId int32) (data.GetRoadmapPostRow, error) {
	return s.db.GetRoadmapPost(ctx, data.GetRoadmapPostParams{ID: postId, ProjectID: projectId})
}

func (s *RoadmapService) InsertPost(ctx context.Context, post models.RoadmapPostModel, authorId *int32, viewerId *int32, projectId int32, userLocation *time.Location, isIdea bool) (data.RoadmapPost, error) {
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
	} else if viewerId != nil {
		toInsert.ViewerID = sql.NullInt32{Int32: *viewerId, Valid: true}
	} else {
		return data.RoadmapPost{}, errors.New("eithor author id or user uuid are required")
	}

	return s.db.InsertRoadmapPost(ctx, toInsert)
}

func (s *RoadmapService) DeletePost(ctx context.Context, postId int32, projectId int32) (bool, error) {
	tx, err := s.sql.Begin()
	if err != nil {
		return false, err
	}

	qtx := s.db.WithTx(tx)

	_, err = qtx.DeleteRoadmapPostCategoriesByPost(ctx, data.DeleteRoadmapPostCategoriesByPostParams{RoadmapPostID: postId, ProjectID: projectId})
	if err != nil {
		tx.Rollback()
		return false, err
	}

	_, err = s.DeleteRoadmapPostVote(ctx, nil, postId, projectId, nil, nil, qtx)
	if err != nil {
		tx.Rollback()
		return false, err
	}

	_, err = s.DeleteRoadmapPostActivity(ctx, postId, projectId, nil, nil, qtx)
	if err != nil {
		tx.Rollback()
		return false, err
	}

	_, err = s.DeleteRoadmapPostComments(ctx, postId, projectId, qtx)
	if err != nil {
		tx.Rollback()
		return false, err
	}

	_, err = qtx.DeleteRoadmapPost(ctx, data.DeleteRoadmapPostParams{ID: postId, ProjectID: projectId})
	if err != nil {
		tx.Rollback()
		return false, err
	}

	tx.Commit()

	return true, err
}

func (s *RoadmapService) InsertPostActivity(ctx context.Context, fromStatusId *int32, toStatusId *int32, postId int32, authorId int32) (data.RoadmapPostActivity, error) {
	params := data.InsertRoadmapPostActivityParams{
		RoadmapPostID: postId,
		AuthorID:      authorId,
	}

	if fromStatusId != nil {
		params.FromStatusID = sql.NullInt32{Int32: *fromStatusId, Valid: true}
	}

	if toStatusId != nil {
		params.ToStatusID = sql.NullInt32{Int32: *toStatusId, Valid: true}
	}

	return s.db.InsertRoadmapPostActivity(ctx, params)
}

func (s *RoadmapService) UpdatePostStatus(ctx context.Context, statusId int32, boardId *int32, postId int32, projectId int32) (data.RoadmapPost, error) {
	BoardID := sql.NullInt32{Valid: false}
	if boardId != nil {
		BoardID = sql.NullInt32{Int32: *boardId, Valid: true}
	}

	return s.db.UpdateRoadmapPostStatus(ctx, data.UpdateRoadmapPostStatusParams{StatusID: sql.NullInt32{Int32: statusId, Valid: statusId > 0}, BoardID: BoardID, ID: postId, ProjectID: projectId})
}

func (s *RoadmapService) GetAvailableReactions() []string {
	return []string{
		"👍", "❤️", "🙁", "😡",
	}
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

func (s *RoadmapService) GetRoadmapPostActivity(ctx context.Context, postId int32, projectId int32) (data.GetRoadmapPostActivityRow, error) {
	return s.db.GetRoadmapPostActivity(ctx, data.GetRoadmapPostActivityParams{ID: postId, ProjectID: projectId})
}

func (s *RoadmapService) GetRoadmapPostStatusActivity(ctx context.Context, postId int32, projectId int32) ([]data.GetRoadmapPostStatusActivityRow, error) {
	return s.db.GetRoadmapPostStatusActivity(ctx, data.GetRoadmapPostStatusActivityParams{RoadmapPostID: postId, ProjectID: projectId})
}

func (s *RoadmapService) GetRoadmapPostComments(ctx context.Context, postId int32, projectId int32, commentId *int32) ([]data.GetRoadmapPostCommentsRow, error) {
	params := data.GetRoadmapPostCommentsParams{
		RoadmapPostID: postId,
		ProjectID:     projectId,
		Column2:       0,
	}

	if commentId != nil {
		params.Column2 = *commentId
	}

	return s.db.GetRoadmapPostComments(ctx, params)
}

func (s *RoadmapService) GetRoadmapPostReactions(ctx context.Context, postId int32, projectId int32, commentId *int32, authorId *int32, viewerId *int32, emoji *string) ([]data.GetRoadmapPostReactionsRow, error) {
	params := data.GetRoadmapPostReactionsParams{
		RoadmapPostID: postId,
		ProjectID:     projectId,
		Column2:       0,
		ID:            0,
		ID_2:          0,
		Column6:       "",
	}

	if commentId != nil {
		params.Column2 = *commentId
	}

	if authorId != nil {
		params.ID = *authorId
	}

	if viewerId != nil {
		params.ID_2 = *viewerId
	}

	if emoji != nil {
		params.Column6 = *emoji
	}

	return s.db.GetRoadmapPostReactions(ctx, params)
}

func (s *RoadmapService) DeleteRoadmapPostActivity(ctx context.Context, postId int32, projectId int32, authorId *int32, viewerId *int32, qtx *data.Queries) (*[]int32, error) {
	params := data.DeleteRoadmapPostActivityParams{
		ID:        postId,
		ProjectID: projectId,

		Column3: 0,
		Column4: 0,
	}

	if authorId != nil {
		params.Column3 = *authorId
	}

	if viewerId != nil {
		params.Column4 = *viewerId
	}

	queries := qtx
	if queries == nil {
		queries = s.db
	}

	deleted, err := queries.DeleteRoadmapPostActivity(ctx, params)
	if err != nil {
		return nil, err
	}

	return &deleted, nil
}

func (s *RoadmapService) DeleteRoadmapPostVote(ctx context.Context, id *int32, postId int32, projectId int32, authorId *int32, viewerId *int32, qtx *data.Queries) (*[]int32, error) {
	params := data.DeleteRoadmapPostVoteParams{
		ID:        postId,
		ProjectID: projectId,

		Column4: 0,
		Column5: 0,
	}

	if id != nil {
		params.Column1 = *id
	}

	if authorId != nil {
		params.Column4 = *authorId
	}

	if viewerId != nil {
		params.Column5 = *viewerId
	}

	queries := qtx
	if queries == nil {
		queries = s.db
	}

	deleted, err := queries.DeleteRoadmapPostVote(ctx, params)
	if err != nil {
		return nil, err
	}

	return &deleted, nil
}

func (s *RoadmapService) DeleteRoadmapPostReaction(ctx context.Context, emoji *string, postId int32, projectId int32, qtx *data.Queries) (*[]int32, error) {
	params := data.DeleteRoadmapPostReactionParams{
		ID:        postId,
		ProjectID: projectId,
	}

	if emoji != nil {
		params.Column1 = *emoji
	}

	queries := qtx
	if queries == nil {
		queries = s.db
	}

	deleted, err := queries.DeleteRoadmapPostReaction(ctx, params)
	if err != nil {
		return nil, err
	}

	return &deleted, nil
}

func (s *RoadmapService) InsertRoadmapPostReaction(ctx context.Context, postId int32, emoji string, parentCommentId *int32, authorId *int32, viewerId *int32) (*data.RoadmapPostReaction, error) {
	params := data.InsertRoadmapPostReactionParams{
		RoadmapPostID: postId,
		Emoji:         emoji,
		AuthorID:      sql.NullInt32{Valid: false},
		ViewerID:      sql.NullInt32{Valid: false},
		CommentID:     sql.NullInt32{Valid: false},
	}

	if authorId != nil {
		params.AuthorID = sql.NullInt32{Int32: *authorId, Valid: true}
	}

	if viewerId != nil {
		params.ViewerID = sql.NullInt32{Int32: *viewerId, Valid: true}
	}

	if !params.AuthorID.Valid && !params.ViewerID.Valid {
		return nil, errors.New("author or viewer id required")
	}

	if parentCommentId != nil {
		params.CommentID = sql.NullInt32{Int32: *parentCommentId, Valid: true}
	}

	inserted, err := s.db.InsertRoadmapPostReaction(ctx, params)
	if err != nil {
		return nil, err
	}

	return &inserted, err
}

func (s *RoadmapService) InsertRoadmapPostVote(ctx context.Context, postId int32, projectId int32, authorId *int32, viewerId *int32) (*data.RoadmapPostVote, error) {
	params := data.InsertRoadmapPostVoteParams{
		RoadmapPostID: postId,
		AuthorID:      sql.NullInt32{Valid: false},
		ViewerID:      sql.NullInt32{Valid: false},
		ProjectID:     projectId,
	}

	if authorId != nil {
		params.AuthorID = sql.NullInt32{Int32: *authorId, Valid: true}
	}

	if viewerId != nil {
		params.ViewerID = sql.NullInt32{Int32: *viewerId, Valid: true}
	}

	if !params.AuthorID.Valid && !params.ViewerID.Valid {
		return nil, errors.New("author or viewer id required")
	}

	inserted, err := s.db.InsertRoadmapPostVote(ctx, params)
	if err != nil {
		return nil, err
	}

	return &inserted, err
}

func (s *RoadmapService) InsertRoadmapPostComment(ctx context.Context, postId int32, replyingToCommentId *int32, authorId *int32, viewerId *int32, content string) (*data.RoadmapPostComment, error) {
	params := data.InsertRoadmapPostCommentParams{
		RoadmapPostID: postId,
		AuthorID:      sql.NullInt32{Valid: false},
		ViewerID:      sql.NullInt32{Valid: false},
		InReplyToID:   sql.NullInt32{Valid: false},
		Content:       content,
	}

	if authorId != nil {
		params.AuthorID = sql.NullInt32{Int32: *authorId, Valid: true}
	}

	if viewerId != nil {
		params.ViewerID = sql.NullInt32{Int32: *viewerId, Valid: true}
	}

	if !params.AuthorID.Valid && !params.ViewerID.Valid {
		return nil, errors.New("author or viewer id required")
	}

	if replyingToCommentId != nil {
		params.InReplyToID = sql.NullInt32{Int32: *replyingToCommentId, Valid: true}
	}

	inserted, err := s.db.InsertRoadmapPostComment(ctx, params)
	if err != nil {
		return nil, err
	}

	return &inserted, err

}

func (s *RoadmapService) DeleteRoadmapPostComment(ctx context.Context, id int32, postId int32, projectId int32, authorId *int32, viewerId *int32, qtx *data.Queries) (int32, error) {
	params := data.DeleteRoadmapPostCommentParams{
		ID:        id,
		ID_2:      postId,
		ProjectID: projectId,

		Column4: 0,
		Column5: 0,
	}

	if authorId != nil {
		params.Column4 = *authorId
	}

	if viewerId != nil {
		params.Column5 = *viewerId
	}

	queries := qtx
	if queries == nil {
		queries = s.db

	}

	return queries.DeleteRoadmapPostComment(ctx, params)
}

func (s *RoadmapService) DeleteRoadmapPostComments(ctx context.Context, postId int32, projectId int32, qtx *data.Queries) ([]int32, error) {
	params := data.DeleteAllRoadmapPostCommentsParams{
		ID:        postId,
		ProjectID: projectId,
	}

	queries := qtx
	if queries == nil {
		queries = s.db

	}

	return queries.DeleteAllRoadmapPostComments(ctx, params)
}
