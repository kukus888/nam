package data

import (
	"context"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ServerDAO struct {
	ID       uint   `json:"id"`
	Alias    string `json:"alias"`
	Hostname string `json:"hostname"`
}

func (s ServerDAO) TableName() string {
	return "server"
}

func (s ServerDAO) ApiName() string {
	return "server"
}

// Inserts Server into Database. New ID is stored in the referenced Server struct.
// Does not roll back transaction, this is merely a facade for an insert statement
func (s ServerDAO) DbInsert(tx pgx.Tx) (*uint, error) {
	var id uint
	err := tx.QueryRow(context.Background(), "INSERT INTO server (alias, hostname) VALUES ($1, $2) RETURNING id", s.Alias, s.Hostname).Scan(&id)
	return &id, err
}

// Deletes specified ServerDAO. Checks for dependent ApplicationInstances
func (s ServerDAO) Delete(pool *pgxpool.Pool) (*int, error) {
	var affectedRows = 0
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	// Check if server exists
	var server ServerDAO
	err = pgxscan.Get(context.Background(), tx, &server, "SELECT * FROM server WHERE id = $1", s.ID)
	if err != nil {
		tx.Rollback(context.Background())
		return nil, err
	}
	if server.ID == 0 {
		tx.Rollback(context.Background())
		return nil, nil
	}
	// Check for dependent ApplicationInstances
	var instances []ApplicationInstanceDAO
	err = pgxscan.Select(context.Background(), tx, &instances, "SELECT * FROM application_instance WHERE server_id = $1", s.ID)
	if err != nil {
		tx.Rollback(context.Background())
		return nil, err
	}
	if len(instances) > 0 {
		// Remove dependent ApplicationInstances
		for _, instance := range instances {
			_, err = instance.Delete(pool)
			if err != nil {
				tx.Rollback(context.Background())
				return nil, err
			}
		}
	}
	// Remove Server
	com, err := tx.Exec(context.Background(), "DELETE FROM server WHERE id = $1", s.ID)
	if err != nil {
		tx.Rollback(context.Background())
		return nil, err
	}
	affectedRows += int(com.RowsAffected())
	return &affectedRows, nil
}
