CREATE TABLE IF NOT EXISTS app_settings (
    id BIGSERIAL PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    value TEXT NOT NULL,
    value_type TEXT NOT NULL,
    validation_regex TEXT NULL,
    description TEXT NOT NULL,
    min_value NUMERIC NULL,
    max_value NUMERIC NULL,
    allowed_values TEXT[] NULL,
    is_editable BOOLEAN NOT NULL DEFAULT TRUE,
    updated_by BIGINT NULL REFERENCES "user" (id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_app_settings_value_type
        CHECK (value_type IN ('int', 'float', 'bool', 'duration', 'size_bytes', 'string')),
    CONSTRAINT chk_app_settings_min_max
        CHECK (min_value IS NULL OR max_value IS NULL OR min_value <= max_value)
);

INSERT INTO app_settings (
    key,
    value,
    value_type,
    validation_regex,
    description,
    min_value,
    max_value,
    allowed_values,
    is_editable
) VALUES
(
    'db_capacity_allocated_bytes',
    '2147483648',
    'size_bytes',
    '^[0-9]+$',
    'Total allocated database size in bytes.',
    1048576,
    NULL,
    NULL,
    TRUE
),
(
    'db_capacity_warn_percent',
    '80',
    'float',
    '^[0-9]+(\\.[0-9]+)?$',
    'Warning threshold in percent of allocated DB size. When reached, the capacity guard runs smart cleanup (cleanup_healthcheck_results).',
    1,
    99.9,
    NULL,
    TRUE
),
(
    'db_capacity_critical_percent',
    '90',
    'float',
    '^[0-9]+(\\.[0-9]+)?$',
    'Critical threshold in percent of allocated DB size. If usage is still at or above this threshold after smart cleanup, oldest healthcheck results are hard-deleted.',
    1,
    99.9,
    NULL,
    TRUE
),
(
    'db_capacity_check_interval',
    '15m',
    'duration',
    '^[0-9]+(ms|s|m|h)$',
    'Capacity guard check interval. Supports ms, s, m, h.',
    NULL,
    NULL,
    NULL,
    TRUE
)
ON CONFLICT (key) DO NOTHING;