-- Create Projects table
CREATE TABLE IF NOT EXISTS projects (
    id SERIAL PRIMARY KEY,
    project_name VARCHAR(100) NOT NULL,
	logo_url VARCHAR(200) NULL,

	app_key VARCHAR(256) NOT NULL,
    user_id uuid REFERENCES auth.users(id) NOT NULL,

    created_on timestamp NOT NULL DEFAULT current_timestamp,
    updated_on timestamp NULL
);

