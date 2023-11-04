-- Update Reaction tables
ALTER TABLE post_reactions ADD COLUMN user_data JSON;
ALTER TABLE post_reactions ADD COLUMN user_id varchar(255);
ALTER TABLE post_reactions ADD COLUMN user_name varchar(255);
ALTER TABLE post_reactions ADD COLUMN user_email varchar(255);
ALTER TABLE post_reactions ADD COLUMN user_role varchar(255);

