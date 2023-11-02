
CREATE TABLE IF NOT EXISTS payments (
	id SERIAL PRIMARY KEY,
	subscription_start_date timestamp not null,
	is_annual boolean not null,
	price decimal not null,

	tier int not null,	
	session_id varchar(255) not null,
	success boolean not null,
	message text not null,

	subscriber_id INT REFERENCES users(id) not null,
	project_id INT REFERENCES projects(id) not null,

	created_on timestamp not null default(current_timestamp)
)