CREATE TABLE application_definition_variable (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    value TEXT NOT NULL,
    description TEXT,
    application_definition_id INTEGER NOT NULL REFERENCES application_definition(id) ON DELETE CASCADE
);