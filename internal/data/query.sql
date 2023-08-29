-- name: GetProject :many
SELECT * FROM projects WHERE user_id = $1 LIMIT 1;

-- name: InsertProject :one
INSERT INTO projects (name, description, accent_color, logo_url, app_key, user_id) VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: GetPostCount :one
SELECT COUNT(id) total_posts FROM posts WHERE author_id = $1;

-- name: GetPosts :many
SELECT id, title, published_on FROM posts WHERE author_id = $1;

-- name: GetUpcomingPosts :one
SELECT COUNT(id) upcoming_posts FROM posts WHERE author_id = $1 AND published_on > current_timestamp;

-- name: InsertPost :one
INSERT INTO posts (title, body, published_on, banner_image_url, author_id, project_id) VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: DeletePost :one
DELETE FROM posts WHERE id = $1 AND author_id = $2 RETURNING id;

