CREATE TABLE IF NOT EXISTS roadmap_boards (
    id SERIAL PRIMARY KEY,
    name text NOT NULL,
    is_private boolean not null default(false),
    description text not null default '',

    created_on timestamp NOT NULL DEFAULT current_timestamp,
    project_id integer REFERENCES projects(id) NOT NULL
);

CREATE TABLE IF NOT EXISTS roadmap_statuses (
    id SERIAL PRIMARY KEY,
    status text NOT NULL,
    color text NOT NULL,
    description text not null default '',

    created_on timestamp NOT NULL DEFAULT current_timestamp,
    project_id integer REFERENCES projects(id) NOT NULL
);

CREATE TABLE IF NOT EXISTS roadmap_posts (
    id SERIAL PRIMARY KEY,
    title text NOT NULL,
    body TEXT NOT NULL,
    due_date date null,

    board_id integer REFERENCES roadmap_boards(id) NULL,
    project_id integer REFERENCES projects(id) NOT NULL,
    status_id integer REFERENCES roadmap_statuses(id)  NULL,
    created_on timestamp NOT NULL DEFAULT current_timestamp
);