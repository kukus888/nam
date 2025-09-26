package data

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GetApplicationDefinitionVariablesByApplicationDefinitionId retrieves all variables associated with a specific application definition
func GetApplicationDefinitionVariablesByApplicationDefinitionId(pool *pgxpool.Pool, appDefId uint64) (*[]ApplicationDefinitionVariableDAO, error) {
	tx, err := pool.Begin(context.Background())
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())
	rows, err := tx.Query(context.Background(), `SELECT * FROM application_definition_variable WHERE application_definition_id = $1`, appDefId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // No results found
	}

	vars, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[ApplicationDefinitionVariableDAO])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // No results found
	} else if err != nil {
		return nil, err
	}
	return &vars, tx.Commit(context.Background())
}

// CreateApplicationDefinitionVariable creates a new variable for an application definition
func CreateApplicationDefinitionVariable(pool *pgxpool.Pool, variable *ApplicationDefinitionVariableDAO) error {
	tx, err := pool.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), `INSERT INTO application_definition_variable (application_definition_id, name, value, description) VALUES ($1, $2, $3, $4)`, variable.ApplicationDefinitionID, variable.Name, variable.Value, variable.Description)
	if err != nil {
		return err
	}

	return tx.Commit(context.Background())
}

// DeleteApplicationDefinitionVariablesByApplicationDefinitionId deletes all variables associated with a specific application definition
func DeleteApplicationDefinitionVariablesByApplicationDefinitionId(pool *pgxpool.Pool, appDefId uint64) error {
	tx, err := pool.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), `DELETE FROM application_definition_variable WHERE application_definition_id = $1`, appDefId)
	if err != nil {
		return err
	}

	return tx.Commit(context.Background())
}

// UpdateApplicationDefinitionVariable updates an existing variable for an application definition
func UpdateApplicationDefinitionVariable(pool *pgxpool.Pool, variable *ApplicationDefinitionVariableDAO) error {
	tx, err := pool.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), `UPDATE application_definition_variable SET name = $1, value = $2, description = $3 WHERE id = $4`, variable.Name, variable.Value, variable.Description, variable.Id)
	if err != nil {
		return err
	}

	return tx.Commit(context.Background())
}

// DeleteApplicationDefinitionVariableById deletes a specific variable by its ID
func DeleteApplicationDefinitionVariableById(pool *pgxpool.Pool, varId uint64) error {
	tx, err := pool.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), `DELETE FROM application_definition_variable WHERE id = $1`, varId)
	if err != nil {
		return err
	}

	return tx.Commit(context.Background())
}
