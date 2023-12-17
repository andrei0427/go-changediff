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
SELECT COUNT(p.id), COUNT(civ.id), COUNT(cic.id) 
  FROM posts p 
    left join changelog_interactions civ on p.id = civ.post_id and (civ.id is null or civ.interaction_type_id = 1)
    left join changelog_interactions cic on p.id = cic.post_id and (cic.id is null or cic.interaction_type_id = 3) 
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
SELECT a.*, 
   s.id, s.subscription_start_date, s.is_annual, s.tier, 
   CASE 
    WHEN s.is_annual = false 
      THEN (s.subscription_start_date + INTERVAL '1 month') >= CURRENT_TIMESTAMP
    WHEN s.is_annual = true 
      THEN (s.subscription_start_date + INTERVAL '1 year') >= CURRENT_TIMESTAMP 
    ELSE false END as is_active,
   CASE 
    WHEN s.is_annual = false 
      THEN s.subscription_start_date + INTERVAL '1 month'
    WHEN s.is_annual = true 
      THEN s.subscription_start_date + INTERVAL '1 year'
    ELSE NULL END as expires_on
FROM authors a
  LEFT JOIN subscriptions s on 
    (a.id = s.subscriber_id and 
     s.subscription_start_date <= current_timestamp and
     s.success = true and
     s.stopped = false 
    ) 
    or s.id is null
WHERE a.user_id = $1
ORDER BY s.subscription_start_date DESC
LIMIT 1;

-- name: InsertAuthor :one
INSERT INTO authors (first_name, last_name, picture_url, user_id, project_id) VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: UpdateAuthor :one
UPDATE authors SET first_name = $1, last_name = $2, picture_url = $3 WHERE user_id = $4 and project_id = $5 RETURNING *;

