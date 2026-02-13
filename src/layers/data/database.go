package data

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	Pool *pgxpool.Pool
}

// Initializes new pgx database connection with provided connection string
func NewDatabase(dsn string) (*Database, error) {
	p, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, err
	}

	return &Database{Pool: p}, nil
}

// Returns size of database, in bytes, or an error if there was one
func (db *Database) GetDatabaseSize() (*int64, error) {
	var dbSize int64
	err := db.Pool.QueryRow(context.Background(), "SELECT pg_database_size('$1')", db.Pool.Config().ConnConfig.Database).Scan(&dbSize)
	if err != nil {
		return nil, err
	}
	return &dbSize, nil
}

// TableSize represents the size of a table, including the size of the table itself, its indexes, and the total size (table + indexes)
type TableSize struct {
	TableName   string `db:"table_name"`
	TableSize   string `db:"table_size"`
	IndexesSize string `db:"indexes_size"`
	TotalSize   string `db:"total_size"`
}

// Gets the table sizes
func GetTableSizes(pool *pgxpool.Pool) (*[]TableSize, error) {
	// Source - https://stackoverflow.com/a/2596678
	// Posted by aib, modified by community. See post 'Timeline' for change history
	// Retrieved 2026-02-10, License - CC BY-SA 3.0
	rows, err := pool.Query(context.Background(), `
	SELECT
    	table_name,
    	pg_size_pretty(table_size) AS table_size,
	    pg_size_pretty(indexes_size) AS indexes_size,
	    pg_size_pretty(total_size) AS total_size
	FROM (
	    SELECT
	        table_name,
	        pg_table_size(table_name) AS table_size,
	        pg_indexes_size(table_name) AS indexes_size,
	        pg_total_relation_size(table_name) AS total_size
	    FROM (
	        SELECT (table_schema || '.' || table_name) AS table_name
	        FROM information_schema.tables
	    ) AS all_tables
	    ORDER BY total_size DESC
	) AS pretty_sizes;
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tableSizes := make([]TableSize, 0)
	for rows.Next() {
		var ts TableSize
		if err := rows.Scan(&ts.TableName, &ts.TableSize, &ts.IndexesSize, &ts.TotalSize); err != nil {
			return nil, err
		}
		tableSizes = append(tableSizes, ts)
	}
	return &tableSizes, nil
}

// CleanUpDatabase performs routine cleanup tasks on the database, squashing the healthcheck_results table. It returns a message indicating the result of the cleanup operation or an error if one occurred.
func CleanUpDatabase(pool *pgxpool.Pool) (string, error) {
	var recordsBefore, recordsAfter, recordsDeleted int64

	err := pool.QueryRow(context.Background(), "SELECT * FROM cleanup_healthcheck_results();").Scan(&recordsBefore, &recordsAfter, &recordsDeleted)
	if err != nil {
		return "", err
	}

	result := fmt.Sprintf("Database cleanup completed: %d records before, %d records after, %d records deleted",
		recordsBefore, recordsAfter, recordsDeleted)

	return result, nil
}

// FlushHealthCheckResults deletes all records from the healthcheck_results table and returns a message indicating how many records were deleted, or an error if one occurred.
func FlushHealthCheckResults(pool *pgxpool.Pool) (string, error) {
	var recordsDeleted int64

	err := pool.QueryRow(context.Background(), `
	WITH deleted AS (
		DELETE FROM healthcheck_results
		RETURNING 1
	)
	SELECT COUNT(*) FROM deleted;
	`).Scan(&recordsDeleted)
	if err != nil {
		return "", err
	}

	result := fmt.Sprintf("Flushed healthcheck_results table: %d records deleted", recordsDeleted)
	return result, nil
}
