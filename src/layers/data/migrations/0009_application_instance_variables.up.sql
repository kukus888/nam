CREATE TABLE application_instance_variable (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    value TEXT NOT NULL,
    description TEXT,
    application_instance_id INTEGER NOT NULL REFERENCES application_instance(id) ON DELETE CASCADE
);