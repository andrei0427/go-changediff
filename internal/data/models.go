// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.20.0

package data

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Author struct {
	ID         int32
	FirstName  string
	LastName   string
	PictureUrl sql.NullString
	UserID     uuid.UUID
	ProjectID  int32
	CreatedOn  time.Time
	UpdatedOn  sql.NullTime
}

type Label struct {
	ID        int32
	Label     string
	Color     string
	ProjectID int32
	CreatedOn time.Time
	UpdatedOn sql.NullTime
}

type Post struct {
	ID          int32
	Title       string
	Body        string
	PublishedOn time.Time
	AuthorID    int32
	ProjectID   int32
	CreatedOn   time.Time
	UpdatedOn   sql.NullTime
	LabelID     sql.NullInt32
}

type PostReaction struct {
	ID        int32
	UserUuid  uuid.UUID
	IpAddr    string
	UserAgent string
	Locale    string
	Reaction  sql.NullString
	PostID    int32
	CreatedOn time.Time
}

type Project struct {
	ID          int32
	Name        string
	Description string
	AccentColor string
	LogoUrl     sql.NullString
	AppKey      string
	UserID      uuid.UUID
	CreatedOn   time.Time
	UpdatedOn   sql.NullTime
}
