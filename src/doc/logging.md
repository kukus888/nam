# Logging Guide

This document defines how logging should work in NAM.

## Goals

- Keep all runtime logs in one structured pipeline.
- Avoid direct stdout and stderr output in runtime code paths.
- Keep logs machine-readable and consistent.
- Make production logging configurable from `config.yaml`.

## Canonical Logging Stack

- Structured logger: `log/slog` with JSON handler.
- Sink: rotating file writer (`lumberjack`) configured by `logging` settings.
- Access logs: Gin request logs written to the same sink.
- Recovery logs: panic recovery events written through slog with request context.

## Configuration

Logging is configured under `logging` in `config.yaml`:

- `level`: `debug`, `info`, `warn`, `error`
- `filepath`: log file location
- `maxsize`: max file size in MB before rotation
- `maxbackups`: number of rotated files to keep
- `maxage`: maximum age in days
- `compress`: whether rotated files are compressed

## Rules For Contributors

1. Use `*slog.Logger` for application logging.
2. Pass logger dependencies explicitly into constructors/controllers/services.
3. Use structured fields, not formatted strings, for runtime context.
4. Do not use `fmt.Print`, `fmt.Printf`, `fmt.Println`, or `println` in runtime server paths.
5. Keep log levels meaningful:
   - `debug`: verbose diagnostic detail
   - `info`: normal lifecycle milestones
   - `warn`: degraded behavior with fallback or partial failure
   - `error`: failed operation

## Field Conventions

Use stable snake_case field names where possible:

- `component` (webserver, healthcheck, action_controller)
- `error`
- `method`
- `path`
- `client_ip`
- `status_code`
- `response_time`
- entity IDs such as `action_id`, `instance_id`, `user_id`

## Gin-Specific Guidance

- Access logging should use `gin.LoggerWithConfig` and write to the same configured writer.
- Panic recovery should use custom Gin recovery that logs with slog and includes request metadata and stack trace.
- Avoid default recovery output streams if they bypass configured sink policy.

## Examples

### Good

```go
logger.Info("Starting web server", "component", "webserver", "address", cfg.WebServer.Address)
logger.Warn("Failed to create action execution", "action_id", actionID, "instance_id", instanceID, "error", err)
```

### Avoid

```go
fmt.Printf("failed to create execution for instance %d: %v\n", instanceID, err)
fmt.Println("HealthcheckService status:", status)
```

## Migration Notes

When refactoring legacy code:

1. Replace `fmt.Print*` calls with logger calls at the appropriate level.
2. Preserve business behavior while changing only observability behavior.
3. Keep error values in structured `error` fields.
4. Verify output appears in configured log files and respects level filtering.

## Verification Checklist

- Start app and verify startup logs, app logs, access logs, and recovery logs go to configured file.
- Confirm runtime server process does not emit repetitive stdout lines.
- Validate debug and info suppression when level is set to `warn` or `error`.
- Force log rotation and verify rotated files are created and retained per config.
