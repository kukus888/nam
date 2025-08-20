CREATE TABLE IF NOT EXISTS "role" (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    color VARCHAR(7) NOT NULL,
    description TEXT
);

CREATE TABLE IF NOT EXISTS "user" (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role_id SERIAL REFERENCES "role" (id)
);

insert into "role" (name, color, description)
values
	('Admin', 'purple', 'Access to everything.'),
	('Operator', 'blue', 'Access to write and edit, but not all settings.'),
	('Viewer', 'green', 'Access to view only.');
