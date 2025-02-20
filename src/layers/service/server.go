package services

import (
	"context"
	"errors"
	"kukus/nam/v2/layers/data"
	"strconv"

	"github.com/jackc/pgx/v5"
)

// Defines business logic regarding server structs

type ServerService struct {
	Database *data.Database
}

// Inserts new Server into database
// Returns: The new ID, error
func (sc *ServerService) CreateServer(server data.ServerDAO) (*uint, error) {
	tx, err := sc.Database.Pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	id, err := server.DbInsert(tx)
	if err != nil {
		tx.Rollback(context.Background())
		return nil, err
	} else {
		tx.Commit(context.Background())
		return id, nil
	}
}

// Reads All Servers from database
func (sc *ServerService) GetAllServers() (*[]data.ServerDAO, error) {
	// Insert AppDef into Db
	tx, err := sc.Database.Pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	daos, err := data.DbQueryTypeAll(tx, data.ServerDAO{})
	if err != nil {
		tx.Rollback(context.Background())
		return nil, err
	} else {
		tx.Commit(context.Background())
		return &daos, nil
	}
}

// Reads Servers from database
func (sc *ServerService) GetServerById(id uint) (*data.ServerDAO, error) {
	tx, err := sc.Database.Pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	daos, err := data.DbQueryTypeWithParams(tx, data.ServerDAO{}, data.DbFilter{
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

// Removes Server from database
func (sc *ServerService) RemoveApplicationById(id uint) error {
	tx, err := sc.Database.Pool.BeginTx(context.Background(), pgx.TxOptions{})
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
	server, err := sc.GetServerById(id)
	if err != nil {
		return err
	} else if server == nil {
		return errors.New("server with id " + strconv.Itoa(int(id)) + " doesn't exist!")
	}
	_, err = server.Delete(tx) // TODO: Log
	if err != nil {
		return err
	} else {
		return tx.Commit(context.Background())
	}
}
