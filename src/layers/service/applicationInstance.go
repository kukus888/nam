package services

import (
	"context"
	"errors"
	"kukus/nam/v2/layers/data"
	"strconv"

	"github.com/jackc/pgx/v5"
)

// Defines business logic regarding application structs

type ApplicationInstanceService struct {
	Database *data.Database
}

// Inserts new ApplicationInstanceDAO into database
// Returns: The new ID, error
func (as *ApplicationInstanceService) CreateApplicationInstance(appInst data.ApplicationInstanceDAO) (*uint, error) {
	tx, err := as.Database.Pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	id, err := appInst.DbInsert(tx)
	if err != nil {
		tx.Rollback(context.Background())
		return nil, err
	} else {
		tx.Commit(context.Background())
		return id, nil
	}
}

// Reads All ApplicationInstanceDAO from database
func (as *ApplicationInstanceService) GetAllApplicationInstances(applicationInstanceId int) (*[]data.ApplicationInstanceDAO, error) {
	// Insert AppDef into Db
	tx, err := as.Database.Pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	daos, err := data.DbQueryTypeWithParams(tx, data.ApplicationInstanceDAO{}, data.DbFilter{
		Column:   "application_definition_id",
		Operator: data.DbOperatorEqual,
		Value:    strconv.Itoa(applicationInstanceId),
	})
	if err != nil {
		tx.Rollback(context.Background())
		return nil, err
	} else {
		tx.Commit(context.Background())
		return &daos, nil
	}
}

// Reads ApplicationInstanceDAO from database
func (as *ApplicationInstanceService) GetApplicationInstanceById(id uint64) (*data.ApplicationInstanceDAO, error) {
	data.GetApplicationInstanceFull(as.Database.Pool, id)
	panic("not implemented")
}

// Removes ApplicationInstanceDAO from database
func (as *ApplicationInstanceService) RemoveApplicationInstanceById(id uint64) error {
	instance, err := as.GetApplicationInstanceById(id)
	if err != nil {
		return err
	} else if instance == nil {
		return errors.New("application with id " + strconv.Itoa(int(id)) + " doesn't exist!")
	}
	tx, err := as.Database.Pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return err
	}
	// Delete TopologyNode first
	_, err = instance.Delete(tx)
	if err != nil {
		tx.Rollback(context.Background())
		return err
	}
	// Delete instance
	_, err = instance.Delete(tx) // TODO: Log
	if err != nil {
		tx.Rollback(context.Background())
		return err
	} else {
		return tx.Commit(context.Background())
	}
}
