-- Create Posts table
CREATE TABLE IF NOT EXISTS posts (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
	body TEXT NOT NULL,
    published_on timestamp NOT NULL,

    author_id uuid REFERENCES auth.users(id) NOT NULL,
    project_id integer REFERENCES projects(id) NOT NULL,

    created_on timestamp NOT NULL DEFAULT current_timestamp,
    updated_on timestamp NULL
);
