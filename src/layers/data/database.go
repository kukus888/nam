package data

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	Pool *pgxpool.Pool
}

// TODO: Impl DB context

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

// Returns size of table, in bytes, or an error if there was one
func (db *Database) GetTableSize(tableName string) (*int64, error) {
	var tableSize int64
	err := db.Pool.QueryRow(context.Background(), "SELECT pg_relation_size('$1')", tableName).Scan(&tableSize)
	if err != nil {
		return nil, err
	}
	return &tableSize, nil
}
