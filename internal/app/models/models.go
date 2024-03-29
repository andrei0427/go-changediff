package models

import (
	"encoding/json"
	"time"

	"github.com/andrei0427/go-changediff/internal/data"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

type InteractionType int32

const (
	InteractionTypeView InteractionType = iota + 1
	InteractionTypeReaction
	InteractionTypeComment
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

type RoadmapPostModel struct {
	ID        *int32 `form:"id"`
	StatusID  int32  `form:"status_id"`
	BoardID   *int32 `form:"board_id"`
	Title     string `form:"title"`
	Content   string `form:"content"`
	IsPrivate bool   `form:"is_private"`
	DueDate   string `form:"due_date"`
	IsIdea    bool   `form:"is_idea"`

	Reactions []RoadmapPostReactionActivityModel
}

type ActivityType int

const (
	ActivityTypeCreation ActivityType = iota + 1
	ActivityTypeStatusUpdate
	ActivityTypeComment
)

type RoadmapPostActivityModel struct {
	ID            int32
	CreatedOn     time.Time
	Who           string
	WhoPictureUrl string
	Type          ActivityType

	CommentActivity      *RoadmapPostCommentModel
	StatusUpdateActivity *RoadmapPostStatusActivityModel
}

type RoadmapPostCommentFormModel struct {
	Comment string `form:"comment"`
}

type RoadmapPostCommentModel struct {
	Comment         *string
	IsPinned        bool
	IsDeleted       bool
	ReplyCount      int64
	ParentCommentID *int32
	Reactions       []RoadmapPostReactionActivityModel
}

type RoadmapPostStatusActivityModel struct {
	FromStatus *RoadmapStatusModel
	ToStatus   *RoadmapStatusModel
}

type RoadmapPostReactionActivityModel struct {
	Who             string
	Emoji           string
	Count           int64
	Reacted         bool
	ParentCommentID *int32
}

type RoadmapPostVoteCount struct {
	Count int64
	Voted bool
}

type RoadmapBoardStatusWithPosts struct {
	Status data.GetStatusesRow
	Posts  []data.GetPostsForBoardRow
}

type RoadmapPostPinModel struct {
	ID       int32
	IsPinned bool
}

type WidgetRoadmapPostModel struct {
	ID           int32
	Title        string
	Board        string
	CreatedOn    time.Time
	DueDate      time.Time
	HasDueDate   bool
	IsPinned     bool
	IsIdea       bool
	CommentCount int64
}
type WidgetRoadmapData struct {
	ID     int32
	Status string
	Color  string
	Posts  []WidgetRoadmapPostModel
}