-- VIEWER --
-- name: InsertViewer :one
INSERT INTO viewers (user_uuid, ip_addr, user_agent, locale, user_data, user_id, user_name, user_email, user_role, project_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING *;

-- name: GetViewer :many
SELECT * FROM viewers WHERE user_uuid = $1 OR (user_id IS NULL OR user_id = $2) LIMIT 1;

-- name: GetViewersByProject :many
SELECT * FROM viewers WHERE project_id = $1 ORDER BY created_on DESC;

-- name: UpdateViewer :one
UPDATE viewers SET user_uuid = $1, ip_addr = $2, user_agent = $3, locale = $4, user_data = $5, user_id = $6, user_name = $7, user_email = $8, user_role = $9 WHERE id = $10 RETURNING *;

-- POSTS --
-- name: GetPostCount :one
SELECT COUNT(id) total_posts FROM posts WHERE project_id = $1;

-- name: GetPosts :many
SELECT p.id, p.title, p.published_on, p.is_published, p.expires_on, l.label, l.color, 
  CASE WHEN p.published_on <= current_timestamp THEN 
    CASE WHEN p.expires_on is not null and p.expires_on <= current_timestamp THEN 2 
       ELSE 1 END 
      ELSE 0 END AS status, 
  COUNT(r.id) as ViewCount FROM posts p 
       left join labels l on p.label_id = l.id 
       left join changelog_interactions r on (p.id = r.post_id and r.interaction_type_id = 1) OR r.id is null 
  WHERE p.project_id = $1
  GROUP BY 1,2,3,4,5,6,7
  ORDER BY p.published_on DESC;


-- name: GetPost :one
SELECT p.*, l.label as Label FROM posts p LEFT JOIN labels l on p.label_id = l.id or p.label_id is null WHERE p.id = $1 AND p.project_id = $2;

-- name: GetPostReactions :many
SELECT 
  CASE WHEN ci.content IS NULL THEN '' ELSE ci.content END as Reaction, 
  COUNT(ci.*) 
FROM changelog_interactions ci 
WHERE ci.project_id = $1 
  AND ($2 = 0 OR ci.post_id = $2)
  AND ($3 = 0 OR ci.viewer_id = $3)
GROUP BY ci.content
ORDER BY ci.content NULLS FIRST;

-- name: GetPostComments :many
SELECT p.id, 
       p.title, 
       ci.content, 
       ci.created_on, 
       v.locale, 
       r.content as Reaction, 
       REPLACE(v.user_name, '"', '') as UserName, 
       REPLACE(v.user_role, '"', '') as UserRole
	FROM posts p 
		JOIN changelog_interactions ci ON ci.post_id = p.id AND ci.interaction_type_id = 3
		JOIN viewers v ON v.id = ci.viewer_id 
		LEFT JOIN changelog_interactions r ON (r.post_id = p.id AND r.viewer_id = v.id AND r.interaction_type_id = 2) OR r.id is null
WHERE p.project_id = $1
    and ($2 = 0 OR v.id = $2)
    and ($3 = 0 OR p.id = $3)
ORDER BY ci.created_on DESC;

-- name: GetPublishedPagedPosts :many
SELECT distinct post.*, l.label, l.color, a.first_name, a.last_name, a.picture_url, r.content as Reaction, 
      CASE WHEN v.id IS NULL THEN 0 ELSE 1 END as Viewed
  FROM posts post 
    join projects proj on post.project_id = proj.id 
	  join authors a on a.id = post.author_id 
    left join labels l on post.label_id = l.id or post.label_id is null 
    left join changelog_interactions r on (r.post_id = post.id and r.interaction_type_id = 2 and r.viewer_id = $4) or r.id is null
    left join changelog_interactions v on (v.post_id = post.id and v.interaction_type_id = 1 and v.viewer_id = $4) or v.id is null 
WHERE proj.app_key = $1 
   AND post.published_on <= CURRENT_TIMESTAMP 
   AND (post.expires_on IS NULL OR post.expires_on >= CURRENT_TIMESTAMP)
   AND post.is_published = true
   AND ($5 = '' OR LOWER(post.title) LIKE $5)
ORDER BY post.published_on DESC 
LIMIT $2 
OFFSET $3;

-- name: InsertPost :one
INSERT INTO posts (title, body, published_on, is_published, label_id, author_id, project_id) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;

-- name: UpdatePost :one
UPDATE posts SET title = $1, body = $2, published_on = $3, is_published = $4, label_id = $5, expires_on = $6, updated_on = CURRENT_TIMESTAMP WHERE id = $7 AND project_id = $8 RETURNING *;

-- name: DeletePost :one
DELETE FROM posts WHERE id = $1 AND project_id = $2 RETURNING id;

-- name: DeleteInteractions :many
DELETE FROM changelog_interactions WHERE post_id = $1 AND project_id = $2 RETURNING id;

-- name: InsertInteraction :one
INSERT INTO changelog_interactions (content, post_id, interaction_type_id, viewer_id, project_id) VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: UpdateInteraction :one
UPDATE changelog_interactions SET content = $1 WHERE viewer_id = $2 AND post_id = $3 AND interaction_type_id = $4 RETURNING *;

-- name: GetReaction :many
SELECT content FROM changelog_interactions WHERE post_id = $1 AND viewer_id = $2 AND project_id = $3 AND interaction_type_id = 2;

-- name: UserViewed :one
SELECT COUNT(id) FROM changelog_interactions WHERE post_id = $1 and viewer_id = $2 AND interaction_type_id = 1;

-- name: AnalyticsUsers :many
SELECT DISTINCT 
   COALESCE(REPLACE(v.user_id, '"', ''), 'N/A') as UserID,
   v.user_uuid,
   v.id,
   COALESCE(REPLACE(v.user_name, '"', ''), 'User') as UserName, 
   COALESCE(REPLACE(v.user_email, '"', ''), 'N/A') as UserEmail, 
   COALESCE(REPLACE(v.user_role, '"', ''), 'N/A') as UserRole, 
   v.locale,
   CAST(STRING_AGG(distinct v.ip_addr, ',') as text) as IPAddress,
   CAST(STRING_AGG(distinct v.user_agent, ',') as text) as UserAgent,
   CASE WHEN v.user_data IS NOT NULL THEN CAST(v.user_data as text) ELSE NULL END as UserData,
   COUNT(DISTINCT civ.id) as ViewCount, 
   COUNT(DISTINCT cir.id) as ImpressionCount, 
   COUNT(DISTINCT cic.id) as CommentCount 
FROM viewers v 
  LEFT JOIN changelog_interactions civ on (civ.project_id = $1 and civ.viewer_id = v.id and civ.interaction_type_id = 1) or civ.id is null
  LEFT JOIN changelog_interactions cir on (cir.project_id = $1 and cir.viewer_id = v.id and cir.interaction_type_id = 2) or cir.id is null
  LEFT JOIN changelog_interactions cic on (cic.project_id = $1 and cic.viewer_id = v.id and cic.interaction_type_id = 3) or cic.id is null
WHERE $2 = 0 OR v.id = $2
GROUP BY v.user_id, v.user_uuid, v.id, UserName, UserEmail, UserRole, v.locale, UserData;

-- IDEAS --
-- name: GetIdeas :many
SELECT rp.id, 
  rp.title,
  rp.created_on,
  rp.is_private, 
  rp.status_id,
  rp.board_id,
  rs.status,
  rs.color,
  rb.name,
  COALESCE(a.first_name, v.user_name, 'Someone') as Who
from roadmap_posts rp 
  left join roadmap_statuses rs on rs.id = rp.status_id
  left join roadmap_boards rb on rb.id = rp.board_id
  left join authors a on a.id = rp.author_id
  left join viewers v on v.id = rp.viewer_id
where rp.project_id = $1
  and rp.is_idea = true
  and ($2 = 0 or a.id = $2)
  and ($3 = 0 or v.id = $3)
order by rp.created_on desc;

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
SELECT id, status, sort_order, color FROM roadmap_statuses WHERE project_id = $1 ORDER BY sort_order;

-- name: GetStatus :one
SELECT id, status, description, sort_order, color FROM roadmap_statuses WHERE id = $1 AND project_id = $2;

-- name: InsertStatus :one
INSERT INTO roadmap_statuses (status, description, color, project_id, created_on, sort_order) VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, $5) RETURNING *;

