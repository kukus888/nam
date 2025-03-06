package data

import (
	"context"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ApplicationDefinitionDAO represents the definition of an application and its general properties
type ApplicationDefinitionDAO struct {
	ID            uint   `json:"id" db:"id"`
	Name          string `json:"name" db:"name"`
	Port          int    `json:"port" db:"port"`
	Type          string `json:"type" db:"type"`
	HealthcheckId *uint  `json:"healthcheck_id" db:"healthcheck_id"`
}

func (s ApplicationDefinitionDAO) TableName() string {
	return "application_definition"
}

func (s ApplicationDefinitionDAO) ApiName() string {
	return "applications"
}

// GetApplicationDefinitions returns a full ApplicationDefinitionDAO object with all its dependencies
func GetApplicationDefinitionById(pool *pgxpool.Pool, id uint64) (*ApplicationDefinitionDAO, error) {
	tx, err := pool.Begin(context.Background())
	if err != nil {
		return nil, err
	}
	var appDefs []ApplicationDefinitionDAO
	err = pgxscan.Select(context.Background(), tx, &appDefs, `SELECT * FROM application_definition ad WHERE ad.id = $1`, id)
	if err != nil {
		return nil, err
	}
	return &appDefs[0], nil
}

// GetApplicationDefinitions returns a full ApplicationDefinitionDAO object with all its dependencies
func GetApplicationDefinitions(pool *pgxpool.Pool) (*[]ApplicationDefinitionDAO, error) {
	tx, err := pool.Begin(context.Background())
	if err != nil {
		return nil, err
	}
	var appDefs []ApplicationDefinitionDAO
	err = pgxscan.Select(context.Background(), tx, &appDefs, `SELECT * FROM application_definition ad`)
	if err != nil {
		return nil, err
	}
	return &appDefs, nil
}

// GetApplicationDefinitions returns a full ApplicationDefinitionDAO object with all its dependencies
func GetApplicationDefinitionsFull(pool *pgxpool.Pool) (*[]ApplicationDefinition, error) {
	tx, err := pool.Begin(context.Background())
	if err != nil {
		return nil, err
	}
	var appDefs []ApplicationDefinition
	err = pgxscan.Select(context.Background(), tx, &appDefs, `
		SELECT
			ad.id AS ApplicationDefinitionID, ad.name AS ApplicationDefinitionName, ad.port AS ApplicationDefinitionPort, ad.type AS ApplicationDefinitionType, ad.healthcheck_id AS HealthcheckID,
			h.url AS HealthcheckUrl, h.timeout AS HealthcheckTimeout, h.check_interval AS HealthcheckCheckInterval, h.expected_status AS HealthcheckExpectedStatus
		FROM application_definition ad
		LEFT JOIN healthcheck h ON ad.healthcheck_id = h.id`)
	if err != nil {
		return nil, err
	}
	return &appDefs, nil
}

// Creates new ApplicationDefinition in DB
// Returns the inserted ApplicationDefinition object
func (appDef ApplicationDefinitionDAO) DbInsert(tx pgx.Tx) (*uint, error) {
	var resId uint
	var err error
	if appDef.HealthcheckId == nil {
		err = tx.QueryRow(context.Background(), "INSERT INTO application_definition (name, port, type) VALUES ($1, $2, $3) RETURNING id", appDef.Name, appDef.Port, appDef.Type).Scan(&resId)
	} else {
		err = tx.QueryRow(context.Background(), "INSERT INTO application_definition (name, port, type, healthcheck_id) VALUES ($1, $2, $3, $4) RETURNING id", appDef.Name, appDef.Port, appDef.Type, appDef.HealthcheckId).Scan(&resId)
	}
	if err != nil {
		tx.Rollback(context.Background())
		return nil, err
	}
	return &resId, nil
}

// Deletes specified ApplicationDefinition and all dependent ApplicationInstances
func (appDef ApplicationDefinitionDAO) Delete(pool *pgxpool.Pool) (*int, error) {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	var affectedRows = 0
	// Check if application definition exists
	var appDefTry ApplicationDefinitionDAO
	err = pgxscan.Get(context.Background(), tx, &appDefTry, "SELECT * FROM application_definition WHERE id = $1", appDef.ID)
	if err != nil {
		tx.Rollback(context.Background())
		return nil, err
	}
	if appDefTry.ID == 0 {
		tx.Rollback(context.Background())
		return nil, nil // The instance is technically deleted
	}
	// Check for dangling instances
	var instances []ApplicationInstanceDAO
	err = pgxscan.Select(context.Background(), tx, &instances, "DELETE FROM application_instance WHERE application_definition_id = $1", appDef.ID)
	if err != nil {
		return nil, err
	}
	if len(instances) > 0 {
		// There are dangling instances >>> delete them
		for _, instance := range instances {
			ra, err := instance.Delete(pool)
			if err != nil {
				tx.Rollback(context.Background())
				return nil, err
			}
			affectedRows += *ra
		}
	}
	// Check if there arent any dangling instances
	// Remove ApplicationDefinitionDAO
	com, err := tx.Exec(context.Background(), "DELETE FROM application_definition WHERE id = $1", appDef.ID)
	if err != nil {
		tx.Rollback(context.Background())
		return nil, err
	}
	affectedRows += int(com.RowsAffected())
	return &affectedRows, nil
}

func (ad ApplicationDefinitionDAO) GetInstances(pool *pgxpool.Pool) ([]ApplicationInstanceDAO, error) {
	var instances []ApplicationInstanceDAO
	err := pgxscan.Select(context.Background(), pool, &instances, "select * from application_instance ai where ai.application_definition_id = $1", ad.ID)
	if err != nil {
		return nil, err
	}
	return instances, nil
}
