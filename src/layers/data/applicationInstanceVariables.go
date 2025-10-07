package data

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GetApplicationInstanceVariablesByApplicationInstanceId retrieves all variables associated with a specific application instance
// It also includes variables inherited from the application definition
func GetApplicationInstanceVariablesByApplicationInstanceId(pool *pgxpool.Pool, appInstanceId uint64) (*[]ApplicationInstanceVariableDAO, error) {
	tx, err := pool.Begin(context.Background())
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())
	rows, err := tx.Query(context.Background(), `SELECT * FROM application_instance_variable WHERE application_instance_id = $1`, appInstanceId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // No results found
	}

	vars, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[ApplicationInstanceVariableDAO])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // No results found
	} else if err != nil {
		return nil, err
	}

	// Collect inherited variables from application definition
	appInstance, err := GetApplicationInstanceById(pool, appInstanceId)
	if err != nil {
		return nil, err
	}
	appDefVars, err := GetApplicationDefinitionVariablesByApplicationDefinitionId(pool, uint64(appInstance.ApplicationDefinitionID))
	if err != nil {
		return nil, err
	}

	// Combine instance variables with inherited variables
	for _, v := range *appDefVars {
		variable := &ApplicationInstanceVariableDAO{
			Name:                  v.Name,
			Value:                 v.Value,
			Description:           v.Description,
			ApplicationInstanceID: uint(appInstanceId),
			IsInherited:           true,
		}
		vars = append(vars, *variable)
	}
	return &vars, tx.Commit(context.Background())
}

// CreateApplicationInstanceVariable creates a new variable for an application instance
func CreateApplicationInstanceVariable(pool *pgxpool.Pool, variable *ApplicationInstanceVariableDAO) error {
	tx, err := pool.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), `INSERT INTO application_instance_variable (application_instance_id, name, value, description) VALUES ($1, $2, $3, $4)`, variable.ApplicationInstanceID, variable.Name, variable.Value, variable.Description)
	if err != nil {
		return err
	}

	return tx.Commit(context.Background())
}

// DeleteApplicationInstanceVariablesByApplicationInstanceId deletes all variables associated with a specific application instance
func DeleteApplicationInstanceVariablesByApplicationInstanceId(pool *pgxpool.Pool, appInstanceId uint64) error {
	tx, err := pool.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), `DELETE FROM application_instance_variable WHERE application_instance_id = $1`, appInstanceId)
	if err != nil {
		return err
	}

	return tx.Commit(context.Background())
}

// UpdateApplicationInstanceVariable updates an existing variable for an application instance
func UpdateApplicationInstanceVariable(pool *pgxpool.Pool, variable *ApplicationInstanceVariableDAO) error {
	tx, err := pool.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), `UPDATE application_instance_variable SET name = $1, value = $2, description = $3 WHERE id = $4`, variable.Name, variable.Value, variable.Description, variable.Id)
	if err != nil {
		return err
	}

	return tx.Commit(context.Background())
}

// DeleteApplicationInstanceVariableById deletes a specific variable by its ID
func DeleteApplicationInstanceVariableById(pool *pgxpool.Pool, varId uint64) error {
	tx, err := pool.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), `DELETE FROM application_instance_variable WHERE id = $1`, varId)
	if err != nil {
		return err
	}

	return tx.Commit(context.Background())
}
