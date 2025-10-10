package data

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

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
	headersJSON, err := json.Marshal(hc.ReqHttpHeader)
	if err != nil {
		return nil, err
	}

	err = tx.QueryRow(context.Background(), `
        INSERT INTO healthcheck (
            name, description, url, method, headers, body, 
            timeout, check_interval, retry_count, retry_interval,
            expected_status, expected_response_body, response_validation,
            verify_ssl, auth_type, auth_credentials, protocol
        ) VALUES (
            $1, $2, $3, $4, $5, $6, 
            $7, $8, $9, $10,
            $11, $12, $13,
            $14, $15, $16,
            $17
        ) RETURNING id
    `,
		hc.Name, hc.Description, hc.ReqUrl, hc.ReqMethod, headersJSON, hc.ReqBody,
		hc.ReqTimeout, hc.CheckInterval, hc.RetryCount, hc.RetryInterval,
		hc.ExpectedStatus, hc.ExpectedResponseBody, hc.ResponseValidation,
		hc.VerifySSL, hc.AuthType, hc.AuthCredentials, hc.Protocol,
	).Scan(&hc.Id)

	if err != nil {
		return nil, err
	}

	return hc.Id, tx.Commit(context.Background())
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
	// Check if no rows were found
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // No healthcheck found with the given ID
	} else if err != nil {
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
	headersJSON, err := json.Marshal(hc.ReqHttpHeader)
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
            auth_type = $15,
            auth_credentials = $16,
            protocol = $17
        WHERE id = $18;
    `,
		hc.Name, hc.Description, hc.ReqUrl, hc.ReqMethod, headersJSON, hc.ReqBody,
		hc.ReqTimeout, hc.CheckInterval, hc.RetryCount, hc.RetryInterval,
		hc.ExpectedStatus, hc.ExpectedResponseBody, hc.ResponseValidation,
		hc.VerifySSL,
		hc.AuthType, hc.AuthCredentials, hc.Protocol,
		hc.Id,
	)

	if err != nil {
		return err
	}

	return tx.Commit(context.Background())
}

// DeleteHealthCheck deletes a healthcheck by its ID
func DeleteHealthCheckById(pool *pgxpool.Pool, id uint) error {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), `
		DELETE FROM healthcheck 
		WHERE id = $1;
	`, id)
	if err != nil {
		return err
	}

	return tx.Commit(context.Background())
}

type HealthcheckTarget struct {
	HealthcheckID         uint   `db:"hc_id"`
	ApplicationInstanceID uint   `db:"application_instance_id"`
	Hostname              string `db:"hostname"`
	Port                  uint   `db:"port"`
	Url                   string `db:"url"`
}

// Performs health check, returns the result
func (hc *Healthcheck) PerformCheck(url string, tlsConfig *tls.Config) (*HealthcheckResult, error) {
	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	httpClient := &http.Client{
		Timeout:   hc.ReqTimeout,
		Transport: tr,
	}
	result := &HealthcheckResult{
		HealthcheckID: *hc.Id,
		TimeStart:     time.Now(),
		IsSuccessful:  false,
	}
	req, err := http.NewRequest(hc.ReqMethod, url, nil)
	if err != nil {
		return result, err
	}
	req.Header = hc.ReqHttpHeader
	resp, err := httpClient.Do(req)
	result.TimeEnd = time.Now()
	result.ResTime = int(result.TimeEnd.Sub(result.TimeStart).Milliseconds())
	if err != nil {
		result.ErrorMessage = err.Error()
		return result, err
	} else {
		result.ResStatus = resp.StatusCode
		// Read response body
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			result.ErrorMessage = err.Error()
		} else {
			result.ResBody = string(bodyBytes)
		}
		// Close response body
		defer resp.Body.Close()
		// Check if the response status matches the expected status
		switch expression := hc.ResponseValidation; expression {
		case "none", "":
			if resp.StatusCode == hc.ExpectedStatus {
				result.IsSuccessful = true
			} else {
				result.ErrorMessage = "Unexpected status code: " + resp.Status
			}
		case "contains":
			if resp.StatusCode == hc.ExpectedStatus && hc.ExpectedResponseBody != "" {
				if !strings.Contains(result.ResBody, hc.ExpectedResponseBody) {
					result.ErrorMessage = "Response body does not contain expected content"
				} else {
					result.IsSuccessful = true
				}
			} else {
				result.ErrorMessage = "Unexpected status code: " + resp.Status
			}
		case "exact":
			if resp.StatusCode == hc.ExpectedStatus && hc.ExpectedResponseBody != "" {
				if result.ResBody != hc.ExpectedResponseBody {
					result.ErrorMessage = "Response body does not match expected content"
				} else {
					result.IsSuccessful = true
				}
			} else {
				result.ErrorMessage = "Unexpected status code: " + resp.Status
			}
		case "regex":
			if resp.StatusCode == hc.ExpectedStatus && hc.ExpectedResponseBody != "" {
				matched, err := regexp.MatchString(hc.ExpectedResponseBody, result.ResBody)
				if err != nil {
					result.ErrorMessage = "Error matching regex: " + err.Error()
				} else if !matched {
					result.ErrorMessage = "Response body does not match expected regex"
				} else {
					result.IsSuccessful = true
				}
			} else {
				result.ErrorMessage = "Unexpected status code: " + resp.Status
			}
		default:
			result.ErrorMessage = "Invalid response validation expression: " + expression
		}
	}
	return result, nil
}

// TODO: Func to clean up old healthcheck records, e.g., older than 30 days or with non existent healthchecks id
