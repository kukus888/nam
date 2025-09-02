package handlers

import (
	"kukus/nam/v2/layers/data"
)

type PageHandler struct {
	Database *data.Database
}

func NewPageHandler(database *data.Database) PageHandler {
	return PageHandler{
		Database: database,
	}
}
