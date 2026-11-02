# NAM Testing Coverage and Gap Analysis

Last updated: 2026-11-02

## Purpose

This document tracks:

- Which components currently have automated tests
- Which components do not
- What test scenarios are covered now
- What tests should be written next (prioritized)

## Current Test Inventory

### Unit tests

- `src/layers/service/cryptoService_test.go`
  - `Encrypt()` / `Decrypt()` roundtrip
  - Random nonce behavior
  - Short ciphertext rejection
  - Wrong-key decrypt failure
  - `EncryptSecret()` and `DecryptDAO()` conversion flow

- `src/layers/service/jwt_test.go`
  - `GenerateToken()` claim content and expiry window
  - `RegenerateToken()` expiry extension

- `src/layers/data/password_test.go`
  - `HashPassword()` and `VerifyPassword()` happy path
  - Salt behavior (same password -> different hashes)
  - Invalid password and malformed hash failures
  - `User.HasRole()` hierarchy behavior

### Integration tests (real PostgreSQL via Testcontainers)

- `src/layers/data/integration_test_helpers_test.go`
  - Test DB container setup/teardown
  - Migration application

- `src/layers/data/user_secrets_integration_test.go`
  - User CRUD lifecycle:
    - `CreateUser()`
    - `GetUserById()` / `GetUserByUsername()`
    - `UpdateWithoutPassword()`
    - `UpdatePassword()`
    - `Delete()`
  - Secrets CRUD lifecycle:
    - `SecretDAO.DbInsert()`
    - `GetSecretById()` / `GetSecretByName()` / `GetAllSecrets()` / `GetSshSecrets()`
    - `UpdateSecret()`
    - `DeleteSecret()`

## Coverage Matrix by Component

Status legend:

- `Covered`: meaningful automated tests exist
- `Partial`: some behavior covered, key paths still missing
- `Not Covered`: no automated tests currently

### Service layer

- `src/layers/service/cryptoService.go` -> Covered
- `src/layers/service/jwt.go` -> Covered
- `src/layers/service/jwtKeyProvider.go` -> Partial (indirectly exercised from jwt tests)
- `src/layers/service/secretsService.go` -> Not Covered
- `src/layers/service/healthcheck.go` -> Not Covered
- `src/layers/service/dashboardCacheService.go` -> Not Covered
- `src/layers/service/offeringDashboardCacheService.go` -> Not Covered
- `src/layers/service/timerService.go` -> Not Covered
- `src/layers/service/databaseCleanupTimer.go` -> Not Covered
- `src/layers/service/databaseFlushHcResultsTimer.go` -> Not Covered
- `src/layers/service/databaseCapacityGuardTimer.go` -> Not Covered
- `src/layers/service/services.go` -> Not Covered

### Data layer

- `src/layers/data/password.go` -> Covered
- `src/layers/data/user.go` -> Partial (CRUD paths covered; validation and error edge cases still missing)
- `src/layers/data/secrets.go` -> Partial (CRUD paths covered; error edge cases still missing)
- `src/layers/data/migration.go` -> Partial (covered indirectly by integration bootstrap)
- `src/layers/data/role.go` -> Not Covered
- `src/layers/data/server.go` -> Not Covered
- `src/layers/data/healthchecks.go` -> Not Covered
- `src/layers/data/healthcheckResult.go` -> Not Covered
- `src/layers/data/dashboard.go` -> Not Covered
- `src/layers/data/offeringDashboard.go` -> Not Covered
- `src/layers/data/actions.go` -> Not Covered
- `src/layers/data/applicationDefinition.go` -> Not Covered
- `src/layers/data/applicationDefinitionVariables.go` -> Not Covered
- `src/layers/data/applicationInstance.go` -> Not Covered
- `src/layers/data/applicationInstanceVariables.go` -> Not Covered
- `src/layers/data/serviceOffering.go` -> Not Covered
- `src/layers/data/topologyNode.go` -> Not Covered
- `src/layers/data/appSettings.go` -> Not Covered
- `src/layers/data/database.go` -> Not Covered
- `src/layers/data/types.go` -> Not Covered
- `src/layers/data/models.go` -> Partial (used heavily, but conversion/parser helpers not directly tested)

### Handler layer (HTTP)

