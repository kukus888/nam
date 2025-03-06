package data

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GetHealthChecksAll(pool *pgxpool.Pool) (*[]Healthcheck, error) {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	rows, err := tx.Query(context.Background(), `SELECT * FROM Healthcheck h ORDER BY id ASC;`)
	if err != nil {
		return nil, err
	}
	res, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[Healthcheck])
	if err != nil {
		return nil, err
	}
	return &res, tx.Commit(context.Background())
}

// Inserts Healthcheck into Database. New ID is stored in the referenced Healthcheck struct.
// Does not roll back transaction, this is merely a facade for an insert statement
func (hc Healthcheck) DbInsert(pool *pgxpool.Pool) (*uint, error) {
	var id uint
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	err = tx.QueryRow(context.Background(), "INSERT INTO healthcheck (url, timeout, check_interval, expected_status) VALUES ($1, $2, $3, $4) RETURNING id", hc.Url, hc.Timeout, hc.CheckInterval, hc.ExpectedStatus).Scan(&id)
	if err != nil {
		return nil, err
	}
	return &id, tx.Commit(context.Background())
}

// Gets Healthcheck from Database by ID
func GetHealthCheckById(pool *pgxpool.Pool, id uint) (*Healthcheck, error) {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	rows, err := tx.Query(context.Background(), `SELECT * FROM Healthcheck h WHERE id = $1;`, id)
	if err != nil {
		return nil, err
	}
	res, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByNameLax[Healthcheck])
	if err != nil {
		return nil, err
	}
	return &res, tx.Commit(context.Background())
}
