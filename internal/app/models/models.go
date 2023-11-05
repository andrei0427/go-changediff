package models

import (
	"encoding/json"

	"github.com/andrei0427/go-changediff/internal/data"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

type UserMetadata struct {
	AvatarUrl string `json:"avatar_url"`
	FullName  string `json:"full_name"`
}

type AuthTokenClaim struct {
	jwt.StandardClaims

	Email    string       `json:"email"`
	Metadata UserMetadata `json:"user_metadata"`
}

type SessionUser struct {
	Id       uuid.UUID
	Email    string
	Metadata UserMetadata
	Timezone string
	Project  *data.Project
	Author   *data.GetAuthorByUserRow
}

type UserInfo struct {
	ID    *json.RawMessage       `json:"id"`
	Name  *json.RawMessage       `json:"name"`
	Email *json.RawMessage       `json:"email"`
	Role  *json.RawMessage       `json:"role"`
	Info  map[string]interface{} `json:"info"`
}

type GeneralSettingsModel struct {
	Name        string `form:"name"`
	Description string `form:"description"`
	AccentColor string `form:"accent_color"`
	FirstName   string `form:"first_name"`
	LastName    string `form:"last_name"`
}

type ProjectModel struct {
	ID          *int32  `form:"id"`
	Name        string  `form:"name"`
	Description string  `form:"description"`
	AccentColor string  `form:"accent_color"`
	AppKey      *string `form:"appkey"`
}

type LabelModel struct {
	ID    *int32 `form:"id"`
	Label string `form:"label"`
	Color string `form:"color"`
}

type PostModel struct {
	ID          *int32 `form:"id"`
	Title       string `form:"title"`
	Content     string `form:"content"`
	PublishedOn string `form:"published_on"`
	ExpiresOn   string `form:"expires_on"`
	IsPublished bool   `form:"is_published"`
	LabelId     *int   `form:"label_id"`
	First       *bool  `form:"first"`
}

type ChangelogComment struct {
	Comment string `form:"comment"`
}

type Search struct {
	Search string `form:"search"`
}

type RoadmapBoardModel struct {
	ID          *int32 `form:"id"`
	Name        string `form:"name"`
	IsPrivate   bool   `form:"is_private"`
	Description string `form:"description"`
}

type RoadmapStatusModel struct {
	ID          *int32 `form:"id"`
	Status      string `form:"status"`
	Color       string `form:"color"`
	Description string `form:"description"`
}
