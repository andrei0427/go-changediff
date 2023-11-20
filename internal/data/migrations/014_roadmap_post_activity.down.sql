ALTER TABLE roadmap_posts DROP COLUMN is_locked;

DROP TABLE IF EXISTS roadmap_post_activity;
DROP TABLE IF EXISTS roadmap_post_comments;
DROP TABLE IF EXISTS roadmap_post_reactions;
DROP TABLE IF EXISTS roadmap_post_votes;
