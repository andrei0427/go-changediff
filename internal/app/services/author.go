package services

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/andrei0427/go-changediff/internal/app/models"
	"github.com/andrei0427/go-changediff/internal/data"
	"github.com/google/uuid"
)

type AuthorService struct {
	db *data.Queries
}

func NewAuthorService(db *data.Queries) *AuthorService {
	return &AuthorService{db: db}
}

func (s *AuthorService) GetAuthorByUser(ctx context.Context, userId uuid.UUID) (*data.GetAuthorByUserRow, error) {
	authors, err := s.db.GetAuthorByUser(ctx, userId)

	if len(authors) == 0 {
		return nil, err
	}

	return &authors[0], err
}

func (s *AuthorService) UpdateAuthorForUser(ctx context.Context, userId uuid.UUID, projectId int32, user models.GeneralSettingsModel, imageUrl *string) (data.Author, error) {
	model := data.UpdateAuthorParams{
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		UserID:     userId,
		ProjectID:  projectId,
		PictureUrl: sql.NullString{},
	}

	if imageUrl != nil {
		model.PictureUrl = sql.NullString{String: *imageUrl, Valid: true}
	}

	return s.db.UpdateAuthor(ctx, model)

}

func (s *AuthorService) InsertAuthorForUser(ctx context.Context, user models.SessionUser) (*data.Author, error) {
	if user.Project == nil {
		return nil, errors.New("a project is required")
	}

	toInsert := data.InsertAuthorParams{
		PictureUrl: sql.NullString{String: user.Metadata.AvatarUrl, Valid: len(user.Metadata.AvatarUrl) > 0},
		ProjectID:  user.Project.ID,
		UserID:     user.Id,
	}

	nameParts := strings.Split(user.Metadata.FullName, " ")

	if len(nameParts) > 0 {
		toInsert.FirstName = nameParts[0]
	}

	if len(nameParts) > 1 {
		toInsert.LastName = strings.Join(nameParts[1:], " ")
	}

	insertedAuthor, err := s.db.InsertAuthor(ctx, toInsert)
	return &insertedAuthor, err
}
