package v1

import (
	data "kukus/nam/v2/layers/data"
	handlers "kukus/nam/v2/layers/handler"
	services "kukus/nam/v2/layers/service"
	"strconv"

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
		idGroup.PUT("/", handlers.MethodNotImplemented)
		idGroup.DELETE("/", handlers.MethodNotImplemented)
		NewApplicationInstanceController(ac.Service.Database).Init(idGroup.Group("/instances"))
	}
}

// GetAll ApplicationDefinition
func (ac *HealthcheckController) GetAll(ctx *gin.Context) {
	dtos, err := data.GetHealthChecksAll(ac.Service.Database.Pool)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to read application list", "trace": err})
		return
	}
	ctx.JSON(200, dtos)
}

// GetAll ApplicationDefinition
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

// Create ApplicationDefinition
func (ac *HealthcheckController) NewHealthcheck(ctx *gin.Context) {
	var hc data.Healthcheck
	if err := ctx.ShouldBindJSON(&hc); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid JSON", "trace": err})
		return
	}
	id, err := hc.DbInsert(ac.Service.Database.Pool)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to create ApplicationDefinition", "trace": err})
		return
	}
	ctx.JSON(201, id)
}

// Update whole ApplicationDefinition
func (ac *HealthcheckController) Patch(ctx *gin.Context) {
	ctx.AbortWithStatus(409)
}

// Update part of ApplicationDefinition
func (ac *HealthcheckController) Put(ctx *gin.Context) {
	ctx.AbortWithStatus(409)
}

// Delete ApplicationDefinition
func (ac *HealthcheckController) Delete(ctx *gin.Context) {
	ctx.AbortWithStatus(409)
}
