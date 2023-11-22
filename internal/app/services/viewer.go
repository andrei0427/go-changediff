package services

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/andrei0427/go-changediff/internal/app/models"
	"github.com/andrei0427/go-changediff/internal/data"
	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

type ViewerService struct {
	db *data.Queries
}

func NewViewerService(db *data.Queries) *ViewerService {
	return &ViewerService{db: db}
}

func (s *ViewerService) GetViewer(ctx context.Context, userUuid uuid.UUID, userId *string) (*data.Viewer, error) {
	params := data.GetViewerParams{
		UserUuid: userUuid,
		UserID:   sql.NullString{Valid: false},
	}

	if userId != nil {
		params.UserID = sql.NullString{String: *userId, Valid: true}
	}

	viewers, err := s.db.GetViewer(ctx, params)

	if len(viewers) == 0 {
		return nil, err
	}

	return &viewers[0], err
}

func (s *ViewerService) SaveViewer(ctx context.Context,
	userUuid uuid.UUID,
	ipAddr string,
	userAgent string,
	locale string,
	userInfo *models.UserInfo,
	projectId int32,
) (data.Viewer, error) {

	insertParams := data.InsertViewerParams{
		UserUuid:  userUuid,
		IpAddr:    ipAddr,
		UserAgent: userAgent,
		Locale:    locale,
		ProjectID: projectId,
	}

	if userInfo != nil {
		if userInfo.ID != nil {
			insertParams.UserID = sql.NullString{String: string(*userInfo.ID), Valid: true}
		}

		if userInfo.Email != nil {
			insertParams.UserEmail = sql.NullString{String: string(*userInfo.Email), Valid: true}
		}

		if userInfo.Info != nil {
			if marshalled, err := json.Marshal(userInfo.Info); err == nil {
				insertParams.UserData = pqtype.NullRawMessage{RawMessage: marshalled, Valid: true}
			}
		}

		if userInfo.Name != nil {
			insertParams.UserName = sql.NullString{String: string(*userInfo.Name), Valid: true}
		}

		if userInfo.Role != nil {
			insertParams.UserRole = sql.NullString{String: string(*userInfo.Role), Valid: true}
		}
	}

	var userId *string
	if userInfo != nil && userInfo.ID != nil {
		strUserID := string(*userInfo.ID)
		userId = &strUserID
	}

	existingViewer, _ := s.GetViewer(ctx, userUuid, userId)
	if existingViewer != nil {
		updateParams := data.UpdateViewerParams{
			ID:        existingViewer.ID,
			UserUuid:  insertParams.UserUuid,
			IpAddr:    insertParams.IpAddr,
			UserAgent: insertParams.UserAgent,
			Locale:    insertParams.Locale,
			UserData:  insertParams.UserData,
			UserID:    insertParams.UserID,
			UserName:  insertParams.UserName,
			UserEmail: insertParams.UserEmail,
			UserRole:  insertParams.UserRole,
		}

		return s.db.UpdateViewer(ctx, updateParams)
	}

	return s.db.InsertViewer(ctx, insertParams)
}
