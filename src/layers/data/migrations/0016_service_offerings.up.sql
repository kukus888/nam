CREATE TABLE IF NOT EXISTS service_offering (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    created_by_user_id INTEGER REFERENCES "user" (id) ON DELETE SET NULL,
    updated_by_user_id INTEGER REFERENCES "user" (id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS service_offering_application_definition (
    service_offering_id INTEGER NOT NULL REFERENCES service_offering (id) ON DELETE CASCADE,
    application_definition_id INTEGER NOT NULL REFERENCES application_definition (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (service_offering_id, application_definition_id)
);

CREATE INDEX IF NOT EXISTS idx_service_offering_name
    ON service_offering (name);

CREATE INDEX IF NOT EXISTS idx_service_offering_app_def_by_offering
    ON service_offering_application_definition (service_offering_id);

CREATE INDEX IF NOT EXISTS idx_service_offering_app_def_by_application
    ON service_offering_application_definition (application_definition_id);

CREATE TABLE IF NOT EXISTS service_offering_audit (
    id BIGSERIAL PRIMARY KEY,
    service_offering_id INTEGER REFERENCES service_offering (id) ON DELETE SET NULL,
    action VARCHAR(120) NOT NULL,
    details JSONB,
    created_by_user_id INTEGER REFERENCES "user" (id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_service_offering_audit_created_at
    ON service_offering_audit (created_at DESC);

CREATE INDEX IF NOT EXISTS idx_service_offering_audit_offering_id
    ON service_offering_audit (service_offering_id);