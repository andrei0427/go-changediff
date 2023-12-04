package services

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/andrei0427/go-changediff/internal/app/models"
	"github.com/andrei0427/go-changediff/internal/data"
)

type PostService struct {
	db  *data.Queries
	sql *sql.DB
}

func NewPostService(db *data.Queries, sql *sql.DB) *PostService {
	return &PostService{db: db, sql: sql}
}

func (s *PostService) GetPostCountForProject(ctx context.Context, projectId int32) (int64, error) {
	posts, err := s.db.GetPostCount(ctx, projectId)
	return posts, err
}

func (s *PostService) GetPosts(ctx context.Context, projectId int32) ([]data.GetPostsRow, error) {
	posts, err := s.db.GetPosts(ctx, projectId)
	return posts, err
}

func (s *PostService) GetPost(ctx context.Context, postId int32, projectId int32) (data.GetPostRow, error) {
	return s.db.GetPost(ctx, data.GetPostParams{ID: postId, ProjectID: projectId})
}

func (s *PostService) GetPostReactions(ctx context.Context, projectId int32, postId *int32, viewerId *int32) ([]data.GetPostReactionsRow, error) {
	params := data.GetPostReactionsParams{
		ProjectID: projectId,
	}

	if postId != nil {
		params.Column2 = sql.NullInt32{Int32: *postId, Valid: true}
	}

	if viewerId != nil {
		params.Column3 = sql.NullInt32{Int32: *viewerId, Valid: true}
	}

	return s.db.GetPostReactions(ctx, params)
}

func (s *PostService) GetPostComments(ctx context.Context, projectId int32, postId *int32, viewerId *int32) ([]data.GetPostCommentsRow, error) {
	params := data.GetPostCommentsParams{ProjectID: projectId, Column2: 0, Column3: 0}

	if viewerId != nil {
		params.Column2 = *viewerId
	}

	if postId != nil {
		params.Column3 = *postId
	}

	return s.db.GetPostComments(ctx, params)
}

func (s *PostService) GetAnalytics(ctx context.Context, projectId int32, viewerId *int32) ([]data.AnalyticsUsersRow, error) {
	params := data.AnalyticsUsersParams{
		ProjectID: projectId,
		Column2:   0,
	}

	if viewerId != nil {
		params.Column2 = *viewerId
	}

	return s.db.AnalyticsUsers(ctx, params)
}

func (s *PostService) GetPublishedPagedPosts(ctx context.Context, projectKey string, pageNo int32, search string, viewerId int32) ([]data.GetPublishedPagedPostsRow, error) {
	var offset int32 = 0
	if pageNo > 1 {
		offset = pageNo - 1*5
	}

	searchStr := search
	if len(searchStr) > 0 {
		searchStr = "%" + strings.ToLower(search) + "%"
	}

	posts, err := s.db.GetPublishedPagedPosts(ctx, data.GetPublishedPagedPostsParams{AppKey: projectKey, Limit: 5, Offset: offset, ViewerID: viewerId, Column5: searchStr})

	return posts, err
}

func (s *PostService) InsertPostComment(ctx context.Context, viewerId int32, postId int32, projectId int32, comment string) (data.ChangelogInteraction, error) {
	return s.db.InsertInteraction(ctx, data.InsertInteractionParams{
		ViewerID:          viewerId,
		Content:           sql.NullString{String: comment, Valid: true},
		PostID:            postId,
		ProjectID:         projectId,
		InteractionTypeID: int32(models.InteractionTypeComment),
	})
}

func (s *PostService) InsertPost(ctx context.Context, post models.PostModel, authorId int32, projectId int32, userLocation *time.Location) (data.Post, error) {
	toInsert := data.InsertPostParams{
		Title:       post.Title,
		Body:        post.Content,
		AuthorID:    authorId,
		ProjectID:   projectId,
		IsPublished: sql.NullBool{Bool: post.IsPublished, Valid: true},
		LabelID:     sql.NullInt32{Valid: false},
	}

	parsedDate, err := time.ParseInLocation("2006-01-02T15:04", post.PublishedOn, userLocation)
	if err != nil {
		return data.Post{}, errors.New("error parsing publish date")
	}

	toInsert.PublishedOn = parsedDate.UTC()

	if post.LabelId != nil && *post.LabelId > 0 {
		toInsert.LabelID = sql.NullInt32{Int32: int32(*post.LabelId), Valid: true}
	}

	return s.db.InsertPost(ctx, toInsert)
}

