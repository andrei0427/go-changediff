// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.20.0
// source: query.sql

package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

const deleteLabel = `-- name: DeleteLabel :one
DELETE FROM labels WHERE id = $1 AND project_id = $2 RETURNING id
`

type DeleteLabelParams struct {
	ID        int32
	ProjectID int32
}

func (q *Queries) DeleteLabel(ctx context.Context, arg DeleteLabelParams) (int32, error) {
	row := q.db.QueryRowContext(ctx, deleteLabel, arg.ID, arg.ProjectID)
	var id int32
	err := row.Scan(&id)
	return id, err
}

const deletePost = `-- name: DeletePost :one
DELETE FROM posts WHERE id = $1 AND project_id = $2 RETURNING id
`

type DeletePostParams struct {
	ID        int32
	ProjectID int32
}

func (q *Queries) DeletePost(ctx context.Context, arg DeletePostParams) (int32, error) {
	row := q.db.QueryRowContext(ctx, deletePost, arg.ID, arg.ProjectID)
	var id int32
	err := row.Scan(&id)
	return id, err
}

const getAuthorByUser = `-- name: GetAuthorByUser :many
SELECT id, first_name, last_name, picture_url, user_id, project_id, created_on, updated_on FROM authors WHERE user_id = $1 LIMIT 1
`

