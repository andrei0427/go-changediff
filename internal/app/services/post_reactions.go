package services

import (
	"context"

	"github.com/andrei0427/go-changediff/internal/data"
	"github.com/google/uuid"
)

type PostReactionsService struct {
	db *data.Queries
}

func NewPostReactionsService(db *data.Queries) *PostReactionsService {
	return &PostReactionsService{db: db}
}

func (s *PostReactionsService) GetReaction(ctx context.Context, userId uuid.UUID, postId int32) (*string, error) {
	result, err := s.db.GetReaction(ctx, data.GetReactionParams{UserUuid: userId, PostID: postId})

	if err != nil {
		return nil, err
	}

	if len(result) == 0 || !result[0].Valid {
		return nil, nil
	}

	return &result[0].String, nil
}

func (s *PostReactionsService) SaveReaction(ctx context.Context, params data.InsertReactionParams) (*data.PostReaction, error) {
	// Saving a 'view' reaction - only insert if one doesnt yet exist
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
			updatedReaction, err := s.db.UpdateReaction(ctx, data.UpdateReactionParams{UserUuid: params.UserUuid, PostID: params.PostID, Reaction: params.Reaction})
			if err != nil {
				return nil, err
			}

			return &updatedReaction, err
		}
	}

	return nil, nil
}
