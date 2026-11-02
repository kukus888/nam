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
    'db_capacity_hard_delete_batch_rows',
    '1000',
    'int',
    '^[0-9]+$',
    'Maximum oldest healthcheck rows to delete in one SQL batch.',
    100,
    100000,
    NULL,
    TRUE
),
(
    'db_capacity_max_delete_rows_per_run',
    '20000',
    'int',
    '^[0-9]+$',
    'Maximum oldest healthcheck rows to delete per timer run.',
    1000,
    1000000,
    NULL,
    TRUE
),
(
    'db_capacity_guard_enabled',
    'false',
    'bool',
    '^(true|false)$',
    'Enable or disable periodic DB capacity guard timer.',
    NULL,
    NULL,
    ARRAY['true', 'false'],
    TRUE
),
(
    'db_capacity_dry_run',
    'true',
    'bool',
    '^(true|false)$',
    'When true, capacity guard logs actions without deleting data.',
    NULL,
    NULL,
    ARRAY['true', 'false'],
    TRUE
)
ON CONFLICT (key) DO NOTHING;