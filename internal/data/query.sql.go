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

const deletePost = `-- name: DeletePost :one
DELETE FROM posts WHERE id = $1 AND author_id = $2 RETURNING id
`

type DeletePostParams struct {
	ID       int32
	AuthorID uuid.UUID
}

func (q *Queries) DeletePost(ctx context.Context, arg DeletePostParams) (int32, error) {
	row := q.db.QueryRowContext(ctx, deletePost, arg.ID, arg.AuthorID)
	var id int32
	err := row.Scan(&id)
	return id, err
}

const getPost = `-- name: GetPost :one
SELECT id, title, body, published_on, banner_image_url, author_id, project_id, created_on, updated_on FROM posts WHERE id = $1 AND author_id = $2
`

type GetPostParams struct {
	ID       int32
	AuthorID uuid.UUID
}

func (q *Queries) GetPost(ctx context.Context, arg GetPostParams) (Post, error) {
	row := q.db.QueryRowContext(ctx, getPost, arg.ID, arg.AuthorID)
	var i Post
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.Body,
		&i.PublishedOn,
		&i.BannerImageUrl,
		&i.AuthorID,
		&i.ProjectID,
		&i.CreatedOn,
		&i.UpdatedOn,
	)
	return i, err
}

const getPostCount = `-- name: GetPostCount :one
SELECT COUNT(id) total_posts FROM posts WHERE author_id = $1
`

// POSTS --
func (q *Queries) GetPostCount(ctx context.Context, authorID uuid.UUID) (int64, error) {
	row := q.db.QueryRowContext(ctx, getPostCount, authorID)
	var total_posts int64
	err := row.Scan(&total_posts)
	return total_posts, err
}

const getPosts = `-- name: GetPosts :many
SELECT id, title, published_on FROM posts WHERE author_id = $1
`

type GetPostsRow struct {
	ID          int32
	Title       string
	PublishedOn time.Time
}

func (q *Queries) GetPosts(ctx context.Context, authorID uuid.UUID) ([]GetPostsRow, error) {
	rows, err := q.db.QueryContext(ctx, getPosts, authorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetPostsRow
	for rows.Next() {
		var i GetPostsRow
		if err := rows.Scan(&i.ID, &i.Title, &i.PublishedOn); err != nil {
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

const getUpcomingPosts = `-- name: GetUpcomingPosts :one
SELECT COUNT(id) upcoming_posts FROM posts WHERE author_id = $1 AND published_on > current_timestamp
`

func (q *Queries) GetUpcomingPosts(ctx context.Context, authorID uuid.UUID) (int64, error) {
	row := q.db.QueryRowContext(ctx, getUpcomingPosts, authorID)
	var upcoming_posts int64
	err := row.Scan(&upcoming_posts)
	return upcoming_posts, err
}

const insertPost = `-- name: InsertPost :one
INSERT INTO posts (title, body, published_on, banner_image_url, author_id, project_id) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, title, body, published_on, banner_image_url, author_id, project_id, created_on, updated_on
`

type InsertPostParams struct {
	Title          string
	Body           string
	PublishedOn    time.Time
	BannerImageUrl sql.NullString
	AuthorID       uuid.UUID
	ProjectID      int32
}

func (q *Queries) InsertPost(ctx context.Context, arg InsertPostParams) (Post, error) {
	row := q.db.QueryRowContext(ctx, insertPost,
		arg.Title,
		arg.Body,
		arg.PublishedOn,
		arg.BannerImageUrl,
		arg.AuthorID,
		arg.ProjectID,
	)
	var i Post
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.Body,
		&i.PublishedOn,
		&i.BannerImageUrl,
		&i.AuthorID,
		&i.ProjectID,
		&i.CreatedOn,
		&i.UpdatedOn,
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

const updatePost = `-- name: UpdatePost :one
UPDATE posts SET title = $1, body = $2, published_on = $3, banner_image_url = $4 WHERE id = $5 AND author_id = $6 RETURNING id, title, body, published_on, banner_image_url, author_id, project_id, created_on, updated_on
`

type UpdatePostParams struct {
	Title          string
	Body           string
	PublishedOn    time.Time
	BannerImageUrl sql.NullString
	ID             int32
	AuthorID       uuid.UUID
}

func (q *Queries) UpdatePost(ctx context.Context, arg UpdatePostParams) (Post, error) {
	row := q.db.QueryRowContext(ctx, updatePost,
		arg.Title,
		arg.Body,
		arg.PublishedOn,
		arg.BannerImageUrl,
		arg.ID,
		arg.AuthorID,
	)
	var i Post
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.Body,
		&i.PublishedOn,
		&i.BannerImageUrl,
		&i.AuthorID,
		&i.ProjectID,
		&i.CreatedOn,
		&i.UpdatedOn,
	)
	return i, err
}
