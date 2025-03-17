package data

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GetHealthChecksAll retrieves all healthchecks from the database
func GetHealthChecksAll(pool *pgxpool.Pool) (*[]Healthcheck, error) {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())

	rows, err := tx.Query(context.Background(), `
        SELECT * FROM healthcheck 
        ORDER BY id ASC;
    `)
	if err != nil {
		return nil, err
	}

	res, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[Healthcheck])
	if err != nil {
		return nil, err
	}

	return &res, tx.Commit(context.Background())
}

// DbInsert inserts a new healthcheck into the database
func (hc *Healthcheck) DbInsert(pool *pgxpool.Pool) (*uint, error) {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())

	// Convert headers to JSON
	headersJSON, err := json.Marshal(hc.Headers)
	if err != nil {
		return nil, err
	}

	err = tx.QueryRow(context.Background(), `
        INSERT INTO healthcheck (
            name, description, url, method, headers, body, 
            timeout, check_interval, retry_count, retry_interval,
            expected_status, expected_response_body, response_validation,
            verify_ssl, ssl_expiry_alert,
            auth_type, auth_credentials
        ) VALUES (
            $1, $2, $3, $4, $5, $6, 
            $7, $8, $9, $10,
            $11, $12, $13,
            $14, $15,
            $16, $17
        ) RETURNING id
    `,
		hc.Name, hc.Description, hc.Url, hc.Method, headersJSON, hc.Body,
		hc.Timeout, hc.CheckInterval, hc.RetryCount, hc.RetryInterval,
		hc.ExpectedStatus, hc.ExpectedResponseBody, hc.ResponseValidation,
		hc.VerifySSL, hc.SSLExpiryAlert,
		hc.AuthType, hc.AuthCredentials,
	).Scan(&hc.ID)

	if err != nil {
		return nil, err
	}

	return &hc.ID, tx.Commit(context.Background())
}

// GetHealthCheckById retrieves a healthcheck by its ID
func GetHealthCheckById(pool *pgxpool.Pool, id uint) (*Healthcheck, error) {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())

	rows, err := tx.Query(context.Background(), `
        SELECT * FROM healthcheck 
        WHERE id = $1;
    `, id)
	if err != nil {
		return nil, err
	}

	hc, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByNameLax[Healthcheck])
	if err != nil {
		return nil, err
	}

	// Parse headers from JSON
	/*var headers []http.Header
	if err := json.Unmarshal([]byte(hc.Headers), &headers); err != nil {
		return nil, err
	}
	hc.Headers = headers*/

	return &hc, tx.Commit(context.Background())
}

// UpdateHealthCheck updates an existing healthcheck
func (hc *Healthcheck) Update(pool *pgxpool.Pool) error {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	// Convert headers to JSON
	headersJSON, err := json.Marshal(hc.Headers)
	if err != nil {
		return err
	}

	_, err = tx.Exec(context.Background(), `
        UPDATE healthcheck SET
            name = $1,
            description = $2,
            url = $3,
            method = $4,
            headers = $5,
            body = $6,
            timeout = $7,
            check_interval = $8,
            retry_count = $9,
            retry_interval = $10,
            expected_status = $11,
            expected_response_body = $12,
            response_validation = $13,
            verify_ssl = $14,
            ssl_expiry_alert = $15,
            auth_type = $16,
            auth_credentials = $17
        WHERE id = $18
    `,
		hc.Name, hc.Description, hc.Url, hc.Method, headersJSON, hc.Body,
		hc.Timeout, hc.CheckInterval, hc.RetryCount, hc.RetryInterval,
		hc.ExpectedStatus, hc.ExpectedResponseBody, hc.ResponseValidation,
		hc.VerifySSL, hc.SSLExpiryAlert,
		hc.AuthType, hc.AuthCredentials,
		hc.ID,
	)

	if err != nil {
		return err
	}

	return tx.Commit(context.Background())
}
