-- name: GetProjects :many
SELECT * FROM projects WHERE user_id = $1 ORDER BY created_on DESC;