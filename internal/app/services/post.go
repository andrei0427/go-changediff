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

func (s *PostService) GetPostCountForUser(ctx context.Context, userId uuid.UUID) (int64, error) {
	posts, err := s.db.GetPostCount(ctx, userId)
	return posts, err
}

func (s *PostService) GetPosts(ctx context.Context, userId uuid.UUID) ([]data.GetPostsRow, error) {
	posts, err := s.db.GetPosts(ctx, userId)
	return posts, err
}

func (s *PostService) GetPost(ctx context.Context, postId int32, userId uuid.UUID) (data.Post, error) {
	post, err := s.db.GetPost(ctx, data.GetPostParams{ID: postId, AuthorID: userId})
	return post, err
}

func (s *PostService) InsertPost(ctx context.Context, post models.PostModel, bannerUrl *string, userId uuid.UUID, projectId int32) (data.Post, error) {
	toInsert := data.InsertPostParams{
		Title:       post.Title,
		Body:        post.Content,
		AuthorID:    userId,
		ProjectID:   projectId,
		PublishedOn: time.Now(),
	}

	if bannerUrl != nil {
		toInsert.BannerImageUrl = sql.NullString{String: *bannerUrl}
	}

	if post.PublishedOn != nil {
		if parsedDate, err := time.Parse(time.DateOnly, *post.PublishedOn); err == nil {
			toInsert.PublishedOn = parsedDate
		}
	}

	return s.db.InsertPost(ctx, toInsert)
}

func (s *PostService) DeletePost(ctx context.Context, postId int32, userId uuid.UUID) (int32, error) {
	return s.db.DeletePost(ctx, data.DeletePostParams{ID: postId, AuthorID: userId})
}

func (s *PostService) UpdatePost(ctx context.Context, post models.PostModel, bannerUrl *string, userId uuid.UUID, projectId int32) (data.Post, error) {
	toUpdate := data.UpdatePostParams{
		Title:    post.Title,
		Body:     post.Content,
		AuthorID: userId,
	}

	if post.Id != nil {
		toUpdate.ID = int32(*post.Id)
	} else {
		return data.Post{}, errors.New("ID is required when updating")
	}

	existingPost, err := s.GetPost(ctx, toUpdate.ID, userId)
	if err != nil {
		return data.Post{}, errors.New("could not find post to update")
	}

	if bannerUrl != nil {
		toUpdate.BannerImageUrl = sql.NullString{String: *bannerUrl}
	} else {
		toUpdate.BannerImageUrl = existingPost.BannerImageUrl
	}

	if post.PublishedOn != nil {
		if parsedDate, err := time.Parse(time.DateOnly, *post.PublishedOn); err == nil {
			toUpdate.PublishedOn = parsedDate
		}
	} else {
		toUpdate.PublishedOn = existingPost.PublishedOn
	}

	return s.db.UpdatePost(ctx, toUpdate)
}
