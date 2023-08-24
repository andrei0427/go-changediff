package services

import (
	"context"
	"database/sql"

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
	posts, err := s.db.GetPosts(ctx, userId)

	return posts, err
}

func (s *PostService) InsertPost(ctx context.Context, post models.PostModel, bannerUrl *string, userId uuid.UUID, projectId int32) (data.Post, error) {
	toInsert := data.InsertPostParams{
		Title:       post.Title,
		Body:        post.Content,
		PublishedOn: sql.NullTime{},
		AuthorID:    userId,
		ProjectID:   projectId,
	}

	if bannerUrl != nil {
		toInsert.BannerImageUrl = sql.NullString{String: *bannerUrl}
	}

	return s.db.InsertPost(ctx, toInsert)
}
