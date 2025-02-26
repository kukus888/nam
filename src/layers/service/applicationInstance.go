package services

import (
	"errors"
	"kukus/nam/v2/layers/data"
	"strconv"
)

// Defines business logic regarding application structs

type ApplicationInstanceService struct {
	Database *data.Database
}

// Inserts new ApplicationInstance into database
// Returns: The new ID, error
func (as *ApplicationInstanceService) CreateApplicationInstance(appInst data.ApplicationInstanceDAO) (*uint, error) {
	return appInst.Create(as.Database.Pool)
}

// Reads All ApplicationInstance from database, with Server, ApplicationDefinition and Healthcheck
func (as *ApplicationInstanceService) GetApplicationInstancesFull() (*[]data.ApplicationInstance, error) {
	return data.GetAllApplicationInstancesFull(as.Database.Pool)
}

// Reads ApplicationInstanceDAO from database
func (as *ApplicationInstanceService) GetApplicationInstanceById(id uint64) (*data.ApplicationInstance, error) {
	return data.GetApplicationInstanceFull(as.Database.Pool, id)
}

// Removes ApplicationInstanceDAO from database
func (as *ApplicationInstanceService) RemoveApplicationInstanceById(id uint64) error {
	instance, err := as.GetApplicationInstanceById(id)
	if err != nil {
		return err
	} else if instance == nil {
		return errors.New("application with id " + strconv.Itoa(int(id)) + " doesn't exist!")
	}

	_, err = instance.Delete(as.Database.Pool) // TODO: Log
	return err
}