-- name: UpdateStatus :one
UPDATE roadmap_statuses SET status = $1, description = $2, color = $3 WHERE id = $4 AND project_id = $5 RETURNING *;

-- name: GetNextSortOrderForStatus :one
SELECT MAX(COALESCE(sort_order, 0)) + 1 as NextSortOrder FROM roadmap_statuses WHERE project_id = $1;

-- name: UpdateStatusOrder :one
UPDATE roadmap_statuses SET sort_order = $1 WHERE id = $2 AND project_id = $3 RETURNING id, status, sort_order, color;

-- name: DeleteStatus :one
DELETE FROM roadmap_statuses WHERE id = $1 and project_id = $2 RETURNING id;

-- name: HasPostsForBoard :one
SELECT COUNT(*) FROM roadmap_posts WHERE board_id = $1;

-- name: HasPostsForStatus :one
SELECT COUNT(*) FROM roadmap_posts WHERE status_id = $1;

-- name: GetPostsForBoard :many
SELECT rp.id,
  rp.status_id,
  rp.board_id,
  rp.title,
  rp.is_idea,
  rp.is_pinned,
  COUNT(rpv.id) as Votes,  
  $3 = any(array_agg(rpv.author_id))
    or $4 = any(array_agg(rpv.viewer_id)) as Voted
from roadmap_posts rp 
  left join authors a on a.id = rp.author_id
  left join viewers v on v.id = rp.viewer_id
  left join roadmap_post_votes rpv on rpv.roadmap_post_id = rp.id
where (rp.board_id IS NULL OR rp.board_id = $1) and rp.project_id = $2 
group by rp.id, rp.board_id, rp.status_id, rp.title
order by is_pinned DESC, due_date;

