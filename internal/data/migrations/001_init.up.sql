-- Create Projects table
CREATE TABLE IF NOT EXISTS projects (
    id SERIAL PRIMARY KEY,
    name text NOT NULL,
    description TEXT NOT NULL,
    accent_color text NOT NULL DEFAULT '#000000',

	logo_url text NULL,

	app_key text NOT NULL,
    user_id uuid REFERENCES auth.users(id) NOT NULL,

    created_on timestamp NOT NULL DEFAULT current_timestamp,
    updated_on timestamp NULL
);

