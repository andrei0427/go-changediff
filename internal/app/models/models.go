package models

import (
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
	Author   *data.Author
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
	Id          *int64 `form:"id"`
	Title       string `form:"title"`
	Content     string `form:"content"`
	PublishedOn string `form:"published_on"`
	LabelId     *int   `form:"label_id"`
	First       *bool  `form:"first"`
}

type ChangelogComment struct {
	Comment string `form:"comment"`
}
