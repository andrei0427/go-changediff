-- PROJECTS --
-- name: GetProject :many
SELECT * FROM projects WHERE user_id = $1 LIMIT 1;

-- name: GetProjectByKey :one
SELECT * FROM projects where app_key = $1 LIMIT 1;

-- name: InsertProject :one
INSERT INTO projects (name, description, accent_color, logo_url, app_key, user_id) VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: UpdateProject :one
UPDATE projects SET name = $1, description = $2, accent_color = $3, logo_url = $4, updated_on = CURRENT_TIMESTAMP WHERE id = $5 AND user_id = $6 RETURNING *;

-- LABELS --
-- name: GetLabels :many
SELECT * from labels WHERE project_id = $1 ORDER BY created_on;

-- name: InsertLabel :one
INSERT INTO labels (label, color, project_id) VALUES ($1, $2, $3) RETURNING *;

-- name: UpdateLabel :one
UPDATE labels SET label = $1, color = $2 WHERE id = $3 AND project_id = $4 RETURNING *;

-- name: UnsetLabels :many
UPDATE posts SET label_id = NULL WHERE id = $1 AND project_id = $2 RETURNING id;

-- name: DeleteLabel :one
DELETE FROM labels WHERE id = $1 AND project_id = $2 RETURNING id;

-- AUTHOR --
-- name: GetAuthorByUser :many
SELECT * FROM authors WHERE user_id = $1 LIMIT 1;

-- name: InsertAuthor :one
INSERT INTO authors (first_name, last_name, picture_url, user_id, project_id) VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- POSTS --
-- name: GetPostCount :one
SELECT COUNT(id) total_posts FROM posts WHERE project_id = $1;

-- name: GetPosts :many
SELECT p.id, p.title, p.published_on, l.label, l.color, CASE WHEN p.published_on <= current_timestamp THEN 1 ELSE 0 END AS status FROM posts p left join labels l on p.label_id = l.id or p.label_id is null WHERE p.project_id = $1;

-- name: GetPost :one
SELECT p.*, l.label as Label FROM posts p LEFT JOIN labels l on p.label_id = l.id or p.label_id is null WHERE p.id = $1 AND p.project_id = $2;

-- name: GetPublishedPagedPosts :many
SELECT post.*, l.label, l.color, a.first_name, a.last_name, a.picture_url FROM posts post join projects proj on post.project_id = proj.id join authors a on a.id = post.author_id left join labels l on post.label_id = l.id or post.label_id is null WHERE proj.app_key = $1 AND post.published_on <= CURRENT_TIMESTAMP ORDER BY post.published_on DESC LIMIT $2 OFFSET $3;

-- name: InsertPost :one
INSERT INTO posts (title, body, published_on, label_id, author_id, project_id) VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: UpdatePost :one
UPDATE posts SET title = $1, body = $2, published_on = $3, label_id = $4, updated_on = CURRENT_TIMESTAMP WHERE id = $5 AND project_id = $6 RETURNING *;

-- name: DeletePost :one
DELETE FROM posts WHERE id = $1 AND project_id = $2 RETURNING id;

