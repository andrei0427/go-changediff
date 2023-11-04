-- Create Reaction tables
CREATE TABLE IF NOT EXISTS post_reactions (
    id SERIAL PRIMARY KEY,
    user_uuid uuid NOT NULL,
    ip_addr varchar(255) NOT NULL,
    user_agent varchar(255) NOT NULL,
    locale VARCHAR(255) NOT NULL,
    reaction VARCHAR(1) NULL, -- if null, then this record is a 'view'

    post_id integer REFERENCES posts(id) NOT NULL,
    created_on timestamp NOT NULL DEFAULT current_timestamp
);
