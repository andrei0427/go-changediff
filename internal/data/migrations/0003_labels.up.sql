-- Create Labels table
CREATE TABLE IF NOT EXISTS labels (
    id SERIAL PRIMARY KEY,
    label VARCHAR(255) NOT NULL,
	color VARCHAR(255) NOT NULL,

    project_id integer REFERENCES projects(id) NOT NULL,
    created_on timestamp NOT NULL DEFAULT current_timestamp,
    updated_on timestamp NULL
);

ALTER TABLE posts ADD COLUMN label_id INT NULL;
ALTER TABLE posts ADD CONSTRAINT fk_label FOREIGN KEY (label_id) REFERENCES labels(id);
