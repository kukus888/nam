package data

import "time"

type Server struct {
	ID       uint   `json:"server_id" db:"serverid"`
	Alias    string `json:"alias" db:"serveralias"`
	Hostname string `json:"hostname" db:"serverhostname"`
}

type Healthcheck struct {
	ID             uint          `json:"id" db:"healthcheckid"`
	Url            string        `json:"url" db:"healthcheckurl"`
	Timeout        time.Duration `json:"timeout" db:"healthchecktimeout"`
	CheckInterval  time.Duration `json:"check_interval" db:"healthcheckcheckinterval"`
	ExpectedStatus int           `json:"expected_status" db:"healthcheckexpectedstatus"`
}

// ApplicationDefinition represents the definition of an application and its general properties
type ApplicationDefinition struct {
	ID          uint   `json:"id" db:"applicationdefinitionid"`
	Name        string `json:"name" db:"applicationdefinitionname"`
	Port        int    `json:"port" db:"applicationdefinitionport"`
	Type        string `json:"type" db:"applicationdefinitiontype"`
	Healthcheck `json:"healthcheck_id"`
}

// ApplicationInstance represents an instance of an application
type ApplicationInstance struct {
	ID                    uint   `json:"id" db:"applicationinstanceid"`
	Name                  string `json:"name" db:"applicationinstancename"`
	Server                `json:"server_id"`
	ApplicationDefinition `json:"application_definition_id"`
}
