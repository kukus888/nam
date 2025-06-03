package data

import (
	"context"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func (s Server) TableName() string {
	return "server"
}

func (s Server) ApiName() string {
	return "server"
}

// Inserts Server into Database. New Id is stored in the referenced Server struct.
// Does not roll back transaction, this is merely a facade for an insert statement
func (s Server) DbInsert(tx pgx.Tx) (*uint, error) {
	var id uint
	err := tx.QueryRow(context.Background(), "INSERT INTO server (alias, hostname) VALUES ($1, $2) RETURNING id", s.Alias, s.Hostname).Scan(&id)
	if err != nil {
		tx.Rollback(context.Background())
		return nil, err
	}
	return &id, tx.Commit(context.Background())
}

// Deletes specified Server. Checks for dependent ApplicationInstances
func (s Server) Delete(pool *pgxpool.Pool) (*int, error) {
	var affectedRows = 0
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	// Check if server exists
	var server Server
	err = pgxscan.Get(context.Background(), tx, &server, "SELECT * FROM server WHERE id = $1", s.Id)
	if err != nil {
		tx.Rollback(context.Background())
		return nil, err
	}
	if server.Id == 0 {
		tx.Rollback(context.Background())
		return nil, nil
	}
	// Check for dependent ApplicationInstances
	var instances []ApplicationInstance
	err = pgxscan.Select(context.Background(), tx, &instances, "SELECT * FROM application_instance WHERE server_id = $1", s.Id)
	if err != nil {
		tx.Rollback(context.Background())
		return nil, err
	}
	if len(instances) > 0 {
		// Remove dependent ApplicationInstances
		for _, instance := range instances {
			err := DeleteApplicationInstanceById(pool, uint64(instance.Id))
			if err != nil {
				tx.Rollback(context.Background())
				return nil, err
			}
		}
	}
	// Remove Server
	com, err := tx.Exec(context.Background(), `DELETE FROM "server" s WHERE s.id = $1`, s.Id)
	if err != nil {
		tx.Rollback(context.Background())
		return nil, err
	}
	affectedRows += int(com.RowsAffected())
	return &affectedRows, tx.Commit(context.Background())
}

func GetServerAll(pool *pgxpool.Pool) (*[]Server, error) {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
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
	return &res, nil
}

func (server *Server) Update(pool *pgxpool.Pool) error {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return err
	}
	rows, err := tx.Query(context.Background(), `update "server" s set alias = $2, hostname = $3 where s.id = $1 returning *`, server.Id, server.Alias, server.Hostname)
	if err != nil {
		tx.Rollback(context.Background())
		return err
	}
	rows.Scan(&server)
	return tx.Commit(context.Background())
}
