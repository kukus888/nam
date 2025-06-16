package data

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Inserts Server into Database. New Id is stored in the referenced Server struct.
// Does not roll back transaction, this is merely a facade for an insert statement
func (s Server) DbInsert(pool *pgxpool.Pool) (*uint, error) {
	var id uint
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())
	err = tx.QueryRow(context.Background(), "INSERT INTO server (alias, hostname) VALUES ($1, $2) RETURNING id", s.Alias, s.Hostname).Scan(&id)
	if err != nil {
		return nil, err
	}
	return &id, tx.Commit(context.Background())
}

func ServerDeleteById(pool *pgxpool.Pool, id uint) error {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())
	// Delete ApplicationInstances first
	instances, err := tx.Query(context.Background(), `
		select * from application_instance ai 
		where ai.server_id = $1
	`, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil // No results found, done!
	} else if err != nil {
		return err
	}
	ais, err := pgx.CollectRows(instances, pgx.RowToStructByNameLax[ApplicationInstance])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil // No results found
	} else if err != nil {
		return err
	}
	if len(ais) > 0 {
		for _, ai := range ais {
			err = DeleteApplicationInstanceById(pool, uint64(ai.Id))
			if err != nil {
				return err
			}
		}
	}
	// Delete the server
	_, err = tx.Exec(context.Background(), `
		delete from server s 
		where s.id = $1
	`, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil // No results found, done!
	} else if err != nil {
		return err
	}
	return tx.Commit(context.Background())
}

func GetServerAll(pool *pgxpool.Pool) (*[]Server, error) {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	defer tx.Rollback(context.Background())
	if err != nil {
		return nil, err
	}
	rows, err := tx.Query(context.Background(), `SELECT s.id as server_id, s.alias as server_alias, s.hostname as server_hostname FROM Server s ORDER BY id ASC;`)
	if err != nil {
		return nil, err
	}
	res, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[Server])
	if err != nil {
		return nil, err
	}
	return &res, tx.Commit(context.Background())
}

func GetServerById(pool *pgxpool.Pool, id uint) (*Server, error) {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	defer tx.Rollback(context.Background())
	if err != nil {
		return nil, err
	}
	rows, err := tx.Query(context.Background(), `SELECT s.id as server_id, s.alias as server_alias, s.hostname as server_hostname FROM Server s WHERE s.id = $1`, id)
	if err != nil {
		return nil, err
	}
	res, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByNameLax[Server])
	if err != nil {
		return nil, err
	}
	return &res, tx.Commit(context.Background())
}

func (server *Server) Update(pool *pgxpool.Pool) error {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())
	if server.Id == 0 {
		return errors.New("id is required for db update, got id = 0")
	}
	_, err = tx.Exec(context.Background(), `update server set alias = $2, hostname = $3 where id = $1`, server.Id, server.Alias, server.Hostname)
	if err != nil {
		return err
	}
	return tx.Commit(context.Background())
}
