package data

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

// ApplicationInstanceDAO represents an instance of an application
type ApplicationInstanceDAO struct {
	ID           uint   `json:"id"`
	ServerId     uint   `db:"server_id" json:"server_id"`
	DefinitionId uint   `db:"application_definition_id" json:"application_definition_id"`
	Name         string `json:"name"`
}

func (s ApplicationInstanceDAO) TableName() string {
	return "application_instance"
}

func (s ApplicationInstanceDAO) ApiName() string {
	return "instances"
}

// Creates new ApplicationInstance in DB, with underlying TopologyNode struct.
// Returns the inserted ApplicationInstance object
func (dto ApplicationInstanceDAO) DbInsert(tx pgx.Tx) (*uint, error) {
	if dto.Name == "" {
		return nil, errors.New("application instance must have set name")
	}
	// Create underlying topologyNode
	tn := TopologyNode{Type: "application_instance"}
	tnId, err := tn.DbInsert(tx)
	if err != nil {
		tx.Rollback(context.Background())
		return nil, err
	}
	// Insert instance into DB
	dto.ID = *tnId
	var resId uint
	err = tx.QueryRow(context.Background(), "INSERT INTO application_instance (id, name, server_id, application_definition_id) VALUES ($1, $2, $3, $4) RETURNING id", dto.ID, dto.Name, dto.ServerId, dto.DefinitionId).Scan(&resId)
	if err != nil {
		tx.Rollback(context.Background())
		return nil, err
	}
	return &resId, nil
}

// Deletes specified ApplicationInstanceDAO, with their corresponding TopologyNodes
func (dao ApplicationInstanceDAO) Delete(tx pgx.Tx) (*int, error) {
	var ra = 0
	// Remove ApplicationInstanceDAO
	com, err := tx.Exec(context.Background(), "DELETE FROM application_instance WHERE id = $1", dao.ID)
	if err != nil {
		tx.Rollback(context.Background())
		return nil, err
	}
	ra += int(com.RowsAffected())
	return &ra, nil
}
