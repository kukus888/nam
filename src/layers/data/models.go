package data

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	Id       uint   `json:"server_id" db:"server_id"`
	Alias    string `json:"alias" db:"server_alias"`
	Hostname string `json:"hostname" db:"server_hostname"`
}

type Healthcheck struct {
	Id             *uint         `json:"id" db:"id"`
	Name           string        `json:"name" db:"name"`
	Description    string        `json:"description" db:"description"`
	ReqUrl         string        `json:"url" db:"url"`
	ReqMethod      string        `json:"method" db:"method"`   // GET, POST, etc.
	ReqHttpHeader  http.Header   `json:"headers" db:"headers"` // Custom headers
	ReqBody        string        `json:"body" db:"body"`       // Request body for POST/PUT
	ReqTimeout     time.Duration `json:"timeout" db:"timeout"`
	CheckInterval  time.Duration `json:"check_interval" db:"check_interval"`
	RetryCount     int           `json:"retry_count" db:"retry_count"`       // Number of retries before marking as unhealthy
	RetryInterval  time.Duration `json:"retry_interval" db:"retry_interval"` // Time between retries
	ExpectedStatus int           `json:"expected_status" db:"expected_status"`
	Protocol       string        `json:"protocol" db:"protocol"` // http, https

	// Response validation
	ExpectedResponseBody string `json:"expected_response_body" db:"expected_response_body"` // Expected response content
	ResponseValidation   string `json:"response_validation" db:"response_validation"`       // none, contains, exact, regex

	// SSL/TLS
	VerifySSL bool `json:"verify_ssl" db:"verify_ssl"`

	// Authentication
	AuthType        string `json:"auth_type" db:"auth_type"`               // none, basic, bearer, custom
	AuthCredentials string `json:"auth_credentials" db:"auth_credentials"` // stored securely
}

