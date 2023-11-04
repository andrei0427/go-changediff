-- Create Comments tables
CREATE TABLE IF NOT EXISTS post_comments (
    id SERIAL PRIMARY KEY,
    user_uuid uuid NOT NULL,
    comment TEXT NOT NULL,

    post_id integer REFERENCES posts(id) NOT NULL,
    created_on timestamp NOT NULL DEFAULT current_timestamp
);
