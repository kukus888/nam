package data

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// Topology represents a Topology element in the topology yaml file
type TopologyNode struct {
	ID   uint   `json:"id"`
	Type string `json:"type"` // Type of Topology element, e.g. loadbalancer, firewall, etc. \n The Type corresponds to the table name in the database
}

func (tn TopologyNode) TableName() string {
	return "topology_node"
}

func (tn TopologyNode) ApiName() string {
	return "nodes"
}

// Inserts TopologyNode into Database. New ID is stored in the referenced TopologyNode
// Does not roll back transaction, this is merely a facade for an insert statement
func (tn TopologyNode) DbInsert(tx pgx.Tx) (*uint, error) {
	var id uint
	err := tx.QueryRow(context.Background(), "INSERT INTO topology_node (type) VALUES ($1) RETURNING id", tn.Type).Scan(&id)
	return &id, err
}

// Deletes specified TopologyNode, with their corresponding TopologyNodes
func (tn TopologyNode) Delete(tx pgx.Tx) (*int, error) {
	var ra = 0
	// Remove TopologyNode
	com, err := tx.Exec(context.Background(), "DELETE FROM topology_node WHERE id = $1", tn.ID)
	if err != nil {
		tx.Rollback(context.Background())
		return nil, err
	}
	ra += int(com.RowsAffected())
	return &ra, nil
}
