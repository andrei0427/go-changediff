-- Create Reaction tables

CREATE TABLE IF NOT EXISTS viewers (
    id SERIAL PRIMARY KEY,
    user_uuid uuid NOT NULL,
    ip_addr text NOT NULL,
    user_agent text NOT NULL,
    locale text NOT NULL,

    user_data JSON,
    user_id text,
    user_name text,
    user_email text,
    user_role text,

    project_id integer REFERENCES projects(id) NOT NULL,
    created_on timestamp NOT NULL DEFAULT current_timestamp
);

CREATE TABLE IF NOT EXISTS changelog_interaction_types (
    id SERIAL PRIMARY KEY,
    type text NOT NULL,
    created_on timestamp not null default current_timestamp
);

INSERT INTO changelog_interaction_types (id, type) VALUES
(1, 'VIEW'),
(2, 'REACTION'),
(3, 'COMMENT');

CREATE TABLE IF NOT EXISTS changelog_interactions (
    id SERIAL PRIMARY KEY,
    content text null,

    interaction_type_id int REFERENCES changelog_interaction_types(id) NOT NULL,
    post_id integer REFERENCES posts(id) NOT NULL,
    viewer_id integer REFERENCES viewers(id) NOT NULL, 
    project_id integer REFERENCES projects(id) NOT NULL,
    created_on timestamp not null default current_timestamp
);