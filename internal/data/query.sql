-- PROJECTS --
-- name: GetProject :many
SELECT * FROM projects WHERE user_id = $1 LIMIT 1;

-- name: GetProjectByKey :one
SELECT * FROM projects where app_key = $1 LIMIT 1;

-- name: InsertProject :one
INSERT INTO projects (name, description, accent_color, logo_url, app_key, user_id) VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;


-- POSTS --
-- name: GetPostCount :one
SELECT COUNT(id) total_posts FROM posts WHERE author_id = $1;

-- name: GetPosts :many
SELECT id, title, published_on FROM posts WHERE author_id = $1;

-- name: GetPost :one
SELECT * FROM posts WHERE id = $1 AND author_id = $2;

-- name: GetPublishedPagedPosts :many
SELECT post.* FROM posts post join projects proj on post.project_id = proj.id WHERE proj.app_key = $1 AND post.published_on <= CURRENT_TIMESTAMP ORDER BY post.published_on LIMIT $2 OFFSET $3;

-- name: InsertPost :one
INSERT INTO posts (title, body, published_on, banner_image_url, author_id, project_id) VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: UpdatePost :one
UPDATE posts SET title = $1, body = $2, published_on = $3, banner_image_url = $4 WHERE id = $5 AND author_id = $6 RETURNING *;

-- name: DeletePost :one
DELETE FROM posts WHERE id = $1 AND author_id = $2 RETURNING id;

