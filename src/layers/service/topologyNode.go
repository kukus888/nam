package services

import (
	"context"
	"errors"
	"kukus/nam/v2/layers/data"
	"strconv"

	"github.com/jackc/pgx/v5"
)

// Defines business logic regarding application structs

type TopologyNodeService struct {
	Database *data.Database
}

// Inserts new TopologyNode into database
// Returns: The new ID, error
func (tns *TopologyNodeService) CreateTopologyNode(tn data.TopologyNode) (*uint, error) {
	tx, err := tns.Database.Pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	id, err := tn.DbInsert(tx)
	if err != nil {
		tx.Rollback(context.Background())
		return nil, err
	} else {
		tx.Commit(context.Background())
		return id, nil
	}
}

// Reads All TopologyNode from database
func (tns *TopologyNodeService) GetAllTopologyNodes() (*[]data.TopologyNode, error) {
	// Insert AppDef into Db
	tx, err := tns.Database.Pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	s, err := data.DbQueryTypeAll(tx, data.TopologyNode{})
	if err != nil {
		tx.Rollback(context.Background())
		return nil, err
	} else {
		tx.Commit(context.Background())
		return &s, nil
	}
}

// Reads TopologyNode from database
func (tns *TopologyNodeService) GetTopologyNodeById(id uint) (*data.TopologyNode, error) {
	tx, err := tns.Database.Pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	nodes, err := data.DbQueryTypeWithParams(tx, data.TopologyNode{}, data.DbFilter{
		Column:   "id",
		Operator: data.DbOperatorEqual,
		Value:    strconv.Itoa(int(id)),
	})
	if err != nil {
		tx.Rollback(context.Background())
		return nil, err
	} else if len(nodes) == 0 {
		tx.Commit(context.Background())
		return nil, nil
	} else {
		tx.Commit(context.Background())
		return &nodes[0], nil
	}
}

// Removes TopologyNode from database
func (tns *TopologyNodeService) RemoveTopologyNodeById(id uint) error {
	node, err := tns.GetTopologyNodeById(id)
	if err != nil {
		return err
	} else if node == nil {
		return errors.New("topologyNode with id " + strconv.Itoa(int(id)) + " doesn't exist!")
	}
	tx, err := tns.Database.Pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return err
	}
	_, err = node.Delete(tx) // TODO: Log
	if err != nil {
		return err
	} else {
		return tx.Commit(context.Background())
	}
}
