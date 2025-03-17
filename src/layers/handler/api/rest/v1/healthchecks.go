package v1

import (
	"fmt"
	data "kukus/nam/v2/layers/data"
	handlers "kukus/nam/v2/layers/handler"
	services "kukus/nam/v2/layers/service"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type HealthcheckController struct {
	Service services.HealthcheckService
}

func NewHealthcheckController(db *data.Database) *HealthcheckController {
	return &HealthcheckController{
		Service: services.HealthcheckService{
			Database: db,
		},
	}
}

// Initializes new Controller on declared RouterGroup, with specified resources
func (ac *HealthcheckController) Init(routerGroup *gin.RouterGroup) {
	routerGroup.POST("/", ac.NewHealthcheck)
	routerGroup.GET("/", ac.GetAll)
	routerGroup.PATCH("/", handlers.MethodNotAllowed)
	routerGroup.PUT("/", handlers.MethodNotAllowed)
	routerGroup.DELETE("/", handlers.MethodNotAllowed)
	idGroup := routerGroup.Group("/:hcId")
	{
		idGroup.POST("/", handlers.MethodNotAllowed)
		idGroup.GET("/", ac.GetById)
		idGroup.PATCH("/", handlers.MethodNotImplemented)
		idGroup.PUT("/", ac.UpdateHealthcheck)
		idGroup.DELETE("/", handlers.MethodNotImplemented)
		NewApplicationInstanceController(ac.Service.Database).Init(idGroup.Group("/instances"))
	}
}

// GetAll Healthcheck
func (ac *HealthcheckController) GetAll(ctx *gin.Context) {
	dtos, err := data.GetHealthChecksAll(ac.Service.Database.Pool)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to read application list", "trace": err})
		return
	}
	ctx.JSON(200, dtos)
}

// GetAll Healthcheck
func (ac *HealthcheckController) GetById(ctx *gin.Context) {
	hcId, err := strconv.Atoi(ctx.Param("hcId"))
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Must include ID of application"})
	}
	dtos, err := data.GetHealthCheckById(ac.Service.Database.Pool, uint(hcId))
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to read application list", "trace": err})
		return
	} else if dtos == nil {
		ctx.AbortWithStatus(404)
	} else {
		ctx.JSON(200, dtos)
	}
}

// Create Healthcheck
func (ac *HealthcheckController) NewHealthcheck(ctx *gin.Context) {
	var hc data.Healthcheck
	if err := ctx.ShouldBindJSON(&hc); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid JSON", "trace": err})
		return
	}
	id, err := hc.DbInsert(ac.Service.Database.Pool)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to create Healthcheck", "trace": err})
		return
	}
	ctx.JSON(201, id)
}

// Update whole Healthcheck
func (ac *HealthcheckController) UpdateHealthcheck(ctx *gin.Context) {
	var dto HealthcheckDTO
	err := ctx.ShouldBindBodyWithJSON(&dto)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid JSON", "trace": err.Error()})
		return
	}
	dao, err := dto.ToDAO()
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Error converting DTO to DAO", "trace": err.Error()})
		return
	}
	err = dao.Update(ac.Service.Database.Pool)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to update Healthcheck", "trace": err.Error()})
		return
	}
	ctx.JSON(200, dto)
}

// Delete Healthcheck
func (ac *HealthcheckController) Delete(ctx *gin.Context) {
	ctx.AbortWithStatus(409)
}

type HealthcheckDTO struct {
	ID             uint   `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	Url            string `json:"url"`
	Method         string `json:"method"`  // GET, POST, etc.
	Headers        string `json:"headers"` // Custom headers, per HTTP standard. One per line, key: value
	Body           string `json:"body"`    // Request body for POST/PUT
	Timeout        int    `json:"timeout"`
	CheckInterval  int    `json:"check_interval"` // Time between checks
	RetryCount     int    `json:"retry_count"`    // Number of retries before marking as unhealthy
	RetryInterval  int    `json:"retry_interval"` // Time between retries
	ExpectedStatus int    `json:"expected_status"`

	// Response validation
	ExpectedResponseBody string `json:"expected_response_body"` // Expected response content
	ResponseValidation   string `json:"response_validation"`    // contains, exact, regex

	// SSL/TLS
	VerifySSL      bool `json:"verify_ssl"`
	SSLExpiryAlert bool `json:"ssl_expiry_alert"`

	// Authentication
	AuthType        string `json:"auth_type"`        // none, basic, bearer, custom
	AuthCredentials string `json:"auth_credentials"` // stored securely
}

func (dto *HealthcheckDTO) ToDAO() (*data.Healthcheck, error) {
	timeout, err := time.ParseDuration(strconv.Itoa(dto.Timeout) + "s")
	if err != nil {
		return nil, err
	}
	retryInterval, err := time.ParseDuration(strconv.Itoa(dto.RetryInterval) + "s")
	if err != nil {
		return nil, err
	}
	checkInterval, err := time.ParseDuration(strconv.Itoa(dto.CheckInterval) + "s")
	if err != nil {
		return nil, err
	}
	headers := make([]http.Header, 0)
	if dto.Headers != "" {
		for _, header := range strings.Split(dto.Headers, "\n") {
			parts := strings.Split(header, ":")
			if len(parts) != 2 {
				return nil, fmt.Errorf("Invalid header format: %s", header)
			}
			headers = append(headers, http.Header{parts[0]: {parts[1]}})
		}
	}
	return &data.Healthcheck{
		ID:                   dto.ID,
		Name:                 dto.Name,
		Description:          dto.Description,
		Url:                  dto.Url,
		Method:               dto.Method,
		Headers:              headers,
		Body:                 dto.Body,
		Timeout:              timeout,
		CheckInterval:        checkInterval,
		RetryCount:           dto.RetryCount,
		RetryInterval:        retryInterval,
		ExpectedStatus:       dto.ExpectedStatus,
		ExpectedResponseBody: dto.ExpectedResponseBody,
		ResponseValidation:   dto.ResponseValidation,
		VerifySSL:            dto.VerifySSL,
		SSLExpiryAlert:       dto.SSLExpiryAlert,
		AuthType:             dto.AuthType,
		AuthCredentials:      dto.AuthCredentials,
	}, nil
}
