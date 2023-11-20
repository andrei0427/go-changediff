CREATE TABLE IF NOT EXISTS roadmap_categories (
	id SERIAL PRIMARY KEY,
	name text NOT NULL,
	emoji text NOT NULL,
	is_private boolean not null default(false),

	project_id INT REFERENCES projects(id) NOT NULL,
	creator_id INT REFERENCES authors(id) NOT NULL,
	created_on TIMESTAMP NOT NULL DEFAULT(CURRENT_TIMESTAMP)
);

ALTER TABLE roadmap_statuses ADD COLUMN is_private boolean not null default(false);
ALTER TABLE roadmap_posts ADD COLUMN is_private boolean not null default(false);
ALTER TABLE roadmap_posts ADD COLUMN author_id int references authors(id);
ALTER TABLE roadmap_posts ADD COLUMN viewer_id int references viewers(id) null;
ALTER TABLE roadmap_posts ADD COLUMN is_idea boolean not null default(false);

CREATE TABLE IF NOT EXISTS roadmap_post_categories (
	id SERIAL PRIMARY KEY,
	roadmap_post_id INT REFERENCES roadmap_posts(id) not null,
	roadmap_category_id INT REFERENCES roadmap_posts(id) not null,

	project_id INT REFERENCES projects(id) NOT NULL,
	created_on TIMESTAMP NOT NULL DEFAULT(CURRENT_TIMESTAMP)
);