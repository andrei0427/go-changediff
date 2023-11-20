ALTER TABLE roadmap_posts ADD COLUMN is_locked boolean not null default(false);

CREATE TABLE IF NOT EXISTS roadmap_post_activity (
	id SERIAL PRIMARY KEY,
	from_status_id int references roadmap_statuses(id) null,
	to_status_id int references roadmap_statuses(id) null,

	roadmap_post_id int REFERENCES roadmap_posts(id) NOT NULL,
	author_id int REFERENCES authors(id) NOT NULL,
	created_on timestamp not null default current_timestamp
);

CREATE TABLE IF NOT EXISTS roadmap_post_comments (
	id SERIAL PRIMARY KEY,
	content text not null,
	is_pinned boolean not null default(false),
	is_deleted boolean not null default(false),

	in_reply_to_id int references roadmap_post_comments(id) NULL,

	roadmap_post_id int REFERENCES roadmap_posts(id) NOT NULL,
	author_id int REFERENCES authors(id) NULL,
	viewer_id int REFERENCES viewers(id) NULL,
	created_on timestamp not null default current_timestamp
);

CREATE TABLE IF NOT EXISTS roadmap_post_reactions (
	id SERIAL PRIMARY KEY,
	emoji varchar(1) not null,

	comment_id int references roadmap_post_comments(id) NULL,

	roadmap_post_id int REFERENCES roadmap_posts(id) NOT NULL,
	author_id int REFERENCES authors(id) NULL,
	viewer_id int REFERENCES viewers(id) NULL,
	created_on timestamp not null default current_timestamp
);

CREATE TABLE IF NOT EXISTS roadmap_post_votes (
	id SERIAL PRIMARY KEY,

	roadmap_post_id int REFERENCES roadmap_posts(id) NOT NULL,
	author_id int REFERENCES authors(id) NULL,
	viewer_id int REFERENCES viewers(id) NULL,
	created_on timestamp not null default current_timestamp
);