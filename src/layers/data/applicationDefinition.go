package data

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// ApplicationDefinitionDAO represents the definition of an application and its general properties
type ApplicationDefinitionDAO struct {
	ID            uint   `json:"id"`
	Name          string `json:"name"`
	Port          int    `json:"port"`
	Type          string `json:"type"`
	HealthcheckId *uint  `db:"healthcheck_id" json:"healthcheck_id"`
}

func (s ApplicationDefinitionDAO) TableName() string {
	return "application_definition"
}

func (s ApplicationDefinitionDAO) ApiName() string {
	return "applications"
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

// Deletes specified ApplicationDefinitionDAO and all dependent ApplicationInstances
func (appDef ApplicationDefinitionDAO) Delete(tx pgx.Tx) (*int, error) {
	var ra = 0
	// Check if there arent any dangling instances
	// Remove ApplicationDefinitionDAO
	com, err := tx.Exec(context.Background(), "DELETE FROM application_definition WHERE id = $1", appDef.ID)
	if err != nil {
		return nil, err
	}
	ra += int(com.RowsAffected())
	return &ra, nil
}
