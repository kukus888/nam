package data

import (
	"time"
)

type Server struct {
	ID       uint   `json:"server_id" db:"serverid"`
	Alias    string `json:"alias" db:"serveralias"`
	Hostname string `json:"hostname" db:"serverhostname"`
}

type Healthcheck struct {
	ID             uint          `json:"id" db:"id"`
	Name           string        `json:"name" db:"name"`
	Url            string        `json:"url" db:"url"`
	Timeout        time.Duration `json:"timeout" db:"timeout"`
	CheckInterval  time.Duration `json:"check_interval" db:"check_interval"`
	ExpectedStatus int           `json:"expected_status" db:"expected_status"`
}

type HealthcheckRecord struct {
	ID               uint64      `json:"id"`
	Healthcheck      Healthcheck `json:"healthcheck"`
	Timestamp        time.Time   `json:"timestamp"`
	HttpResponseCode uint        `json:"http_response_code"`
	HttpResponseBody string      `json:"http_response_body"`
	Healthy          bool        `json:"healthy"`
}

// ApplicationDefinition represents the definition of an application and its general properties
type ApplicationDefinition struct {
	ID          uint         `json:"id" db:"id"`
	Name        string       `json:"name" db:"name"`
	Port        int          `json:"port" db:"port"`
	Type        string       `json:"type" db:"type"`
	Healthcheck *Healthcheck `json:"healthcheck"`
}

// ApplicationInstance represents an instance of an application
type ApplicationInstance struct {
	ID                      uint   `json:"id" db:"applicationinstanceid"`
	Name                    string `json:"name" db:"applicationinstancename"`
	TopologyNodeID          uint   `json:"topology_node_id" db:"topologynodeid"`
	ApplicationDefinitionID uint   `json:"application_definition_id"`
	Server                  Server `json:"server"`
}
