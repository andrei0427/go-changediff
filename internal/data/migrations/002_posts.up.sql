-- Create Authors table
CREATE TABLE IF NOT EXISTS authors (
    id SERIAL PRIMARY KEY,
	first_name text NOT NULL,
	last_name text NOT NULL,
	picture_url text NULL,

    user_id uuid REFERENCES auth.users(id) NOT NULL,
    project_id integer REFERENCES projects(id) NOT NULL,

    created_on timestamp NOT NULL DEFAULT current_timestamp,
    updated_on timestamp NULL
);

-- Create Posts table
CREATE TABLE IF NOT EXISTS posts (
    id SERIAL PRIMARY KEY,
    title text NOT NULL,
	body TEXT NOT NULL,
    published_on timestamp NOT NULL,

    author_id INT REFERENCES authors(id) NOT NULL,
    project_id integer REFERENCES projects(id) NOT NULL,

    created_on timestamp NOT NULL DEFAULT current_timestamp,
    updated_on timestamp NULL
);
