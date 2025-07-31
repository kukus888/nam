CREATE TABLE IF NOT EXISTS "user" (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS "role" (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    description TEXT
);

CREATE TABLE IF NOT EXISTS "role_user_mapping" (
    user_id SERIAL REFERENCES "user" (id),
    role_id SERIAL REFERENCES "role" (id),
    PRIMARY KEY (user_id, role_id)
);

CREATE TABLE IF NOT EXISTS rolegroup (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    description TEXT
);

CREATE TABLE IF NOT EXISTS "rolegroup_role_mapping" (
    rolegroup_id SERIAL REFERENCES "rolegroup" (id),
    role_id SERIAL REFERENCES "role" (id),
    PRIMARY KEY (rolegroup_id, role_id)
);

CREATE TABLE IF NOT EXISTS "rolegroup_user_mapping" (
    rolegroup_id SERIAL REFERENCES "rolegroup" (id),
    user_id SERIAL REFERENCES "user" (id),
    PRIMARY KEY (rolegroup_id, user_id)
);