package services

import (
	"context"
	"kukus/nam/v2/layers/data"

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
	return data.GetApplicationDefinitions(as.Database.Pool)
}

// Reads ApplicationDefinition from database
func (as *ApplicationService) GetApplicationById(id uint) (*data.ApplicationDefinitionDAO, error) {
	return data.GetApplicationDefinitionById(as.Database.Pool, uint64(id))
}

// Removes ApplicationDefinition from database
func (as *ApplicationService) RemoveApplicationById(id uint64) error {
	appDef := data.ApplicationDefinitionDAO{
		Id: uint(id),
	}
	_, err := appDef.Delete(as.Database.Pool)
	return err
}
