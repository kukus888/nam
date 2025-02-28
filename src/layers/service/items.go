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
type ItemType struct {
	DisplayName  string // Name to display
	HtmxEndpoint string // which endpoint will be called at /api/htmx
}

func (is *ItemService) GetAllItemTypes() []ItemType {
	return []ItemType{
		{DisplayName: "Application Instance", HtmxEndpoint: "instances"},
		{DisplayName: "Application Definition", HtmxEndpoint: "definitions"},
		{DisplayName: "Server", HtmxEndpoint: "servers"},
		{DisplayName: "Healthcheck", HtmxEndpoint: "healthchecks"},
		{DisplayName: "Topology Nodes", HtmxEndpoint: "nodes"},
	}
}
