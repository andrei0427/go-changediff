// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.23.0

package data

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
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
	IsPublished sql.NullBool
	ExpiresOn   sql.NullTime
}

type PostComment struct {
	ID        int32
	UserUuid  uuid.UUID
	Comment   string
	PostID    int32
	CreatedOn time.Time
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
	UserData  pqtype.NullRawMessage
	UserID    sql.NullString
	UserName  sql.NullString
	UserEmail sql.NullString
	UserRole  sql.NullString
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

type RoadmapBoard struct {
	ID          int32
	Name        string
	IsPrivate   bool
	Description string
	CreatedOn   time.Time
	ProjectID   int32
}

type RoadmapCategory struct {
	ID        int32
	Name      string
	Emoji     string
	IsPrivate bool
	ProjectID int32
	CreatorID int32
	CreatedOn time.Time
}

type RoadmapPost struct {
	ID        int32
	Title     string
	Body      string
	DueDate   sql.NullTime
	BoardID   sql.NullInt32
	ProjectID int32
	StatusID  sql.NullInt32
	CreatedOn time.Time
	IsPrivate bool
	AuthorID  sql.NullInt32
	UserUuid  uuid.NullUUID
	IsIdea    bool
}

type RoadmapPostCategory struct {
	ID                int32
	RoadmapPostID     int32
	RoadmapCategoryID int32
	ProjectID         int32
	CreatedOn         time.Time
}

type RoadmapStatus struct {
	ID          int32
	Status      string
	Color       string
	Description string
	CreatedOn   time.Time
	ProjectID   int32
	IsPrivate   bool
	SortOrder   int32
}

type Subscription struct {
	ID                    int32
	SubscriptionStartDate time.Time
	IsAnnual              bool
	Price                 string
	Tier                  int32
	SessionID             sql.NullString
	Success               bool
	Stopped               bool
	Message               sql.NullString
	SubscriberID          int32
	ProjectID             int32
	CreatedOn             time.Time
}