-- name: InsertRoadmapPost :one
INSERT INTO roadmap_posts (title, body, due_date, is_private, author_id, is_idea, viewer_id, board_id, status_id, project_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING *;

-- name: UpdateRoadmapPost :one
UPDATE roadmap_posts SET title = $1, body = $2, due_date = $3, is_private = $4, board_id = $5, status_id = $6 WHERE id = $7 AND project_id = $8 RETURNING *;

-- name: ToggleRoadmapPostPin :one
UPDATE roadmap_posts SET is_pinned = not is_pinned WHERE id = $1 AND project_id = $2 RETURNING is_pinned;

-- name: GetRoadmapPost :one
SELECT p.id, 
p.title, 
p.is_private, 
p.body,
p.due_date,
p.status_id,
p.is_pinned,
p.is_idea,
p.board_id, 
p.created_on, 
COUNT(v.id) as Votes,
$3 = any(array_agg(v.author_id))
or $4 = any(array_agg(v.viewer_id)) as Voted
FROM roadmap_posts p 
  left join roadmap_post_votes v on p.id = v.roadmap_post_id 
WHERE p.id = $1 AND p.project_id = $2 
GROUP BY p.id, p.title, p.is_private, p.body, p.due_date, p.status_id, p.board_id, p.created_on;

-- name: UpdateRoadmapPostStatus :one
UPDATE roadmap_posts SET status_id = $1, board_id = $2 WHERE id = $3 AND project_id = $4 RETURNING *;

-- name: DeleteRoadmapPost :one
DELETE FROM roadmap_posts WHERE id = $1 AND project_id = $2 RETURNING id;

-- name: DeleteRoadmapPostCategoriesByPost :many
DELETE FROM roadmap_post_categories where roadmap_post_id = $1 AND project_id = $2 RETURNING id;

-- name: GetRoadmapPostActivity :one
select
  p.id as ID,
  a.first_name as Who,
  p.created_on as CreatedOn,
  a.picture_url as WhoPictureUrl
from roadmap_posts p 
  join authors a on a.id = p.author_id
where p.id = $1 and p.project_id = $2;

-- name: GetRoadmapPostStatusActivity :many
select 
   t.id as ID,
   a.first_name as Who,
   a.picture_url as WhoPictureUrl,
   sf.ID as StatusFromID,
   sf.Status as StatusFrom,
   sf.Color as StatusFromColor,
   sf.Description as StatusFromDescription,
   st.ID as StatusToID,
   st.Status as StatusTo,
   st.Color as StatusToColor,
   st.Description as StatusToDescription,
   t.created_on as CreatedOn
from
  roadmap_post_activity t 
    join authors a on a.id = t.author_id
    join roadmap_posts p on p.id = t.roadmap_post_id
    left join roadmap_statuses sf on sf.id = t.from_status_id
    left join roadmap_statuses st on st.id =  t.to_status_id
where t.roadmap_post_id = $1 and p.project_id = $2;

-- name: InsertRoadmapPostActivity :one
insert into roadmap_post_activity (from_status_id, to_status_id, roadmap_post_id, author_id) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: DeleteRoadmapPostActivity :many
delete from roadmap_post_activity rpa 
   using roadmap_posts rp 
   where rpa.roadmap_post_id = rp.id 
     and rp.id = $1 
     and rp.project_id = $2
     and ($3 = 0 OR rp.author_id = $3)
     and ($4 = 0 OR rp.viewer_id = $4)
RETURNING rpa.id;

-- name: GetRoadmapPostComments :many
select 
   c.id as ID,
   COALESCE(a.first_name, v.user_name, 'Someone') as Who,
   a.picture_url as WhoPictureUrl,
   c.content as Comment,
   c.created_on as CreatedOn,
   c.is_pinned as IsPinned,
   c.is_deleted as IsDeleted,
   COUNT(rc.id) as ReplyCount
from
  roadmap_post_comments c
    inner join roadmap_posts p on p.id = c.roadmap_post_id
    left join authors a on a.id = c.author_id
    left join viewers v on v.id = c.viewer_id
    left join roadmap_post_comments rc on rc.in_reply_to_id = c.id
where c.roadmap_post_id = $1 
    and (($2 = 0 and c.in_reply_to_id is null) or $2 = c.in_reply_to_id)
    and p.project_id = $3
group by
  c.id, Who, WhoPictureUrl, Comment, CreatedOn, IsPinned, IsDeleted;

-- name: GetRoadmapPostReactions :many
select
  string_agg(distinct a.first_name, ',') as Authors,
  string_agg(distinct v.user_name, ',') as Viewers,
  emoji as Reaction,
  count(r.id) as count,
  r.comment_id as CommentID,
  $4 = any (array_agg(a.ID))
  or $5 = any (array_agg(v.ID)) as Reacted
from
  roadmap_post_reactions r
  inner join roadmap_posts p on p.id = r.roadmap_post_id
  left join authors a on r.author_id = a.id
  left join viewers v on r.viewer_id = v.id
  left join roadmap_post_comments c on r.comment_id = c.id
where
  r.roadmap_post_id = $1
  and (($2 = 0 and r.comment_id is null) or $2 = r.comment_id)
  and p.project_id = $3
  and ($6 = '' OR emoji = $6)
group by
  r.comment_id,
  emoji;

-- name: InsertRoadmapPostVote :one
insert into roadmap_post_votes (roadmap_post_id, author_id, viewer_id) 
  select $1, $2, $3 from roadmap_posts p where p.id = $1 and p.project_id = $4
RETURNING *;

-- name: GetRoadmapPostVoteStatus :one
select 
   CASE WHEN $1 = any(array_agg(rpv.author_id)) OR $2 = any(array_agg(rpv.viewer_id)) THEN rpv.id ELSE null END as VoteID,
   COUNT(rpv.id) as Votes
from roadmap_posts rp left join roadmap_post_votes rpv on rp.id = rpv.roadmap_post_id
where
  rp.id = $3 and
  rp.project_id = $4
group by rpv.id;

-- name: GetRoadmapPostOwner :one
select author_id, viewer_id from roadmap_post_votes where id = $1;

-- name: DeleteRoadmapPostVote :many
delete from roadmap_post_votes rpv 
   using roadmap_posts rp 
   where rpv.roadmap_post_id = rp.id 
     and ($1 = 0 or rpv.id = $1)
     and rp.id = $2 
     and rp.project_id = $3 
     and ($4 = 0 OR rp.author_id = $4)
     and ($5 = 0 OR rp.viewer_id = $5)
RETURNING rpv.id;


-- name: InsertRoadmapPostComment :one
insert into roadmap_post_comments (content, in_reply_to_id, roadmap_post_id, author_id, viewer_id) VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: DeleteRoadmapPostComment :one
update roadmap_post_comments rpc 
   set is_deleted = true 
   from roadmap_posts rp 
   where rpc.roadmap_post_id = rp.id 
     and rpc.id = $1
     and rp.id = $2 
     and rp.project_id = $3 
     and (($4 = 0 OR rp.author_id = $4)
        or ($4 = 0 OR rpc.author_id = $4)
        or ($5 = 0 OR rpc.viewer_id = $5))
RETURNING rpc.id;

-- name: DeleteAllRoadmapPostComments :many
delete from roadmap_post_comments rpc 
  using roadmap_posts rp
   where rpc.roadmap_post_id = rp.id 
     and rp.id = $1 
     and rp.project_id = $2 
RETURNING rpc.id;

-- name: TogglePinRoadmapPostComment :one
update roadmap_post_comments rpc
set is_pinned = not rpc.is_pinned 
from roadmap_posts rp
where rpc.roadmap_post_id = rp.id
  and rpc.id = $1
  and rp.id = $2
  and rp.project_id = $3
RETURNING rpc.id;

-- name: InsertRoadmapPostReaction :one
insert into roadmap_post_reactions (emoji, comment_id, roadmap_post_id, author_id, viewer_id) VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: DeleteRoadmapPostReaction :many
delete from roadmap_post_reactions rpr 
   using roadmap_posts rp 
   where roadmap_post_id = rp.id 
     and ($1 = '' or rpr.emoji = $1)
     and rp.id = $2 
     and rp.project_id = $3 
RETURNING rpr.id;

-- name: ToggleLockRoadmapPost :one
update roadmap_posts set is_locked = !is_locked where id = $1 RETURNING id;

-- name: GetWidgetRoadmapData :many
select 
 rp.id,
 s.id,
 s.status,
 s.color,
 b.name,
 rp.title, 
 rp.created_on,
 rp.due_date,
 rp.is_pinned,
 rp.is_idea,
 COUNT(rpc.id) as CommentCount
from roadmap_posts rp
  inner join roadmap_statuses s on s.id = rp.status_id and s.is_private = false
  inner join roadmap_boards b on b.id = rp.board_id and b.is_private = false
  left join roadmap_post_comments rpc on rpc.roadmap_post_id = rp.id
where rp.project_id = $1 and rp.is_private = false
group by rp.id, s.id, s.status, s.color, b.name, rp.title, rp.created_on, rp.due_date, rp.is_pinned, rp.is_idea
order by s.sort_order ASC, rp.is_pinned DESC, rp.created_on ASC;


