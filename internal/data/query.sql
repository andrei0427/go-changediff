-- PROJECTS --
-- name: GetProject :many
SELECT * FROM projects WHERE user_id = $1 LIMIT 1;

-- name: GetProjectByKey :one
SELECT * FROM projects where app_key = $1 LIMIT 1;

-- name: InsertProject :one
INSERT INTO projects (name, description, accent_color, logo_url, app_key, user_id) VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: UpdateProject :one
UPDATE projects SET name = $1, description = $2, accent_color = $3, logo_url = $4, updated_on = CURRENT_TIMESTAMP WHERE id = $5 AND user_id = $6 RETURNING *;

-- name: DashboardQuery :one
SELECT COUNT(p.id), COUNT(rv.id), COUNT(rc.id) 
  FROM posts p 
    left join post_reactions rv on p.id = rv.id and rv.reaction is null 
    left join post_reactions rc on p.id = rc.id and (rc.id is null or rc.reaction is not null) 
  WHERE p.project_id = $1;

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
SELECT p.id, p.title, p.published_on, p.is_published, l.label, l.color, CASE WHEN p.published_on <= current_timestamp THEN 1 ELSE 0 END AS status, COUNT(r.id) as ViewCount FROM posts p left join labels l on p.label_id = l.id or p.label_id is null left join post_reactions r on (p.id = r.post_id and r.reaction is null) OR r.id is null WHERE p.project_id = $1 GROUP BY 1,2,3,4,5,6;

-- name: GetPost :one
SELECT p.*, l.label as Label FROM posts p LEFT JOIN labels l on p.label_id = l.id or p.label_id is null WHERE p.id = $1 AND p.project_id = $2;

-- name: GetPostReactions :many
SELECT 
  CASE WHEN r.reaction IS NULL THEN '' ELSE r.reaction END as Reaction, 
  COUNT(r.*) 
FROM posts p 
  JOIN post_reactions r ON r.post_id = p.id 
WHERE p.project_id = $1 
    and ($2 = r.user_id or $2 = '' or $2 is null)
    and ($3::varchar = r.user_uuid::varchar or $3 = '' or $3 is null)
    and (p.id = $4 OR $4 = 0)
GROUP BY r.reaction 
ORDER BY r.reaction NULLS FIRST;

-- name: GetPostComments :many
SELECT p.id, p.title, c.comment, c.created_on, r.locale, ur.reaction, REPLACE(r.user_name, '"', '') as UserName, REPLACE(r.user_role, '"', '') as UserRole
	FROM posts p 
		JOIN post_comments c ON c.post_id = p.id 
		JOIN post_reactions r ON r.user_uuid = c.user_uuid AND r.post_id = p.id AND r.reaction IS NULL 
		LEFT JOIN post_reactions ur ON ur.user_uuid = c.user_uuid AND ur.post_id = p.id AND ur.reaction IS NOT NULL 
WHERE p.project_id = $1
    and ($2 = r.user_id or $2 = '' or $2 is null)
    and ($3::varchar = r.user_uuid::varchar or $3 = '' or $3 is null)
    and (p.id = $4 OR $4 = 0)
ORDER BY c.created_on DESC;

-- name: GetPublishedPagedPosts :many
SELECT post.*, l.label, l.color, a.first_name, a.last_name, a.picture_url, r.reaction, CASE WHEN v.id IS NULL THEN 0 ELSE 1 END as Viewed
  FROM posts post 
    join projects proj on post.project_id = proj.id 
	join authors a on a.id = post.author_id 
	left join labels l on post.label_id = l.id or post.label_id is null 
	left join post_reactions r on (r.post_id = post.id and r.user_uuid = $4 and r.reaction is not null) or r.id is null 
	left join post_reactions v on (v.post_id = post.id and v.user_uuid = $4 and v.reaction is null) or v.id is null 
WHERE proj.app_key = $1 
   AND post.published_on <= CURRENT_TIMESTAMP 
   AND post.is_published = true
   AND ($5 = '' OR LOWER(post.title) LIKE $5)
ORDER BY post.published_on DESC 
LIMIT $2 
OFFSET $3;

-- name: InsertPost :one
INSERT INTO posts (title, body, published_on, is_published, label_id, author_id, project_id) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;

-- name: UpdatePost :one
UPDATE posts SET title = $1, body = $2, published_on = $3, is_published = $4, label_id = $5, updated_on = CURRENT_TIMESTAMP WHERE id = $6 AND project_id = $7 RETURNING *;

