package data

import (
	"context"
	"errors"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ApplicationInstanceDAO represents an instance of an application
type ApplicationInstanceDAO struct {
	Id                      uint   `json:"id" db:"id"`
	ServerId                uint   `db:"server_id" json:"server_id"`
	ApplicationDefinitionId uint   `db:"application_definition_id" json:"application_definition_id"`
	TopologyNodeId          uint   `db:"topology_node_id" json:"topology_node_id"`
	Name                    string `json:"name" db:"name"`
}

func GetApplicationInstanceFullById(pool *pgxpool.Pool, id uint64) (*ApplicationInstance, error) {
	tx, err := pool.Begin(context.Background())
	if err != nil {
		return nil, err
	}
	var inst []ApplicationInstance
	err = pgxscan.Select(context.Background(), tx, &inst, `
		SELECT 
			ai.id AS ApplicationInstanceID, ai.name AS ApplicationInstanceName, ai.server_id AS ServerId, ai.application_definition_id AS ApplicationDefinitionID, ai.topology_node_id AS TopologyNodeID,
			s.alias AS ServerAlias, s.hostname AS ServerHostname,
			ad.name AS ApplicationDefinitionName, ad.port AS ApplicationDefinitionPort, ad.port AS ApplicationDefinitionType, ad.healthcheck_id AS HealthcheckID,
			h.url AS HealthcheckUrl, h.timeout AS HealthcheckTimeout, h.check_interval AS HealthcheckCheckInterval, h.expected_status AS HealthcheckExpectedStatus
		FROM application_instance ai
		LEFT JOIN "server" s ON ai.server_id = s.id
		LEFT JOIN application_definition ad ON ai.application_definition_id = ad.id
		LEFT JOIN healthcheck h ON ad.healthcheck_id = h.id
		WHERE ai.id = $1`, id)
	if err != nil {
		return nil, err
	}

	return &inst[0], nil
}

// Returns a slice of all application instances.
func GetAllApplicationInstancesFull(pool *pgxpool.Pool) (*[]ApplicationInstance, error) {
	tx, err := pool.Begin(context.Background())
	if err != nil {
		return nil, err
	}
	var inst []ApplicationInstance
	err = pgxscan.Select(context.Background(), tx, &inst, `
		SELECT 
			ai.id AS ApplicationInstanceID, ai.name AS ApplicationInstanceName, ai.server_id AS ServerId, ai.application_definition_id AS ApplicationDefinitionID, ai.topology_node_id AS TopologyNodeID,
			s.alias AS ServerAlias, s.hostname AS ServerHostname,
			ad.name AS ApplicationDefinitionName, ad.port AS ApplicationDefinitionPort, ad.port AS ApplicationDefinitionType, ad.healthcheck_id AS HealthcheckID,
			h.url AS HealthcheckUrl, h.timeout AS HealthcheckTimeout, h.check_interval AS HealthcheckCheckInterval, h.expected_status AS HealthcheckExpectedStatus
		FROM application_instance ai
		LEFT JOIN "server" s ON ai.server_id = s.id
		LEFT JOIN application_definition ad ON ai.application_definition_id = ad.id
		LEFT JOIN healthcheck h ON ad.healthcheck_id = h.id`)
	if err != nil {
		return nil, err
	}

	return &inst, nil
}

// Creates new ApplicationInstance in DB, with underlying TopologyNode struct.
// Returns the inserted ApplicationInstance ID
func CreateApplicationInstance(pool *pgxpool.Pool, instance ApplicationInstanceDAO) (*uint, error) {
	tx, err := pool.Begin(context.Background())
	if err != nil {
		return nil, err
	}
	// Validate data
	if instance.Name == "" {
		return nil, errors.New("application instance must have set name")
	}
	if instance.ServerId == 0 {
		return nil, errors.New("application instance must have set server_id")
	}
	if instance.ApplicationDefinitionId == 0 {
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
	err = tx.QueryRow(context.Background(), "INSERT INTO application_instance (id, name, server_id, application_definition_id) VALUES ($1, $2, $3, $4) RETURNING id", instance.Id, instance.Name, instance.ServerId, instance.ApplicationDefinitionId).Scan(&resId)
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

// Deletes specified ApplicationInstance, with their corresponding TopologyNodes
func (instance ApplicationInstanceDAO) Delete(pool *pgxpool.Pool) (*int, error) {
	return nil, DeleteApplicationInstanceById(pool, uint64(instance.Id)) // Delete all healthcheck results for this instance
}
