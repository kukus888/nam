UPDATE app_settings
SET description = 'Warning threshold in percent of allocated DB size. When reached, the capacity guard runs smart cleanup (cleanup_healthcheck_results).'
WHERE key = 'db_capacity_warn_percent';

UPDATE app_settings
SET description = 'Critical threshold in percent of allocated DB size. If usage is still at or above this threshold after smart cleanup, oldest healthcheck results are hard-deleted.'
WHERE key = 'db_capacity_critical_percent';

DELETE FROM app_settings
WHERE key IN (
    'db_capacity_dry_run',
    'db_capacity_guard_enabled',
    'db_capacity_hard_delete_batch_rows',
    'db_capacity_max_delete_rows_per_run'
);