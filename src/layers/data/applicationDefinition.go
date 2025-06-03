package data

import (
	"context"
	"errors"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func (s ApplicationDefinitionDAO) TableName() string {
	return "application_definition"
}

func (s ApplicationDefinitionDAO) ApiName() string {
	return "applications"
}

// GetApplicationDefinitions returns a full ApplicationDefinitionDAO object with all its dependencies
func GetApplicationDefinitionById(pool *pgxpool.Pool, id uint64) (*ApplicationDefinitionDAO, error) {
	if id == 0 {
		return nil, errors.New("ID is required") // No ID provided, return nil
	}
	tx, err := pool.Begin(context.Background())
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())
	var appDefs []ApplicationDefinitionDAO
	err = pgxscan.Select(context.Background(), tx, &appDefs, `SELECT * FROM application_definition ad WHERE ad.id = $1`, id)
	if err != nil {
		return nil, err
	}
	return &appDefs[0], tx.Commit(context.Background())
}

// GetApplicationDefinitionsAll returns all ApplicationDefinitionDAO objects from the database
// Returns a slice of ApplicationDefinitionDAO objects
func GetApplicationDefinitionsAll(pool *pgxpool.Pool) (*[]ApplicationDefinitionDAO, error) {
	tx, err := pool.Begin(context.Background())
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())
	var appDefs []ApplicationDefinitionDAO
	err = pgxscan.Select(context.Background(), tx, &appDefs, `SELECT * FROM application_definition ad`)
	if err != nil {
		return nil, err
	}
	return &appDefs, tx.Commit(context.Background())
}

// GetApplicationDefinitions returns a full ApplicationDefinitionDAO object with all its dependencies
func GetApplicationDefinitionsFull(pool *pgxpool.Pool) (*[]ApplicationDefinition, error) {
	tx, err := pool.Begin(context.Background())
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())
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
	return &appDefs, tx.Commit(context.Background())
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

func DeleteApplicationDefinitionById(pool *pgxpool.Pool, id uint64) error {
	appDef := ApplicationDefinitionDAO{Id: uint(id)}
	_, err := appDef.Delete(pool)
	return err
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
	err = pgxscan.Get(context.Background(), tx, &appDefTry, "SELECT * FROM application_definition WHERE id = $1", appDef.Id)
	if err != nil {
		tx.Rollback(context.Background())
		return nil, err
	}
	if appDefTry.Id == 0 {
		tx.Rollback(context.Background())
		return nil, nil // The instance is technically deleted
	}
	// Check for dangling instances
	var instances []ApplicationInstance
	err = pgxscan.Select(context.Background(), tx, &instances, "DELETE FROM application_instance WHERE application_definition_id = $1", appDef.Id)
	if err != nil {
		return nil, err
	}
	if len(instances) > 0 {
		// There are dangling instances >>> delete them
		for _, instance := range instances {
			err := DeleteApplicationInstanceById(pool, uint64(instance.Id))
			if err != nil {
				tx.Rollback(context.Background())
				return nil, err
			}
		}
	}
	// Check if there arent any dangling instances
	// Remove ApplicationDefinitionDAO
	com, err := tx.Exec(context.Background(), "DELETE FROM application_definition WHERE id = $1", appDef.Id)
	if err != nil {
		tx.Rollback(context.Background())
		return nil, err
	}
	affectedRows += int(com.RowsAffected())
	return &affectedRows, tx.Commit(context.Background())
}

// GetApplicationInstancesByApplicationDefinitionId returns all ApplicationInstance objects for a given ApplicationDefinition ID
func GetApplicationInstancesByApplicationDefinitionId(pool *pgxpool.Pool, id uint64) (*[]ApplicationInstance, error) {
	tx, err := pool.Begin(context.Background())
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())
	var instances []ApplicationInstance
	err = pgxscan.Select(context.Background(), tx, &instances, `SELECT * FROM application_instance ai WHERE ai.application_definition_id = $1`, id)
	if err != nil {
		return nil, err
	}
	return &instances, tx.Commit(context.Background())
}
