-- Add secrets table to support SOPS encrypted secrets
CREATE TABLE IF NOT EXISTS secrets (
    id BIGSERIAL PRIMARY KEY,
    type VARCHAR(50) NOT NULL, -- 'private_key', 'certificate', 'password', 'api_key', 'ssh_key', etc.
    name VARCHAR(255) NOT NULL UNIQUE, -- Human readable name/identifier
    description TEXT,
    data BYTEA NOT NULL, -- encrypted data
    metadata JSONB DEFAULT '{}', -- Additional metadata (key algorithm, cert expiry, etc.)
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by BIGINT REFERENCES "user"(id),
    updated_by BIGINT REFERENCES "user"(id)
);

-- Index for fast lookups by type and name
CREATE INDEX IF NOT EXISTS idx_secrets_type ON secrets(type);
CREATE INDEX IF NOT EXISTS idx_secrets_name ON secrets(name);
CREATE INDEX IF NOT EXISTS idx_secrets_type_name ON secrets(type, name);
