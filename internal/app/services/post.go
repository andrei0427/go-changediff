package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/andrei0427/go-changediff/internal/app/models"
	"github.com/andrei0427/go-changediff/internal/data"
	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

type PostService struct {
	db *data.Queries
}

func NewPostService(db *data.Queries) *PostService {
	return &PostService{db: db}
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
	post, err := s.db.GetPost(ctx, data.GetPostParams{ID: postId, ProjectID: projectId})
	return post, err
}

func (s *PostService) GetPostReactions(ctx context.Context, postId *int32, projectId int32, userUuid *string, userId *string) ([]data.GetPostReactionsRow, error) {
	params := data.GetPostReactionsParams{
		ProjectID: projectId,
	}

	if postId != nil {
		params.ID = *postId
	}

	if userId != nil && len(*userId) > 0 {
		params.UserID = sql.NullString{String: *userId, Valid: true}
	}

	if userUuid != nil && len(*userUuid) > 0 {
		params.Column3 = *userUuid
	}

	reactions, err := s.db.GetPostReactions(ctx, params)
	return reactions, err
}

func (s *PostService) GetPostComments(ctx context.Context, postId *int32, projectId int32, userUuid *string, userId *string) ([]data.GetPostCommentsRow, error) {
	params := data.GetPostCommentsParams{ProjectID: projectId}

	if postId != nil {
		params.ID = *postId
	}

	if userId != nil && len(*userId) > 0 {
		params.UserID = sql.NullString{String: *userId, Valid: true}
	}

	if userUuid != nil && len(*userUuid) > 0 {
		params.Column3 = *userUuid
	}

	comments, err := s.db.GetPostComments(ctx, params)
	return comments, err
}

func (s *PostService) GetAnalytics(ctx context.Context, projectId int32, userUuid *string, userId *string) ([]data.AnalyticsUsersRow, error) {
	params := data.AnalyticsUsersParams{
		ProjectID: projectId,
	}

	if userUuid != nil && len(*userUuid) > 0 {
		params.Column3 = *userUuid
	}

	if userId != nil && len(*userId) > 0 {
		params.UserID = sql.NullString{String: *userId, Valid: true}
	}

	return s.db.AnalyticsUsers(ctx, params)
}

func (s *PostService) GetPublishedPagedPosts(ctx context.Context, projectKey string, pageNo int32, search string, userId uuid.UUID) ([]data.GetPublishedPagedPostsRow, error) {
	var offset int32 = 0
	if pageNo > 1 {
		offset = pageNo - 1*5
	}

	searchStr := search
	if len(searchStr) > 0 {
		searchStr = "%" + strings.ToLower(search) + "%"
	}

	posts, err := s.db.GetPublishedPagedPosts(ctx, data.GetPublishedPagedPostsParams{AppKey: projectKey, Limit: 5, Offset: offset, UserUuid: userId, Column5: searchStr})

	return posts, err
}

func (s *PostService) InsertPostComment(ctx context.Context, userId uuid.UUID, comment string, postId int32) (data.PostComment, error) {
	return s.db.InsertComment(ctx, data.InsertCommentParams{UserUuid: userId, Comment: comment, PostID: postId})
}

func (s *PostService) InsertPost(ctx context.Context, post models.PostModel, authorId int32, projectId int32, userLocation *time.Location) (data.Post, error) {
	toInsert := data.InsertPostParams{
		Title:       post.Title,
		Body:        post.Content,
		AuthorID:    authorId,
		ProjectID:   projectId,
		IsPublished: sql.NullBool{Bool: post.IsPublished, Valid: true},
	}

	parsedDate, err := time.ParseInLocation("2006-01-02T15:04", post.PublishedOn, userLocation)
	if err != nil {
		return data.Post{}, errors.New("error parsing publish date")
	}

	toInsert.PublishedOn = parsedDate.UTC()

	if post.LabelId != nil {
		toInsert.LabelID = sql.NullInt32{Int32: int32(*post.LabelId), Valid: true}
	}

	return s.db.InsertPost(ctx, toInsert)
}

func (s *PostService) DeletePost(ctx context.Context, postId int32, projectId int32) (int32, error) {
	return s.db.DeletePost(ctx, data.DeletePostParams{ID: postId, ProjectID: projectId})
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

func (s *PostService) GetReaction(ctx context.Context, userId uuid.UUID, postId int32) (*string, error) {
	result, err := s.db.GetReaction(ctx, data.GetReactionParams{UserUuid: userId, PostID: postId})

	if err != nil {
		return nil, err
	}

	if len(result) == 0 || !result[0].Valid {
		return nil, nil
	}

	return &result[0].String, nil
}

func (s *PostService) SaveReaction(ctx context.Context, params data.InsertReactionParams, userInfo *models.UserInfo) (*data.PostReaction, error) {
	// Saving a 'view' reaction - only insert if one doesnt yet exist
	if userInfo.ID != nil {
		params.UserID = sql.NullString{String: string(*userInfo.ID), Valid: true}
	}

	if userInfo.Email != nil {
		params.UserEmail = sql.NullString{String: string(*userInfo.Email), Valid: true}
	}

	if userInfo.Info != nil {
		if marshalled, err := json.Marshal(userInfo.Info); err == nil {
			params.UserData = pqtype.NullRawMessage{RawMessage: marshalled, Valid: true}
		}
	}

	if userInfo.Name != nil {
		params.UserName = sql.NullString{String: string(*userInfo.Name), Valid: true}
	}

	if userInfo.Role != nil {
		params.UserRole = sql.NullString{String: string(*userInfo.Role), Valid: true}
	}

	if !params.Reaction.Valid {
		alreadyViewed, err := s.db.UserViewed(ctx, data.UserViewedParams{UserUuid: params.UserUuid, PostID: params.PostID})
		if err != nil {
			return nil, err
		}

		if alreadyViewed == 0 {
			savedReaction, err := s.db.InsertReaction(ctx, params)
			return &savedReaction, err
		}
	} else {
		existingReaction, err := s.GetReaction(ctx, params.UserUuid, params.PostID)
		if err != nil {
			return nil, err
		}

		if existingReaction == nil {
			savedReaction, err := s.db.InsertReaction(ctx, params)
			if err != nil {
				return nil, err
			}

			return &savedReaction, err
		} else {
			updatedReaction, err := s.db.UpdateReaction(ctx, data.UpdateReactionParams{UserUuid: params.UserUuid,
				PostID:   params.PostID,
				Reaction: params.Reaction,
			})
			if err != nil {
				return nil, err
			}

			return &updatedReaction, err
		}
	}

	return nil, nil
}
