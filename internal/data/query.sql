-- name: GetProject :many
SELECT * FROM projects WHERE user_id = $1 LIMIT 1;