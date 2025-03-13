package data

import (
	"net/http"
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
	Description    string        `json:"description" db:"description"`
	Url            string        `json:"url" db:"url"`
	Method         string        `json:"method" db:"method"`   // GET, POST, etc.
	Headers        []http.Header `json:"headers" db:"headers"` // Custom headers
	Body           string        `json:"body" db:"body"`       // Request body for POST/PUT
	Timeout        time.Duration `json:"timeout" db:"timeout"`
	CheckInterval  time.Duration `json:"check_interval" db:"check_interval"`
	RetryCount     int           `json:"retry_count" db:"retry_count"`       // Number of retries before marking as unhealthy
	RetryInterval  time.Duration `json:"retry_interval" db:"retry_interval"` // Time between retries
	ExpectedStatus int           `json:"expected_status" db:"expected_status"`

	// Response validation
	ExpectedResponseBody string `json:"expected_response_body" db:"expected_response_body"` // Expected response content
	ResponseValidation   string `json:"response_validation" db:"response_validation"`       // contains, exact, regex

	// SSL/TLS
	VerifySSL      bool `json:"verify_ssl" db:"verify_ssl"`
	SSLExpiryAlert bool `json:"ssl_expiry_alert" db:"ssl_expiry_alert"`

	// Authentication
	AuthType        string `json:"auth_type" db:"auth_type"`               // none, basic, bearer, custom
	AuthCredentials string `json:"auth_credentials" db:"auth_credentials"` // stored securely
}

type HealthcheckRecord struct {
	ID               uint64     `json:"id" db:"id"`
	HealthcheckID    uint       `json:"healthcheck_id" db:"healthcheck_id"`
	Timestamp        time.Time  `json:"timestamp" db:"timestamp"`
	Status           string     `json:"status" db:"status"`
	HttpResponseCode int        `json:"http_response_code" db:"http_response_code"`
	HttpResponseBody string     `json:"http_response_body" db:"http_response_body"`
	ResponseTime     int64      `json:"response_time" db:"response_time"` // in milliseconds
	ErrorMessage     string     `json:"error_message" db:"error_message"`
	SSLValid         *bool      `json:"ssl_valid" db:"ssl_valid"`
	SSLExpiry        *time.Time `json:"ssl_expiry" db:"ssl_expiry"`
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
