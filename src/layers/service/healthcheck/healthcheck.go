package healthcheck

import (
	data "kukus/nam/v2/layers/data"
)

// Implements business logic behind healthcheck objects
type HealthcheckService struct {
	db *data.Database
}

// Inserts the Healthcheck object into the database
func DbInsert(hc *data.Healthcheck) (*data.Healthcheck, error) {
	main.Database.Pool
	return nil, nil
}
