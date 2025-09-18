package data

import (
	"context"
	"errors"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Fetches appropriate ApplicationInstance struct from DB
func GetApplicationInstanceById(pool *pgxpool.Pool, id uint64) (*ApplicationInstance, error) {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())

	rows, err := tx.Query(context.Background(), "SELECT * FROM application_instance WHERE id = $1", id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // No results found
	} else if err != nil {
		return nil, err
	}

	res, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByNameLax[ApplicationInstance])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // No results found
	} else if err != nil {
		return nil, err
	}
	return &res, tx.Commit(context.Background())
}

// Fetches appropriate ApplicationInstanceFull struct from DB, which includes server and application definition details
// Returns the first instance found, or an error if none found
func GetApplicationInstanceFullById(pool *pgxpool.Pool, id uint64) (*ApplicationInstanceFull, error) {
	tx, err := pool.Begin(context.Background())
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())
	var inst []ApplicationInstanceFull
	err = pgxscan.Select(context.Background(), tx, &inst, `
		SELECT 
			ai.id AS application_instance_id, ai.name AS application_instance_name, ai.server_id AS server_id, ai.application_definition_id AS application_definition_id, ai.topology_node_id AS topology_node_id, ai.maintenance_mode AS maintenance_mode,
			s.alias AS server_alias, s.hostname AS server_hostname,
			ad.name AS application_definition_name, ad.port AS application_definition_port, ad.type AS application_definition_type, ad.healthcheck_id AS healthcheck_id
		FROM application_instance ai
		LEFT JOIN "server" s ON ai.server_id = s.id
		LEFT JOIN application_definition ad ON ai.application_definition_id = ad.id
		WHERE ai.id = $1`, id)
	if err != nil {
		return nil, err
	}

	return &inst[0], tx.Commit(context.Background())
}

// Returns a slice of all application instances.
func GetAllApplicationInstancesFull(pool *pgxpool.Pool) (*[]ApplicationInstanceFull, error) {
	tx, err := pool.Begin(context.Background())
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())
	var inst []ApplicationInstanceFull
	err = pgxscan.Select(context.Background(), tx, &inst, `
		SELECT 
			ai.id AS application_instance_id, ai.name AS application_instance_name, ai.server_id AS server_id, ai.application_definition_id AS application_definition_id, ai.topology_node_id AS topology_node_id, ai.maintenance_mode AS maintenance_mode,
			s.alias AS server_alias, s.hostname AS server_hostname,
			ad.name AS application_definition_name, ad.port AS application_definition_port, ad.type AS application_definition_type, ad.healthcheck_id AS healthcheck_id
		FROM application_instance ai
		LEFT JOIN "server" s ON ai.server_id = s.id
		LEFT JOIN application_definition ad ON ai.application_definition_id = ad.id`)
	if err != nil {
		return nil, err
	}

	return &inst, tx.Commit(context.Background())
}

// Returns a slice of all application instances.
func GetAllApplicationInstancesFullByHealthcheckId(pool *pgxpool.Pool, healthcheckId uint64) (*[]ApplicationInstanceFull, error) {
	tx, err := pool.Begin(context.Background())
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())
	var inst []ApplicationInstanceFull
	err = pgxscan.Select(context.Background(), tx, &inst, `
		SELECT 
			ai.id AS application_instance_id, ai.name AS application_instance_name, ai.server_id AS server_id, ai.application_definition_id AS application_definition_id, ai.topology_node_id AS topology_node_id, ai.maintenance_mode AS maintenance_mode,
			s.alias AS server_alias, s.hostname AS server_hostname,
			ad.name AS application_definition_name, ad.port AS application_definition_port, ad.type AS application_definition_type, ad.healthcheck_id AS healthcheck_id
		FROM application_instance ai
		LEFT JOIN "server" s ON ai.server_id = s.id
		LEFT JOIN application_definition ad ON ai.application_definition_id = ad.id
		WHERE ad.healthcheck_id = $1`, healthcheckId)
	if err != nil {
		return nil, err
	}

	return &inst, tx.Commit(context.Background())
}

