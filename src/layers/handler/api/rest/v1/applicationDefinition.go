package v1

import (
	data "kukus/nam/v2/layers/data"
	handlers "kukus/nam/v2/layers/handler"
	services "kukus/nam/v2/layers/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ApplicationController struct {
	Service services.ApplicationService
}

func NewApplicationController(db *data.Database) *ApplicationController {
	return &ApplicationController{
		Service: services.ApplicationService{
			Database: db,
		},
	}
}

// Initializes new Controller on declared RouterGroup, with specified resources
func (ac *ApplicationController) Init(routerGroup *gin.RouterGroup) {
	routerGroup.POST("/", ac.NewApplication)
	routerGroup.GET("/", ac.GetAll)
	routerGroup.PATCH("/", handlers.MethodNotAllowed)
	routerGroup.PUT("/", handlers.MethodNotAllowed)
	routerGroup.DELETE("/", handlers.MethodNotAllowed)
	idGroup := routerGroup.Group("/:appId")
	{
		idGroup.POST("/", handlers.MethodNotAllowed)
		idGroup.GET("/", ac.GetById)
		idGroup.PATCH("/", handlers.MethodNotImplemented)
		idGroup.PUT("/", handlers.MethodNotImplemented)
		idGroup.DELETE("/", ac.DeleteById)
		NewApplicationInstanceController(ac.Service.Database).Init(idGroup.Group("/instances"))
	}
}

// GetAll ApplicationDefinition
func (ac *ApplicationController) GetAll(ctx *gin.Context) {
	dtos, err := ac.Service.GetAllApplications()
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to read application list", "trace": err})
		return
	}
	ctx.JSON(200, dtos)
}

func (ac *ApplicationController) DeleteById(ctx *gin.Context) {
	appId, err := strconv.Atoi(ctx.Param("appId"))
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Must include ID of application"})
		return
	}
	err = data.DeleteApplicationDefinitionById(ac.Service.Database.Pool, uint64(appId))
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to delete application", "trace": err})
		return
	}
	ctx.Status(204) // No Content
}

// GetAll ApplicationDefinition
func (ac *ApplicationController) GetById(ctx *gin.Context) {
	appId, err := strconv.Atoi(ctx.Param("appId"))
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Must include ID of application"})
	}
	dtos, err := ac.Service.GetApplicationById(uint(appId))
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
func (ac *ApplicationController) NewApplication(ctx *gin.Context) {
	var appDef data.ApplicationDefinitionDAO
	if err := ctx.ShouldBindJSON(&appDef); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid JSON", "trace": err})
		return
	}
	id, err := ac.Service.CreateApplication(appDef)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to create ApplicationDefinition", "trace": err})
		return
	}
	ctx.JSON(201, id)
}

// Update whole ApplicationDefinition
func (ac *ApplicationController) Patch(ctx *gin.Context) {
	ctx.AbortWithStatus(409)
}

// Update part of ApplicationDefinition
func (ac *ApplicationController) Put(ctx *gin.Context) {
	ctx.AbortWithStatus(409)
}

// Delete ApplicationDefinition
func (ac *ApplicationController) Delete(ctx *gin.Context) {
	ctx.AbortWithStatus(409)
}
