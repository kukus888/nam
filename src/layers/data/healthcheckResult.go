package data

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Inserts a new healthcheck result into the database
// Returns the ID of the inserted result or an error
func (hr HealthcheckResult) DbInsert(pool *pgxpool.Pool) (*uint64, error) {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())

	err = tx.QueryRow(context.Background(), `
		INSERT INTO healthcheck_result (
			healthcheck_id, application_instance_id, is_successful,
			time_start, time_end, res_status, res_body,
			res_time, error_message
		) VALUES (
			$1, $2, $3, $4, $5, $6
		) RETURNING id;
	`, hr.HealthcheckID, hr.ApplicationInstanceID, hr.IsSuccessful,
		hr.TimeStart, hr.TimeEnd, hr.ResStatus, hr.ResBody,
		hr.ResTime, hr.ErrorMessage).Scan(&hr.ID)
	if err != nil {
		return nil, err
	}
	return &hr.ID, tx.Commit(context.Background())
}

func HealthcheckGetLatestResultByApplicationInstanceId(pool *pgxpool.Pool, id uint64) (*HealthcheckResult, error) {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())

	rows, err := tx.Query(context.Background(), `
		select * from healthcheck_results hr 
		where application_instance_id = $1
		order by hr.id desc
		limit 1
	`, id)
	if err != nil {
		return nil, err
	}
	res, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByNameLax[HealthcheckResult])
	if err != nil {
		return nil, err
	}
	return &res, tx.Commit(context.Background())
}

type ApplicationDefinitionHealthcheckResult struct {
	ApplicationInstanceID uint64    `json:"application_instance_id" db:"application_instance_id"`
	InstanceName          string    `json:"instance_name" db:"instance_name"`
	ServerHostname        string    `json:"server_hostname" db:"server_hostname"`
	ID                    uint64    `json:"id" db:"id"`
	HealthcheckID         uint      `json:"healthcheck_id" db:"healthcheck_id"`
	IsSuccessful          bool      `json:"is_successful" db:"is_successful"`
	TimeStart             time.Time `json:"time_start" db:"time_start"`
	TimeEnd               time.Time `json:"time_end" db:"time_end"`
	ResStatus             int       `json:"res_status" db:"res_status"`
	ResBody               string    `json:"res_body" db:"res_body"`
	ResTime               int       `json:"res_time" db:"res_time"` // in milliseconds
	ErrorMessage          string    `json:"error_message" db:"error_message"`
}

func HealthcheckGetLatestResultByApplicationDefinitionId(pool *pgxpool.Pool, id uint64) (*[]ApplicationDefinitionHealthcheckResult, error) {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())

	rows, err := tx.Query(context.Background(), `
		SELECT
		  ai.id AS application_instance_id,
		  ai.name AS instance_name,
		  s.hostname AS server_hostname,
		  hcr.id AS id,
		  hcr.healthcheck_id AS healthcheck_id,
		  hcr.is_successful AS is_successful,
		  hcr.time_start AS time_start,
		  hcr.time_end AS time_end,
		  hcr.res_status AS res_status,
		  hcr.res_body AS res_body,
		  hcr.res_time AS res_time,
		  hcr.error_message AS error_message
		FROM application_instance ai
		JOIN server s ON ai.server_id = s.id
		LEFT JOIN LATERAL (
		    SELECT *
		    FROM healthcheck_results hcr
		    WHERE hcr.application_instance_id = ai.id
		    ORDER BY hcr.time_end DESC
		    LIMIT 1
		) hcr ON TRUE
		WHERE ai.application_definition_id = $1
		ORDER BY ai.id desc;
	`, id)
	if err != nil {
		return nil, err
	}

	res, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[ApplicationDefinitionHealthcheckResult])
	if err != nil {
		return nil, err
	}

	return &res, tx.Commit(context.Background())
}

// Inserts a healthcheck results into the database
// Returns an error, and assigns IDs to the corresponding healthcheck results
func HealthcheckResultBatchInsert(pool *pgxpool.Pool, hrs *[]HealthcheckResult) error {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())
	for hrId := range *hrs {
		hr := (*hrs)[hrId]
		err = tx.QueryRow(context.Background(), `
		INSERT INTO healthcheck_results (
			healthcheck_id, application_instance_id, is_successful,
			time_start, time_end, res_status, res_body,
			res_time, error_message
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		) RETURNING id;
		`, hr.HealthcheckID, hr.ApplicationInstanceID, hr.IsSuccessful,
			hr.TimeStart, hr.TimeEnd, hr.ResStatus, hr.ResBody,
			hr.ResTime, hr.ErrorMessage).Scan(&hr.ID)
		if err != nil {
			return err
		}
		(*hrs)[hrId].ID = hr.ID // Assign the ID back to the slice
	}
	return tx.Commit(context.Background())
}

// Gets all healthcheck results from the database
// Returns a slice of HealthcheckResult or an error
func GetHealthcheckResultsAll(pool *pgxpool.Pool) (*[]HealthcheckResult, error) {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())

	rows, err := tx.Query(context.Background(), `
		SELECT * FROM healthcheck_result 
		ORDER BY id ASC;
	`)
	if err != nil {
		return nil, err
	}

	res, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[HealthcheckResult])
	if err != nil {
		return nil, err
	}

	return &res, tx.Commit(context.Background())
}
