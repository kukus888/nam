package services

import (
	"context"
	"errors"
	"kukus/nam/v2/layers/data"
	"strconv"

	"github.com/jackc/pgx/v5"
)

// Defines business logic regarding application structs

type ApplicationService struct {
	Database *data.Database
}

// Inserts new ApplicationDefinition into database
// Returns: The new ID, error
func (as *ApplicationService) CreateApplication(appDef data.ApplicationDefinitionDAO) (*uint, error) {
	// Insert AppDef into Db
	tx, err := as.Database.Pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	id, err := appDef.DbInsert(tx)
	if err != nil {
		tx.Rollback(context.Background())
		return nil, err
	} else {
		tx.Commit(context.Background())
		return id, nil
	}
}

// Reads All ApplicationDefinition from database
func (as *ApplicationService) GetAllApplications() (*[]data.ApplicationDefinitionDAO, error) {
	// Insert AppDef into Db
	tx, err := as.Database.Pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	daos, err := data.DbQueryTypeAll(tx, data.ApplicationDefinitionDAO{})
	if err != nil {
		tx.Rollback(context.Background())
		return nil, err
	} else {
		tx.Commit(context.Background())
		return &daos, nil
	}
}

// Reads ApplicationDefinition from database
func (as *ApplicationService) GetApplicationById(id uint) (*data.ApplicationDefinitionDAO, error) {
	// Insert AppDef into Db
	tx, err := as.Database.Pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	daos, err := data.DbQueryTypeWithParams(tx, data.ApplicationDefinitionDAO{}, data.DbFilter{
		Column:   "id",
		Operator: data.DbOperatorEqual,
		Value:    strconv.Itoa(int(id)),
	})
	if err != nil {
		tx.Rollback(context.Background())
		return nil, err
	} else if len(daos) == 0 {
		tx.Commit(context.Background())
		return nil, nil
	} else {
		tx.Commit(context.Background())
		return &daos[0], nil
	}
}

// Removes ApplicationDefinition from database
func (as *ApplicationService) RemoveApplicationById(id uint) error {
	tx, err := as.Database.Pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return err
	}
	// Check for dangling instances!
	instances, err := data.DbQueryTypeWithParams(tx, data.ApplicationInstanceDAO{}, data.DbFilter{
		Column:   "application_instance_id",
		Operator: data.DbOperatorEqual,
		Value:    strconv.Itoa(int(id)),
	})
	if err != nil {
		tx.Rollback(context.Background())
		return err
	}
	if len(instances) > 0 {
		for _, instance := range instances {
			_, err = instance.Delete(tx) // TODO: Log
			if err != nil {
				tx.Rollback(context.Background())
				return err
			}
		}
	}
	app, err := as.GetApplicationById(id)
	if err != nil {
		return err
	} else if app == nil {
		return errors.New("application with id " + strconv.Itoa(int(id)) + " doesn't exist!")
	}
	_, err = app.Delete(tx) // TODO: Log
	if err != nil {
		return err
	} else {
		return tx.Commit(context.Background())
	}
}