type HealthcheckDTO struct {
	Id            *uint  `json:"id"`
	Name          string `json:"name" binding:"required"`
	Description   string `json:"description"`
	ReqUrl        string `json:"url" binding:"required"`
	ReqMethod     string `json:"method" binding:"required"` // GET, POST, etc.
	ReqHeader     string `json:"headers"`                   // Custom headers
	ReqBody       string `json:"body"`                      // Request body for POST/PUT
	ReqTimeout    int    `json:"timeout" binding:"required"`
	CheckInterval int    `json:"check_interval" binding:"required"`
	RetryCount    int    `json:"retry_count"`    // Number of retries before marking as unhealthy
	RetryInterval int    `json:"retry_interval"` // Time between retries
	Protocol      string `json:"protocol"`
	UseHttps      bool   `json:"use_https"`

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

func (dto HealthcheckDTO) ToHealthcheck() (*Healthcheck, error) {
	httpHeader := http.Header{}
	if dto.ReqHeader != "" {
		header, err := ParseHeadersFromString(dto.ReqHeader)
		if err == nil {
			httpHeader = *header
		}
	}
	reqTimeout, _ := time.ParseDuration(strconv.Itoa(dto.ReqTimeout) + "s")
	reqInterval, _ := time.ParseDuration(strconv.Itoa(dto.CheckInterval) + "s")
	hc := Healthcheck{
		Id:                   dto.Id,
		Name:                 dto.Name,
		Description:          dto.Description,
		ReqUrl:               dto.ReqUrl,
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
	if dto.UseHttps {
		hc.Protocol = "https"
	} else {
		hc.Protocol = "http"
	}
	return &hc, nil
}

func ParseHeadersFromString(headersStr string) (*http.Header, error) {
	httpHeader := http.Header{}
	defer func() (*http.Header, error) {
		if r := recover(); r != nil {
			return nil, errors.New("Invalid header format")
		}
		return nil, nil
	}()
	headerLines := strings.Split(headersStr, "\n")
	for _, line := range headerLines {
		key := strings.Split(line, ":")[0]
		value := strings.TrimSpace(strings.Split(line, ":")[1])
		httpHeader[key] = []string{value}
	}
	return &httpHeader, nil
}

type HealthcheckResult struct {
	Id                    uint64    `json:"id" db:"id"`
	HealthcheckID         uint      `json:"healthcheck_id" db:"healthcheck_id"`
	ApplicationInstanceID uint      `json:"application_instance_id" db:"application_instance_id"`
	IsSuccessful          bool      `json:"is_successful" db:"is_successful"`
	TimeStart             time.Time `json:"time_start" db:"time_start"`
	TimeEnd               time.Time `json:"time_end" db:"time_end"`
	ResStatus             int       `json:"res_status" db:"res_status"`
	ResBody               string    `json:"res_body" db:"res_body"`
	ResTime               int       `json:"res_time" db:"res_time"` // in milliseconds
	ErrorMessage          string    `json:"error_message" db:"error_message"`
}

// ApplicationDefinition represents the definition of an application and its general properties
type ApplicationDefinition struct {
	Id            uint   `json:"id" db:"application_definition_id"`
	Name          string `json:"name" db:"application_definition_name"`
	Port          int    `json:"port" db:"application_definition_port"`
	Type          string `json:"type" db:"application_definition_type"`
	HealthcheckId *uint  `json:"healthcheck_id" db:"healthcheck_id"` // ID of the healthcheck, if any
}

// ApplicationDefinitionDAO represents the definition of an application and its general properties
type ApplicationDefinitionDAO struct {
	Id            uint   `json:"id" db:"id"`
	Name          string `json:"name" db:"name"`
	Port          int    `json:"port" db:"port"`
	Type          string `json:"type" db:"type"`
	HealthcheckId *uint  `json:"healthcheck_id" db:"healthcheck_id"`
}

// ApplicationInstance represents an instance of an application
type ApplicationInstance struct {
	Id                      uint   `json:"id" db:"id"`
	Name                    string `json:"name" db:"name"`
	MaintenanceMode         bool   `json:"maintenance_mode" db:"maintenance_mode"`
	TopologyNodeID          uint   `json:"topology_node_id" db:"topology_node_id"`
	ApplicationDefinitionID uint   `json:"application_definition_id"`
	ServerID                uint   `json:"server_id" db:"server_id"`
}

// ApplicationInstance represents an instance of an application
// Joined with ApplicationDefinition and Server
// This struct is used to return full information about the application instance
type ApplicationInstanceFull struct {
	Id              uint   `json:"id" db:"application_instance_id"`
	Name            string `json:"name" db:"application_instance_name"`
	MaintenanceMode bool   `json:"maintenance_mode" db:"maintenance_mode"`
	TopologyNodeID  uint   `json:"topology_node_id" db:"topology_node_id"`
	ApplicationDefinition
	Server
}

// User represents a user in the system, as stored in the database
type User struct {
	Id           uint64 `json:"id" db:"id"`
	Username     string `json:"username" db:"username"`
	Email        string `json:"email" db:"email"`
	Color        string `json:"color" db:"color"` // Name of the color for the user, tailwind.css
	PasswordHash string `json:"password_hash" db:"password_hash"`
	RoleId       uint64 `json:"role_id" db:"role_id"`
}

// UserDTO is a DTO for creating users
type UserDTO struct {
	Id       uint64 `json:"id" db:"id"`
	Username string `json:"username" db:"username"`
	Email    string `json:"email" db:"email"`
	Color    string `json:"color" db:"color"`
	Password string `json:"password" db:"password"`
	RoleId   uint64 `json:"role_id" db:"role_id"`
}

type UserChangePasswordDTO struct {
	Id       uint64 `json:"id" db:"id"`
	Password string `json:"password" db:"password"`
}

func (u UserDTO) ToUser() (*User, error) {
	return &User{
		Username: u.Username,
		Email:    u.Email,
	}, nil
}

// UserDTO is a Data Transfer Object for User, used for login
type UserLoginDTO struct {
	Username *string `json:"username"`
	Password *string `json:"password"`
}

// UserFull represents a user with all their details (joins role)
type UserFull struct {
	Id              uint64 `json:"id" db:"user_id"`
	Username        string `json:"username" db:"username"`
	Email           string `json:"email" db:"email"`
	PasswordHash    string `json:"password_hash" db:"password_hash"`
	RoleId          uint64 `json:"role_id" db:"role_id"`
	RoleName        string `json:"role_name" db:"role_name"`
	RoleColor       string `json:"role_color" db:"role_color"` // Name of the color for the role, tailwind.css
	RoleDescription string `json:"role_description" db:"role_description"`
}

type Role struct {
	Id          uint64 `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Color       string `json:"color" db:"color"` // Name of the color for the role, tailwind.css
	Description string `json:"description" db:"description"`
}
