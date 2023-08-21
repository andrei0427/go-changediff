package services

import (
	"context"

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
