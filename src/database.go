package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	Pool *pgxpool.Pool
}

// Loads database connection
func (Db *Database) Start() {
	dsn := "postgres://postgres:heslo123@localhost:5432/postgres"
	p, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		panic(err)
	}
	Db.Pool = p
}

// Writes the s Server object into database.
func (Db *Database) InsertServer(s Server) error {
	rows := []Server{
		s,
	}
	copyCount, err := Db.Pool.CopyFrom(
		context.Background(),
		pgx.Identifier{"Servers"},
		[]string{"alias", "hostname"},
		pgx.CopyFromSlice(len(rows), func(i int) ([]any, error) {
			return []any{rows[i].Alias, rows[i].Hostname}, nil
		}),
	)
	if err != nil {
		return err // TODO: Error handling, unique values, etc
	}
	fmt.Printf("Inserted %d rows\n", copyCount)
	return nil
}

// Gets all Server instances from database
func (Db *Database) QueryAllServers() ([]Server, error) {
	rows, err := Db.Pool.Query(context.Background(), `select alias, hostname from "Servers" s`)
	if err != nil {
		return nil, err
	}
	strs, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, err
	}
	fmt.Printf(strs[0])
	s := Server{}
	rows.Scan(s.Hostname, s.Alias)
	return []Server{s}, nil
}