- `src/layers/handler/login.go` -> Not Covered
- `src/layers/handler/handler.go` -> Not Covered
- `src/layers/handler/actions.go` -> Not Covered
- `src/layers/handler/applications.go` -> Not Covered
- `src/layers/handler/application_instances.go` -> Not Covered
- `src/layers/handler/dashboardHandler.go` -> Not Covered
- `src/layers/handler/healthchecks.go` -> Not Covered
- `src/layers/handler/offerings.go` -> Not Covered
- `src/layers/handler/pageProfile.go` -> Not Covered
- `src/layers/handler/pageSecretsHandler.go` -> Not Covered
- `src/layers/handler/pageSettingsHandler.go` -> Not Covered
- `src/layers/handler/pages.go` -> Not Covered
- `src/layers/handler/server.go` -> Not Covered
- `src/layers/handler/api/rest/v1/*` -> Not Covered
- `src/layers/handler/htmx/*` -> Not Covered

### Application bootstrap and routing

- `src/main.go` -> Not Covered
- `src/webserver.go` -> Not Covered
- `src/configuration.go` -> Not Covered
- `src/database.go` -> Not Covered
- `src/keys.go` -> Not Covered
- `src/rundeck.go` -> Not Covered

## Test Execution Commands

From `src/`:

- Run all default tests:

  `go test ./...`

- Run integration tests only (DB/Testcontainers):

  `go test -tags=integration ./layers/data -run 'Test(User|Secrets)CRUDIntegration' -count=1`

## CI Coverage

- `.github/workflows/tests.yml`
  - `unit` job: runs `go test ./...`
  - `integration` job: runs PostgreSQL integration tests with `-tags=integration`

## Recommended Tests to Write Next

### P0 (highest value / risk reduction)

1. Login + auth middleware flow
   - Files:
     - `src/layers/handler/login.go`
     - `src/webserver.go` (AuthMiddleware/RequireRole)
   - Tests:
     - Valid login returns token/cookie
     - Invalid credentials rejected
     - Missing/invalid token rejected
     - Role enforcement for Admin/Operator/Viewer

2. Secrets service business rules
   - File: `src/layers/service/secretsService.go`
   - Tests:
     - Encrypt before store / decrypt after read
     - Metadata and audit fields handling
     - Error propagation from DAO/crypto layer

3. User + role management edge cases
   - Files:
     - `src/layers/data/user.go`
     - `src/layers/data/role.go`
   - Tests:
     - Duplicate username/email conflicts
     - Required-field validation failures
     - Role assignment/readback consistency

### P1 (critical API and data stability)

1. REST endpoint integration tests (real DB)
   - Files:
     - `src/layers/handler/api/rest/v1/user.go`
     - `src/layers/handler/api/rest/v1/secrets.go`
     - `src/layers/handler/api/rest/v1/server.go`
   - Tests:
     - CRUD success paths
     - Validation failures (bad payloads)
     - Auth and RBAC protections

2. Migration reliability checks
   - File: `src/layers/data/migration.go`
   - Tests:
     - `AutoMigrate()` idempotency
     - Fresh DB -> latest schema success
     - Optional up/down smoke cycle in ephemeral DB

3. Configuration parsing and failure behavior
   - File: `src/configuration.go`
   - Tests:
     - Valid config loads expected defaults
     - Missing required fields fail with clear errors
     - Invalid YAML fails deterministically

### P2 (background systems and performance-sensitive areas)

1. Healthcheck service behavior
   - File: `src/layers/service/healthcheck.go`
   - Tests:
     - Observer sync logic
     - Listener reconnect/error behavior
     - Goroutine lifecycle and shutdown safety

2. Timer services
   - Files:
     - `src/layers/service/timerService.go`
     - `src/layers/service/databaseCleanupTimer.go`
     - `src/layers/service/databaseFlushHcResultsTimer.go`
   - Tests:
     - Enable/disable/run behavior
     - No-op behavior when disabled
     - Error logging and retry expectations

3. Dashboard query layer
   - Files:
     - `src/layers/data/dashboard.go`
     - `src/layers/data/offeringDashboard.go`
   - Tests:
     - Query result shape and ordering
     - Empty dataset behavior

## Gaps to Keep in Mind

- No handler-level HTTP tests exist yet.
- No full E2E workflow test (boot app + HTTP + DB assertions) exists yet.
- Most data-layer modules are still uncovered.
- Healthcheck/timer concurrency behavior is currently untested.

## Suggested Milestone Plan

1. Add HTTP integration tests for login + users + secrets.
2. Add one E2E smoke workflow:
   - bootstrap first admin
   - login
   - create/read/update/delete one core entity
3. Expand coverage to role/server/application definition APIs.
4. Add healthcheck/timer stability tests and race-detector checks.
