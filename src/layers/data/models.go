package data

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	ID       uint   `json:"server_id" db:"serverid"`
	Alias    string `json:"alias" db:"serveralias"`
	Hostname string `json:"hostname" db:"serverhostname"`
}

type Healthcheck struct {
	ID             *uint         `json:"id" db:"id"`
	Name           string        `json:"name" db:"name"`
	Description    string        `json:"description" db:"description"`
	Url            string        `json:"url" db:"url"`
	ReqMethod      string        `json:"method" db:"method"`   // GET, POST, etc.
	ReqHttpHeader  http.Header   `json:"headers" db:"headers"` // Custom headers
	ReqBody        string        `json:"body" db:"body"`       // Request body for POST/PUT
	ReqTimeout     time.Duration `json:"timeout" db:"timeout"`
	CheckInterval  time.Duration `json:"check_interval" db:"check_interval"`
	RetryCount     int           `json:"retry_count" db:"retry_count"`       // Number of retries before marking as unhealthy
	RetryInterval  time.Duration `json:"retry_interval" db:"retry_interval"` // Time between retries
	ExpectedStatus int           `json:"expected_status" db:"expected_status"`

	// Response validation
	ExpectedResponseBody string `json:"expected_response_body" db:"expected_response_body"` // Expected response content
	ResponseValidation   string `json:"response_validation" db:"response_validation"`       // contains, exact, regex

	// SSL/TLS
	VerifySSL bool `json:"verify_ssl" db:"verify_ssl"`

	// Authentication
	AuthType        string `json:"auth_type" db:"auth_type"`               // none, basic, bearer, custom
	AuthCredentials string `json:"auth_credentials" db:"auth_credentials"` // stored securely
}

type HealthcheckDTO struct {
	ID            *uint      `json:"id"`
	Name          string     `json:"name" binding:"required"`
	Description   string     `json:"description"`
	ReqUrl        string     `json:"url" binding:"required"`
	ReqMethod     string     `json:"method" binding:"required"` // GET, POST, etc.
	ReqHeader     [][]string `json:"headers"`                   // Custom headers
	ReqBody       string     `json:"body"`                      // Request body for POST/PUT
	ReqTimeout    int        `json:"timeout" binding:"required"`
	CheckInterval int        `json:"check_interval" binding:"required"`
	RetryCount    int        `json:"retry_count" binding:"required"`    // Number of retries before marking as unhealthy
	RetryInterval int        `json:"retry_interval" binding:"required"` // Time between retries

	// Response validation
	ExpectedStatus       int     `json:"expected_status" binding:"required"`
	ExpectedResponseBody *string `json:"expected_response_body"` // Expected response content
	ResponseValidation   *string `json:"response_validation"`    // contains, exact, regex

	// SSL/TLS
	VerifySSL string `json:"verify_ssl"`

	// Authentication
	AuthType        string `json:"auth_type"`        // none, basic, bearer, custom
	AuthCredentials string `json:"auth_credentials"` // stored securely
}

func (dto HealthcheckDTO) ToHealthcheck() Healthcheck {
	httpHeader := http.Header{}
	for i := range dto.ReqHeader {
		httpHeader[dto.ReqHeader[i][0]] = strings.Split(dto.ReqHeader[i][1], ",")
	}
	reqTimeout, _ := time.ParseDuration(strconv.Itoa(dto.ReqTimeout) + "s")
	reqInterval, _ := time.ParseDuration(strconv.Itoa(dto.CheckInterval) + "s")
	hc := Healthcheck{
		ID:                   dto.ID,
		Name:                 dto.Name,
		Description:          dto.Description,
		Url:                  dto.ReqUrl,
		ReqMethod:            dto.ReqMethod,
		ReqHttpHeader:        httpHeader,
		ReqBody:              dto.ReqBody,
		ReqTimeout:           reqTimeout,
		CheckInterval:        reqInterval,
		RetryCount:           dto.RetryCount,
		ExpectedStatus:       dto.ExpectedStatus,
		ExpectedResponseBody: *dto.ExpectedResponseBody,
		ResponseValidation:   *dto.ResponseValidation,
		AuthType:             dto.AuthType,
		AuthCredentials:      dto.AuthCredentials,
	}
	if dto.VerifySSL == "on" || dto.VerifySSL == "true" {
		hc.VerifySSL = true
	} else {
		hc.VerifySSL = false
	}
	return hc
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