// Returns a slice of all application instances.
func GetApplicationInstancesFullByApplicationDefinitionId(pool *pgxpool.Pool, id uint64) (*[]ApplicationInstanceFull, error) {
	tx, err := pool.Begin(context.Background())
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())
	var inst []ApplicationInstanceFull
	err = pgxscan.Select(context.Background(), tx, &inst, `
		SELECT 
			ai.id AS application_instance_id, ai.name AS application_instance_name, ai.server_id AS server_id, ai.application_definition_id AS application_definition_id, ai.topology_node_id AS topology_node_id, ai.maintenance_mode AS maintenance_mode,
			s.alias AS server_alias, s.hostname AS server_hostname,
			ad.name AS application_definition_name, ad.port AS application_definition_port, ad.type AS application_definition_type, ad.healthcheck_id AS healthcheck_id
		FROM application_instance ai
		LEFT JOIN "server" s ON ai.server_id = s.id
		LEFT JOIN application_definition ad ON ai.application_definition_id = ad.id
		WHERE ai.application_definition_id = $1`, id)
	if err != nil {
		return nil, err
	}

	return &inst, tx.Commit(context.Background())
}

// Creates new ApplicationInstance in DB, with underlying TopologyNode struct.
// Returns the inserted ApplicationInstance ID
func CreateApplicationInstance(pool *pgxpool.Pool, instance ApplicationInstance) (*uint, error) {
	tx, err := pool.Begin(context.Background())
	if err != nil {
		return nil, err
	}
	// Validate data
	if instance.Name == "" {
		return nil, errors.New("application instance must have set name")
	}
	if instance.ServerID == 0 {
		return nil, errors.New("application instance must have set server_id")
	}
	if instance.ApplicationDefinitionID == 0 {
		return nil, errors.New("application instance must have set application_definition_id")
	}

	// Create underlying topologyNode
	var tnid uint
	err = tx.QueryRow(context.Background(), "INSERT INTO topology_node (type) VALUES ($1) RETURNING id", "application_instance").Scan(&tnid)
	if err != nil {
		tx.Rollback(context.Background())
		return nil, err
	}
	// Insert instance into DB
	instance.Id = tnid
	var resId uint
	err = tx.QueryRow(context.Background(), "INSERT INTO application_instance (id, name, server_id, application_definition_id) VALUES ($1, $2, $3, $4) RETURNING id", instance.Id, instance.Name, instance.ServerID, instance.ApplicationDefinitionID).Scan(&resId)
	if err != nil {
		tx.Rollback(context.Background())
		return nil, err
	}
	return &resId, tx.Commit(context.Background())
}

// Deletes specified ApplicationInstance, with their corresponding TopologyNodes
func DeleteApplicationInstanceById(pool *pgxpool.Pool, id uint64) error {
	tx, err := pool.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())
	// Get TopologyNode ID
	var topologyNodeId uint
	err = tx.QueryRow(context.Background(), `
		select tn.id from application_instance ai 
		full join topology_node tn on ai.topology_node_id = tn.id 
		where ai.id = $1`, id).Scan(&topologyNodeId)
	if err != nil && id != 0 {
		return err
	}
	// Delete everything
	_, err = tx.Exec(context.Background(), `
		delete from healthcheck_results hr 
		where hr.application_instance_id = $1`, id)
	if err != nil {
		return err
	}
	_, err = tx.Exec(context.Background(), "DELETE FROM application_instance WHERE id = $1", id)
	if err != nil {
		return err
	}
	_, err = tx.Exec(context.Background(), "DELETE FROM topology_node WHERE id = $1", topologyNodeId)
	if err != nil {
		return err
	}
	return tx.Commit(context.Background())
}

// Toggles maintenance mode for the specified application instance
func ToggleApplicationInstanceMaintenance(pool *pgxpool.Pool, id uint64, maintenanceMode bool) error {
	tx, err := pool.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), "UPDATE application_instance SET maintenance_mode = $1 WHERE id = $2", maintenanceMode, id)
	if err != nil {
		return err
	}
	return tx.Commit(context.Background())
}
