package services

import "kukus/nam/v2/layers/data"

type ItemService struct {
	Database *data.Database
}

func NewItemService(database *data.Database) ItemService {
	return ItemService{
		Database: database,
	}
}

type ItemBO struct {
	Name string
}

// Represents a type of an item, such as TopologyNode, ApplicationDefinition, ...
type ItemTypeBO struct {
}

func (is *ItemService) GetAllItemTypes() []string {
	return []string{
		"TopologyNode",
		"ApplicationDefinition",
		"Server",
		"ApplicationInstance",
		"Healthcheck",
	}
}