// AUTHOR --
func (q *Queries) GetAuthorByUser(ctx context.Context, userID uuid.UUID) ([]Author, error) {
	rows, err := q.db.QueryContext(ctx, getAuthorByUser, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Author
	for rows.Next() {
		var i Author
		if err := rows.Scan(
			&i.ID,
			&i.FirstName,
			&i.LastName,
			&i.PictureUrl,
			&i.UserID,
			&i.ProjectID,
			&i.CreatedOn,
			&i.UpdatedOn,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getLabels = `-- name: GetLabels :many
SELECT id, label, color, project_id, created_on, updated_on from labels WHERE project_id = $1 ORDER BY created_on
`

// LABELS --
func (q *Queries) GetLabels(ctx context.Context, projectID int32) ([]Label, error) {
	rows, err := q.db.QueryContext(ctx, getLabels, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Label
	for rows.Next() {
		var i Label
		if err := rows.Scan(
			&i.ID,
			&i.Label,
			&i.Color,
			&i.ProjectID,
			&i.CreatedOn,
			&i.UpdatedOn,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getPost = `-- name: GetPost :one
SELECT p.id, p.title, p.body, p.published_on, p.author_id, p.project_id, p.created_on, p.updated_on, p.label_id, l.label as Label FROM posts p LEFT JOIN labels l on p.label_id = l.id or p.label_id is null WHERE p.id = $1 AND p.project_id = $2
`

type GetPostParams struct {
	ID        int32
	ProjectID int32
}

type GetPostRow struct {
	ID          int32
	Title       string
	Body        string
	PublishedOn time.Time
	AuthorID    int32
	ProjectID   int32
	CreatedOn   time.Time
	UpdatedOn   sql.NullTime
	LabelID     sql.NullInt32
	Label       sql.NullString
}

func (q *Queries) GetPost(ctx context.Context, arg GetPostParams) (GetPostRow, error) {
	row := q.db.QueryRowContext(ctx, getPost, arg.ID, arg.ProjectID)
	var i GetPostRow
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.Body,
		&i.PublishedOn,
		&i.AuthorID,
		&i.ProjectID,
		&i.CreatedOn,
		&i.UpdatedOn,
		&i.LabelID,
		&i.Label,
	)
	return i, err
}

const getPostComments = `-- name: GetPostComments :many
SELECT c.comment, c.created_on, r.locale, ur.reaction 
	FROM posts p 
		JOIN post_comments c ON c.post_id = p.id 
		JOIN post_reactions r ON r.user_uuid = c.user_uuid AND r.post_id = p.id AND r.reaction IS NULL 
		LEFT JOIN post_reactions ur ON ur.user_uuid = c.user_uuid AND ur.post_id = p.id AND ur.reaction IS NOT NULL 
WHERE p.id = $1 AND p.project_id = $2
ORDER BY c.created_on DESC
`

type GetPostCommentsParams struct {
	ID        int32
	ProjectID int32
}

type GetPostCommentsRow struct {
	Comment   string
	CreatedOn time.Time
	Locale    string
	Reaction  sql.NullString
}

func (q *Queries) GetPostComments(ctx context.Context, arg GetPostCommentsParams) ([]GetPostCommentsRow, error) {
	rows, err := q.db.QueryContext(ctx, getPostComments, arg.ID, arg.ProjectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetPostCommentsRow
	for rows.Next() {
		var i GetPostCommentsRow
		if err := rows.Scan(
			&i.Comment,
			&i.CreatedOn,
			&i.Locale,
			&i.Reaction,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getPostCount = `-- name: GetPostCount :one
SELECT COUNT(id) total_posts FROM posts WHERE project_id = $1
`

// POSTS --
func (q *Queries) GetPostCount(ctx context.Context, projectID int32) (int64, error) {
	row := q.db.QueryRowContext(ctx, getPostCount, projectID)
	var total_posts int64
	err := row.Scan(&total_posts)
	return total_posts, err
}

const getPostReactions = `-- name: GetPostReactions :many
SELECT r.reaction, COUNT(r.*) FROM posts p JOIN post_reactions r ON r.post_id = p.id WHERE p.id = $1 AND p.project_id = $2 GROUP BY r.reaction ORDER BY r.reaction NULLS FIRST
`

type GetPostReactionsParams struct {
	ID        int32
	ProjectID int32
}

type GetPostReactionsRow struct {
	Reaction sql.NullString
	Count    int64
}

func (q *Queries) GetPostReactions(ctx context.Context, arg GetPostReactionsParams) ([]GetPostReactionsRow, error) {
	rows, err := q.db.QueryContext(ctx, getPostReactions, arg.ID, arg.ProjectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetPostReactionsRow
	for rows.Next() {
		var i GetPostReactionsRow
		if err := rows.Scan(&i.Reaction, &i.Count); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getPosts = `-- name: GetPosts :many
SELECT p.id, p.title, p.published_on, l.label, l.color, CASE WHEN p.published_on <= current_timestamp THEN 1 ELSE 0 END AS status, COUNT(r.id) as ViewCount FROM posts p left join labels l on p.label_id = l.id or p.label_id is null left join post_reactions r on p.id = r.post_id OR r.id is null WHERE p.project_id = $1 GROUP BY 1,2,3,4,5,6
`

type GetPostsRow struct {
	ID          int32
	Title       string
	PublishedOn time.Time
	Label       sql.NullString
	Color       sql.NullString
	Status      int32
	Viewcount   int64
}

func (q *Queries) GetPosts(ctx context.Context, projectID int32) ([]GetPostsRow, error) {
	rows, err := q.db.QueryContext(ctx, getPosts, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetPostsRow
	for rows.Next() {
		var i GetPostsRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.PublishedOn,
			&i.Label,
			&i.Color,
			&i.Status,
			&i.Viewcount,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getProject = `-- name: GetProject :many
SELECT id, name, description, accent_color, logo_url, app_key, user_id, created_on, updated_on FROM projects WHERE user_id = $1 LIMIT 1
`

// PROJECTS --
func (q *Queries) GetProject(ctx context.Context, userID uuid.UUID) ([]Project, error) {
	rows, err := q.db.QueryContext(ctx, getProject, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Project
	for rows.Next() {
		var i Project
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.AccentColor,
			&i.LogoUrl,
			&i.AppKey,
			&i.UserID,
			&i.CreatedOn,
			&i.UpdatedOn,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getProjectByKey = `-- name: GetProjectByKey :one
SELECT id, name, description, accent_color, logo_url, app_key, user_id, created_on, updated_on FROM projects where app_key = $1 LIMIT 1
`

func (q *Queries) GetProjectByKey(ctx context.Context, appKey string) (Project, error) {
	row := q.db.QueryRowContext(ctx, getProjectByKey, appKey)
	var i Project
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.AccentColor,
		&i.LogoUrl,
		&i.AppKey,
		&i.UserID,
		&i.CreatedOn,
		&i.UpdatedOn,
	)
	return i, err
}

const getPublishedPagedPosts = `-- name: GetPublishedPagedPosts :many
SELECT post.id, post.title, post.body, post.published_on, post.author_id, post.project_id, post.created_on, post.updated_on, post.label_id, l.label, l.color, a.first_name, a.last_name, a.picture_url, r.reaction, CASE WHEN v.id IS NULL THEN 0 ELSE 1 END as Viewed
  FROM posts post 
    join projects proj on post.project_id = proj.id 
	join authors a on a.id = post.author_id 
	left join labels l on post.label_id = l.id or post.label_id is null 
	left join post_reactions r on (r.post_id = post.id and r.user_uuid = $4 and r.reaction is not null) or r.id is null 
	left join post_reactions v on (v.post_id = post.id and v.user_uuid = $4 and v.reaction is null) or v.id is null 
WHERE proj.app_key = $1 AND post.published_on <= CURRENT_TIMESTAMP 
ORDER BY post.published_on DESC 
LIMIT $2 
OFFSET $3
`

type GetPublishedPagedPostsParams struct {
	AppKey   string
	Limit    int32
	Offset   int32
	UserUuid uuid.UUID
}

type GetPublishedPagedPostsRow struct {
	ID          int32
	Title       string
	Body        string
	PublishedOn time.Time
	AuthorID    int32
	ProjectID   int32
	CreatedOn   time.Time
	UpdatedOn   sql.NullTime
	LabelID     sql.NullInt32
	Label       sql.NullString
	Color       sql.NullString
	FirstName   string
	LastName    string
	PictureUrl  sql.NullString
	Reaction    sql.NullString
	Viewed      int32
}

func (q *Queries) GetPublishedPagedPosts(ctx context.Context, arg GetPublishedPagedPostsParams) ([]GetPublishedPagedPostsRow, error) {
	rows, err := q.db.QueryContext(ctx, getPublishedPagedPosts,
		arg.AppKey,
		arg.Limit,
		arg.Offset,
		arg.UserUuid,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetPublishedPagedPostsRow
	for rows.Next() {
		var i GetPublishedPagedPostsRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Body,
			&i.PublishedOn,
			&i.AuthorID,
			&i.ProjectID,
			&i.CreatedOn,
			&i.UpdatedOn,
			&i.LabelID,
			&i.Label,
			&i.Color,
			&i.FirstName,
			&i.LastName,
			&i.PictureUrl,
			&i.Reaction,
			&i.Viewed,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getReaction = `-- name: GetReaction :many
SELECT reaction FROM post_reactions WHERE user_uuid = $1 AND post_id = $2 AND reaction IS NOT NULL
`

type GetReactionParams struct {
	UserUuid uuid.UUID
	PostID   int32
}

func (q *Queries) GetReaction(ctx context.Context, arg GetReactionParams) ([]sql.NullString, error) {
	rows, err := q.db.QueryContext(ctx, getReaction, arg.UserUuid, arg.PostID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []sql.NullString
	for rows.Next() {
		var reaction sql.NullString
		if err := rows.Scan(&reaction); err != nil {
			return nil, err
		}
		items = append(items, reaction)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const insertAuthor = `-- name: InsertAuthor :one
INSERT INTO authors (first_name, last_name, picture_url, user_id, project_id) VALUES ($1, $2, $3, $4, $5) RETURNING id, first_name, last_name, picture_url, user_id, project_id, created_on, updated_on
`

type InsertAuthorParams struct {
	FirstName  string
	LastName   string
	PictureUrl sql.NullString
	UserID     uuid.UUID
	ProjectID  int32
}

func (q *Queries) InsertAuthor(ctx context.Context, arg InsertAuthorParams) (Author, error) {
	row := q.db.QueryRowContext(ctx, insertAuthor,
		arg.FirstName,
		arg.LastName,
		arg.PictureUrl,
		arg.UserID,
		arg.ProjectID,
	)
	var i Author
	err := row.Scan(
		&i.ID,
		&i.FirstName,
		&i.LastName,
		&i.PictureUrl,
		&i.UserID,
		&i.ProjectID,
		&i.CreatedOn,
		&i.UpdatedOn,
	)
	return i, err
}

const insertComment = `-- name: InsertComment :one
INSERT INTO post_comments (user_uuid, comment, post_id) VALUES ($1, $2, $3) RETURNING id, user_uuid, comment, post_id, created_on
`

type InsertCommentParams struct {
	UserUuid uuid.UUID
	Comment  string
	PostID   int32
}

func (q *Queries) InsertComment(ctx context.Context, arg InsertCommentParams) (PostComment, error) {
	row := q.db.QueryRowContext(ctx, insertComment, arg.UserUuid, arg.Comment, arg.PostID)
	var i PostComment
	err := row.Scan(
		&i.ID,
		&i.UserUuid,
		&i.Comment,
		&i.PostID,
		&i.CreatedOn,
	)
	return i, err
}

const insertLabel = `-- name: InsertLabel :one
INSERT INTO labels (label, color, project_id) VALUES ($1, $2, $3) RETURNING id, label, color, project_id, created_on, updated_on
`

type InsertLabelParams struct {
	Label     string
	Color     string
	ProjectID int32
}

func (q *Queries) InsertLabel(ctx context.Context, arg InsertLabelParams) (Label, error) {
	row := q.db.QueryRowContext(ctx, insertLabel, arg.Label, arg.Color, arg.ProjectID)
	var i Label
	err := row.Scan(
		&i.ID,
		&i.Label,
		&i.Color,
		&i.ProjectID,
		&i.CreatedOn,
		&i.UpdatedOn,
	)
	return i, err
}

const insertPost = `-- name: InsertPost :one
INSERT INTO posts (title, body, published_on, label_id, author_id, project_id) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, title, body, published_on, author_id, project_id, created_on, updated_on, label_id
`

type InsertPostParams struct {
	Title       string
	Body        string
	PublishedOn time.Time
	LabelID     sql.NullInt32
	AuthorID    int32
	ProjectID   int32
}

func (q *Queries) InsertPost(ctx context.Context, arg InsertPostParams) (Post, error) {
	row := q.db.QueryRowContext(ctx, insertPost,
		arg.Title,
		arg.Body,
		arg.PublishedOn,
		arg.LabelID,
		arg.AuthorID,
		arg.ProjectID,
	)
	var i Post
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.Body,
		&i.PublishedOn,
		&i.AuthorID,
		&i.ProjectID,
		&i.CreatedOn,
		&i.UpdatedOn,
		&i.LabelID,
	)
	return i, err
}

const insertProject = `-- name: InsertProject :one
INSERT INTO projects (name, description, accent_color, logo_url, app_key, user_id) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, name, description, accent_color, logo_url, app_key, user_id, created_on, updated_on
`

type InsertProjectParams struct {
	Name        string
	Description string
	AccentColor string
	LogoUrl     sql.NullString
	AppKey      string
	UserID      uuid.UUID
}

func (q *Queries) InsertProject(ctx context.Context, arg InsertProjectParams) (Project, error) {
	row := q.db.QueryRowContext(ctx, insertProject,
		arg.Name,
		arg.Description,
		arg.AccentColor,
		arg.LogoUrl,
		arg.AppKey,
		arg.UserID,
	)
	var i Project
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.AccentColor,
		&i.LogoUrl,
		&i.AppKey,
		&i.UserID,
		&i.CreatedOn,
		&i.UpdatedOn,
	)
	return i, err
}

const insertReaction = `-- name: InsertReaction :one
INSERT INTO post_reactions (user_uuid, ip_addr, user_agent, locale, reaction, post_id) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, user_uuid, ip_addr, user_agent, locale, reaction, post_id, created_on
`

type InsertReactionParams struct {
	UserUuid  uuid.UUID
	IpAddr    string
	UserAgent string
	Locale    string
	Reaction  sql.NullString
	PostID    int32
}

func (q *Queries) InsertReaction(ctx context.Context, arg InsertReactionParams) (PostReaction, error) {
	row := q.db.QueryRowContext(ctx, insertReaction,
		arg.UserUuid,
		arg.IpAddr,
		arg.UserAgent,
		arg.Locale,
		arg.Reaction,
		arg.PostID,
	)
	var i PostReaction
	err := row.Scan(
		&i.ID,
		&i.UserUuid,
		&i.IpAddr,
		&i.UserAgent,
		&i.Locale,
		&i.Reaction,
		&i.PostID,
		&i.CreatedOn,
	)
	return i, err
}

const unsetLabels = `-- name: UnsetLabels :many
UPDATE posts SET label_id = NULL WHERE id = $1 AND project_id = $2 RETURNING id
`

type UnsetLabelsParams struct {
	ID        int32
	ProjectID int32
}

func (q *Queries) UnsetLabels(ctx context.Context, arg UnsetLabelsParams) ([]int32, error) {
	rows, err := q.db.QueryContext(ctx, unsetLabels, arg.ID, arg.ProjectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []int32
	for rows.Next() {
		var id int32
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		items = append(items, id)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateLabel = `-- name: UpdateLabel :one
UPDATE labels SET label = $1, color = $2 WHERE id = $3 AND project_id = $4 RETURNING id, label, color, project_id, created_on, updated_on
`

type UpdateLabelParams struct {
	Label     string
	Color     string
	ID        int32
	ProjectID int32
}

func (q *Queries) UpdateLabel(ctx context.Context, arg UpdateLabelParams) (Label, error) {
	row := q.db.QueryRowContext(ctx, updateLabel,
		arg.Label,
		arg.Color,
		arg.ID,
		arg.ProjectID,
	)
	var i Label
	err := row.Scan(
		&i.ID,
		&i.Label,
		&i.Color,
		&i.ProjectID,
		&i.CreatedOn,
		&i.UpdatedOn,
	)
	return i, err
}

const updatePost = `-- name: UpdatePost :one
UPDATE posts SET title = $1, body = $2, published_on = $3, label_id = $4, updated_on = CURRENT_TIMESTAMP WHERE id = $5 AND project_id = $6 RETURNING id, title, body, published_on, author_id, project_id, created_on, updated_on, label_id
`

type UpdatePostParams struct {
	Title       string
	Body        string
	PublishedOn time.Time
	LabelID     sql.NullInt32
	ID          int32
	ProjectID   int32
}

func (q *Queries) UpdatePost(ctx context.Context, arg UpdatePostParams) (Post, error) {
	row := q.db.QueryRowContext(ctx, updatePost,
		arg.Title,
		arg.Body,
		arg.PublishedOn,
		arg.LabelID,
		arg.ID,
		arg.ProjectID,
	)
	var i Post
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.Body,
		&i.PublishedOn,
		&i.AuthorID,
		&i.ProjectID,
		&i.CreatedOn,
		&i.UpdatedOn,
		&i.LabelID,
	)
	return i, err
}

const updateProject = `-- name: UpdateProject :one
UPDATE projects SET name = $1, description = $2, accent_color = $3, logo_url = $4, updated_on = CURRENT_TIMESTAMP WHERE id = $5 AND user_id = $6 RETURNING id, name, description, accent_color, logo_url, app_key, user_id, created_on, updated_on
`

type UpdateProjectParams struct {
	Name        string
	Description string
	AccentColor string
	LogoUrl     sql.NullString
	ID          int32
	UserID      uuid.UUID
}

func (q *Queries) UpdateProject(ctx context.Context, arg UpdateProjectParams) (Project, error) {
	row := q.db.QueryRowContext(ctx, updateProject,
		arg.Name,
		arg.Description,
		arg.AccentColor,
		arg.LogoUrl,
		arg.ID,
		arg.UserID,
	)
	var i Project
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.AccentColor,
		&i.LogoUrl,
		&i.AppKey,
		&i.UserID,
		&i.CreatedOn,
		&i.UpdatedOn,
	)
	return i, err
}

const updateReaction = `-- name: UpdateReaction :one
UPDATE post_reactions SET reaction = $1 WHERE user_uuid = $2 AND post_id = $3 AND reaction IS NOT NULL RETURNING id, user_uuid, ip_addr, user_agent, locale, reaction, post_id, created_on
`

type UpdateReactionParams struct {
	Reaction sql.NullString
	UserUuid uuid.UUID
	PostID   int32
}

func (q *Queries) UpdateReaction(ctx context.Context, arg UpdateReactionParams) (PostReaction, error) {
	row := q.db.QueryRowContext(ctx, updateReaction, arg.Reaction, arg.UserUuid, arg.PostID)
	var i PostReaction
	err := row.Scan(
		&i.ID,
		&i.UserUuid,
		&i.IpAddr,
		&i.UserAgent,
		&i.Locale,
		&i.Reaction,
		&i.PostID,
		&i.CreatedOn,
	)
	return i, err
}

const userViewed = `-- name: UserViewed :one
SELECT COUNT(id) FROM post_reactions WHERE user_uuid = $1 AND post_id = $2 AND reaction IS NULL
`

type UserViewedParams struct {
	UserUuid uuid.UUID
	PostID   int32
}

func (q *Queries) UserViewed(ctx context.Context, arg UserViewedParams) (int64, error) {
	row := q.db.QueryRowContext(ctx, userViewed, arg.UserUuid, arg.PostID)
	var count int64
	err := row.Scan(&count)
	return count, err
}
