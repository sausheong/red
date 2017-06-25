CREATE TABLE users (
	id serial primary key,
	username varchar(255) unique not null,
	email varchar(255) unique not null,
	name varchar(255),
	last_login timestamp default CURRENT_TIMESTAMP,
	is_admin boolean default false,
	github_token varchar(255)
);

CREATE TABLE sessions (
	id serial primary key,
	uuid varchar(255) not null,
	user_id int references users(id),
	email varchar(255) not null,
	date_created timestamp default CURRENT_TIMESTAMP
);

CREATE TABLE settings (
	queue varchar(255),
	repo varchar(255)
);

