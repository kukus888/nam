# Database
NAM uses PostgreSQL as its database, with connection made using [pgx](https://pkg.go.dev/github.com/jackc/pgx/v5) and [pgxpool](https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool) packages.

## Configuration
Connection to the database in configured in `config.yaml` file;

```yaml
postgres:
  dsn: "postgres://user:password@localhost:5432/postgres?application_name=nam&sslmode=disable"
```
## Migrations
Migrations are handled by [migrate](https://pkg.go.dev/github.com/golang-migrate/migrate/v4) package. For each version of database schemas, there should be two files in `src/layers/data/migrations` folder:
- `up.sql` - file with SQL commands to upgrade the database schema
- `down.sql` - file with SQL commands to downgrade the database schema

This migration folder gets automatically bundled together with the binary, and migrations are run on startup.