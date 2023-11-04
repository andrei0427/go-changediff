DROP TABLE roadmap_post_categories;

ALTER TABLE roadmap_posts DROP COLUMN IF EXISTS user_uuid;
ALTER TABLE roadmap_posts DROP COLUMN IF EXISTS author_id;
ALTER TABLE roadmap_posts DROP COLUMN IF EXISTS is_private;
ALTER TABLE roadmap_posts DROP COLUMN IF EXISTS is_idea;
ALTER TABLE roadmap_statuses DROP COLUMN IF EXISTS is_private;

DROP TABLE roadmap_categories;
