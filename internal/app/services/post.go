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

func (s *PostService) GetPublishedPagedPosts(ctx context.Context, projectKey string, pageNo int32, userId uuid.UUID) ([]data.GetPublishedPagedPostsRow, error) {
	var offset int32 = 0
	if pageNo > 1 {
		offset = pageNo - 1*5
	}

	posts, err := s.db.GetPublishedPagedPosts(ctx, data.GetPublishedPagedPostsParams{AppKey: projectKey, Limit: 5, Offset: offset, UserUuid: userId})

	return posts, err
}

func (s *PostService) InsertPost(ctx context.Context, post models.PostModel, authorId int32, projectId int32, userLocation *time.Location) (data.Post, error) {
	toInsert := data.InsertPostParams{
		Title:     post.Title,
		Body:      post.Content,
		AuthorID:  authorId,
		ProjectID: projectId,
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
		Title:     post.Title,
		Body:      post.Content,
		ProjectID: projectId,
	}

	if post.Id != nil {
		toUpdate.ID = int32(*post.Id)
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
