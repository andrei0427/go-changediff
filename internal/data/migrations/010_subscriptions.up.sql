
CREATE TABLE IF NOT EXISTS subscriptions (
	id SERIAL PRIMARY KEY,
	subscription_start_date timestamp not null,
	is_annual boolean not null,
	price decimal not null,

	tier int not null,	
	session_id text null,
	success boolean not null,
	stopped boolean not null default(false),
	message text null,

	subscriber_id INT REFERENCES authors(id) not null,
	project_id INT REFERENCES projects(id) not null,

	created_on timestamp not null default(current_timestamp)
)