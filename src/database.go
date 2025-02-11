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
		pgx.Identifier{"Server"},
		[]string{"Alias", "Hostname"},
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
func (Db *Database) QueryServerAll() ([]Server, error) {
	rows, err := Db.Pool.Query(context.Background(), `select * from "Server"`)
	if err != nil {
		return nil, err
	}
	servers, err := pgx.CollectRows(rows, pgx.RowToStructByName[Server])
	if err != nil {
		return nil, err
	}
	return servers, nil
}

// Gets Server instances from database by ID
func (Db *Database) QueryServerID(ID string) ([]Server, error) {
	rows, err := Db.Pool.Query(context.Background(), `SELECT * FROM "Server" WHERE "ID" = `+ID)
	if err != nil {
		return nil, err
	}
	servers, err := pgx.CollectRows(rows, pgx.RowToStructByName[Server])
	if err != nil {
		return nil, err
	}
	return servers, nil
}
