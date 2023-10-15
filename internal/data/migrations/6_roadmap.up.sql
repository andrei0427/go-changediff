CREATE TABLE IF NOT EXISTS roadmap_boards (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    is_private boolean not null default(false) ,
    description text not null default '',

    created_on timestamp NOT NULL DEFAULT current_timestamp,
    created_by uuid NOT NULL
);

CREATE TABLE IF NOT EXISTS roadmap_statuses (
    id SERIAL PRIMARY KEY,
    status VARCHAR(255) NOT NULL,
    description text not null default '',

    created_on timestamp NOT NULL DEFAULT current_timestamp,
    created_by uuid NOT NULL
);

CREATE TABLE IF NOT EXISTS roadmap_posts (
    id SERIAL PRIMARY KEY,
    body TEXT NOT NULL,
    due_date date null,

    board_id integer REFERENCES roadmap_boards(id) NOT NULL,
    status_id integer REFERENCES roadmap_statuses(id) NOT NULL,
    created_on timestamp NOT NULL DEFAULT current_timestamp,
    created_by uuid NOT NULL
);