CREATE TABLE action_template (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    bash_script TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE action (
    id SERIAL PRIMARY KEY,
    action_template_id INTEGER NOT NULL REFERENCES action_template(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(50) DEFAULT 'pending', -- pending, running, completed, failed
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_by_user_id INTEGER NOT NULL -- References user table
);

CREATE TABLE action_execution (
    id SERIAL PRIMARY KEY,
    action_id INTEGER NOT NULL REFERENCES action(id) ON DELETE CASCADE,
    application_instance_id INTEGER NOT NULL REFERENCES application_instance(id) ON DELETE CASCADE,
    status VARCHAR(50) DEFAULT 'pending', -- pending, running, completed, failed
    output TEXT,
    error_output TEXT,
    exit_code INTEGER,
    started_at TIMESTAMP,
    completed_at TIMESTAMP
);