func (s *PostService) DeletePost(ctx context.Context, postId int32, projectId int32) (bool, error) {

	tx, err := s.sql.Begin()
	if err != nil {
		return false, err
	}

	defer tx.Rollback()

	qtx := s.db.WithTx(tx)
	_, err = qtx.DeleteInteractions(ctx, data.DeleteInteractionsParams{PostID: postId, ProjectID: projectId})
	if err != nil {
		return false, err
	}

	_, err = qtx.DeletePost(ctx, data.DeletePostParams{ID: postId, ProjectID: projectId})
	if err != nil {
		return false, err
	}

	tx.Commit()
	return true, nil
}

func (s *PostService) UpdatePost(ctx context.Context, post models.PostModel, projectId int32, userLocation *time.Location) (data.Post, error) {
	toUpdate := data.UpdatePostParams{
		Title:       post.Title,
		Body:        post.Content,
		ProjectID:   projectId,
		IsPublished: sql.NullBool{Bool: post.IsPublished, Valid: true},
	}

	if post.ExpiresOn != "" {
		parsedExpiryDate, err := time.ParseInLocation("2006-01-02T15:04", post.ExpiresOn, userLocation)
		if err != nil {
			return data.Post{}, errors.New("error parsing expiry date")
		}

		toUpdate.ExpiresOn = sql.NullTime{Time: parsedExpiryDate.UTC(), Valid: true}
	}

	if post.ID != nil {
		toUpdate.ID = int32(*post.ID)
	} else {
		return data.Post{}, errors.New("ID is required when updating")
	}

	parsedDate, err := time.ParseInLocation("2006-01-02T15:04", post.PublishedOn, userLocation)
	if err != nil {
		return data.Post{}, errors.New("error parsing publish date")
	}

	toUpdate.PublishedOn = parsedDate.UTC()

	if post.LabelId != nil {
		toUpdate.LabelID = sql.NullInt32{Int32: int32(*post.LabelId), Valid: true}
	} else {
		toUpdate.LabelID = sql.NullInt32{Valid: false}
	}

	return s.db.UpdatePost(ctx, toUpdate)
}

func (s *PostService) GetReaction(ctx context.Context, viewerId int32, postId int32) (*string, error) {
	result, err := s.db.GetReaction(ctx, data.GetReactionParams{ViewerID: viewerId, PostID: postId})

	if err != nil {
		return nil, err
	}

	if len(result) == 0 || !result[0].Valid {
		return nil, nil
	}

	return &result[0].String, nil
}

func (s *PostService) SaveInteraction(ctx context.Context, postId int32, viewerId int32, projectId int32, interactionType models.InteractionType, content *string) (*data.ChangelogInteraction, error) {
	alreadyInteracted := false

	if interactionType == models.InteractionTypeView {
		alreadyViewed, err := s.db.UserViewed(ctx, data.UserViewedParams{PostID: postId, ViewerID: viewerId})
		if err != nil {
			return nil, err
		}

		alreadyInteracted = alreadyViewed > 0
	} else if interactionType == models.InteractionTypeReaction {
		alreadyReacted, err := s.db.GetReaction(ctx, data.GetReactionParams{PostID: postId, ViewerID: viewerId, ProjectID: projectId})
		if err != nil {
			return nil, err
		}

		alreadyInteracted = len(alreadyReacted) > 0
	}

	insertParams := data.InsertInteractionParams{
		PostID:            postId,
		ViewerID:          viewerId,
		ProjectID:         projectId,
		Content:           sql.NullString{Valid: false},
		InteractionTypeID: int32(interactionType),
	}

	if content != nil {
		insertParams.Content = sql.NullString{String: *content, Valid: true}
	}

	if alreadyInteracted {
		updatedReaction, err := s.db.UpdateInteraction(ctx, data.UpdateInteractionParams{
			Content:           insertParams.Content,
			ViewerID:          viewerId,
			PostID:            postId,
			InteractionTypeID: int32(interactionType),
		})

		if err != nil {
			return nil, err
		}

		return &updatedReaction, err
	}

	insertedReaction, err := s.db.InsertInteraction(ctx, insertParams)
	if err != nil {
		return nil, err
	}

	return &insertedReaction, err
}