-- name: DeletePost :one
DELETE FROM posts WHERE id = $1 AND project_id = $2 RETURNING id;

-- name: InsertComment :one
INSERT INTO post_comments (user_uuid, comment, post_id) VALUES ($1, $2, $3) RETURNING *;

-- name: InsertReaction :one
INSERT INTO post_reactions (user_uuid, ip_addr, user_agent, locale, reaction, user_id, user_name, user_email, user_role, user_data, post_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING *;

-- name: UpdateReaction :one
UPDATE post_reactions SET reaction = $1 WHERE user_uuid = $2 AND post_id = $3 AND reaction IS NOT NULL RETURNING *;

-- name: GetReaction :many
SELECT reaction FROM post_reactions WHERE user_uuid = $1 AND post_id = $2 AND reaction IS NOT NULL;

-- name: UserViewed :one
SELECT COUNT(id) FROM post_reactions WHERE user_uuid = $1 AND post_id = $2 AND reaction IS NULL;

-- name: AnalyticsUsers :many
SELECT DISTINCT 
   COALESCE(REPLACE(r.user_id, '"', ''), 'N/A') as UserID,
   r.user_uuid,
   COALESCE(REPLACE(r.user_name, '"', ''), 'User') as UserName, 
   COALESCE(REPLACE(r.user_email, '"', ''), 'N/A') as UserEmail, 
   COALESCE(REPLACE(r.user_role, '"', ''), 'N/A') as UserRole, 
   r.locale,
   CAST(STRING_AGG(distinct r.ip_addr, ',') as text) as IPAddress,
   CAST(STRING_AGG(distinct r.user_agent, ',') as text) as UserAgent,
   CASE WHEN r.user_data IS NOT NULL THEN CAST(r.user_data as text) ELSE NULL END as UserData,
   COUNT(DISTINCT r.id) as ViewCount, 
   COUNT(DISTINCT ri.id) as ImpressionCount, 
   COUNT(DISTINCT rc.id) as CommentCount 
FROM post_reactions r 
  JOIN posts p on p.id = r.post_id 
  left join post_comments rc ON r.user_uuid = rc.user_uuid 
  left join post_reactions ri on r.user_uuid = ri.user_uuid and r.post_id = ri.post_id and ri.reaction is not null 
  WHERE p.project_id = $1 
    and ($2 = r.user_id or $2 = '' or $2 is null)
    and ($3::varchar = r.user_uuid::varchar or $3 = '' or $3 is null)
    and r.reaction is null 
GROUP BY r.user_id, r.user_uuid, UserName, UserEmail, UserRole, r.locale, UserData;

-- ROADMAP --

-- name: GetBoards :many
SELECT id, name, is_private FROM roadmap_boards WHERE project_id = $1 order by created_on;

-- name: GetBoard :one
SELECT id, name, description, is_private FROM roadmap_boards WHERE id = $1 AND project_id = $2;

-- name: InsertBoard :one
INSERT INTO roadmap_boards (name, is_private, description, project_id, created_on) VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP) RETURNING *;

-- name: UpdateBoard :one
UPDATE roadmap_boards SET name = $1, is_private = $2, description = $3 WHERE id = $4 AND project_id = $5 RETURNING *;

-- name: DeleteBoard :one
DELETE FROM roadmap_boards WHERE id = $1 and project_id = $2 RETURNING id;

-- name: GetStatuses :many
SELECT id, status, color FROM roadmap_statuses WHERE project_id = $1;

-- name: GetStatus :one
SELECT id, status, description, color FROM roadmap_statuses WHERE id = $1 AND project_id = $2;

-- name: InsertStatus :one
INSERT INTO roadmap_statuses (status, description, color, project_id, created_on) VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP) RETURNING *;

-- name: UpdateStatus :one
UPDATE roadmap_statuses SET status = $1, description = $2, color = $3 WHERE id = $4 AND project_id = $5 RETURNING *;

-- name: DeleteStatus :one
DELETE FROM roadmap_statuses WHERE id = $1 and project_id = $2 RETURNING id;

-- name: HasPostsForBoard :one
SELECT COUNT(*) FROM roadmap_posts WHERE board_id = $1;

-- name: HasPostsForStatus :one
SELECT COUNT(*) FROM roadmap_posts WHERE status_id = $1;